// File: main.go
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"karashi/loader"
	"karashi/tani"
	"karashi/usage"
)

func main() {
	port := flag.String("port", "8080", "HTTP port")
	dbPath := flag.String("db", "yamato.db", "SQLite file path")
	from := flag.String("from", "00000101", "filter start date (YYYYMMDD)")
	to := flag.String("to", "99991231", "filter end date (YYYYMMDD)")
	flag.Parse()

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("DB open failed: %v", err)
	}
	defer db.Close()

	// スキーマ＆マスター初期ロード
	if err := loader.InitDatabase(db); err != nil {
		log.Fatalf("master init failed: %v", err)
	}

	// 単位マスター読み込み
	if _, err := tani.LoadTANIFile("SOU/TANI.CSV"); err != nil {
		log.Fatalf("tani load failed: %v", err)
	}

	// HTTP ハンドラ登録
	mux := http.NewServeMux()
	mux.Handle("/uploadUsage", usage.UploadUsageHandler(db, *from, *to))
	mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	srv := &http.Server{
		Addr:         ":" + *port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("→ starting on :%s", *port)
		openBrowser("http://localhost:" + *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("⏳ shutting down…")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}
	log.Println("✅ server stopped")
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
