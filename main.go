// File: main.go
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"karashi/loader"
	"karashi/tani"
	"karashi/unifiedrecords"

	_ "github.com/mattn/go-sqlite3"
)

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	default:
		return
	}
	exec.Command(cmd, args...).Start()
}

func main() {
	// 1) DB 接続
	db, err := sql.Open("sqlite3", "yamato.db")
	if err != nil {
		log.Fatal("DB接続失敗:", err)
	}
	defer db.Close()

	// 2) schema + CSV マスター初期化 (jcshms/jancode)
	if err := loader.InitDatabase(db); err != nil {
		log.Fatal("マスター初期化失敗:", err)
	}

	// 3) 単位マスター読み込み → 逆マップを構築
	unitMap, err := tani.LoadTANIFile("SOU/TANI.CSV")
	if err != nil {
		log.Fatal("単位マスター読み込み失敗:", err)
	}
	tani.SetMaps(unitMap)

	// 4) unifiedrecords パッケージへ DB 注入
	unifiedrecords.DB = db

	// 5) 静的ファイル配信 & DAT アップロードハンドラ
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/uploadDat", unifiedrecords.UploadDatHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	addr := "http://localhost:8080"
	log.Println("起動完了 →", addr)

	// 6) 少し待って自動でブラウザを開く
	go func() {
		time.Sleep(200 * time.Millisecond)
		openBrowser(addr)
	}()

	// 7) サーバ起動
	log.Fatal(http.ListenAndServe(":8080", nil))
}
