// File: main.go (Corrected)
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"karashi/dat"
	"karashi/db"
	"karashi/inout"
	"karashi/loader"
	"karashi/transaction" // ★ ADD THIS IMPORT
	"karashi/units"
	"karashi/usage"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	conn, err := sql.Open("sqlite3", "yamato.db")
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	defer conn.Close()

	if err := loader.InitDatabase(conn); err != nil {
		log.Fatalf("master init failed: %v", err)
	}
	if _, err := units.LoadTANIFile("SOU/TANI.CSV"); err != nil {
		log.Fatalf("tani init failed: %v", err)
	}
	log.Println("master init complete")

	mux := http.NewServeMux()

	// --- Register all API handlers ---

	// Client-related handlers
	mux.HandleFunc("/api/clients", func(w http.ResponseWriter, r *http.Request) {
		clients, err := db.GetAllClients(conn)
		if err != nil {
			http.Error(w, "Failed to get clients", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clients)
	})

	// Product search handler
	mux.HandleFunc("/api/products/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if len(query) < 2 {
			http.Error(w, "Query must be at least 2 characters", http.StatusBadRequest)
			return
		}
		results, err := db.SearchJcshmsByName(conn, query)
		if err != nil {
			http.Error(w, "Failed to search products", http.StatusInternalServerError)
			log.Printf("SearchJcshmsByName error: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	// Handlers from dedicated packages
	mux.Handle("/api/dat/upload", dat.UploadDatHandler(conn))
	mux.Handle("/api/usage/upload", usage.UploadUsageHandler(conn))
	mux.Handle("/api/inout/save", inout.SaveInOutHandler(conn))

	// ★★★ USE THE NEW, CLEANED-UP TRANSACTION HANDLERS ★★★
	mux.HandleFunc("/api/receipts", transaction.GetReceiptsHandler(conn))
	mux.HandleFunc("/api/transaction/", transaction.GetTransactionHandler(conn))
	mux.HandleFunc("/api/transaction/delete/", transaction.DeleteTransactionHandler(conn))

	// Static file and root handler
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	port := ":8080"
	go openBrowser("http://localhost" + port)
	log.Printf("starting on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Printf("failed to open browser: %v", err)
	}
}
