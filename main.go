// File: main.go
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"karashi/aggregation"
	"karashi/backup"
	"karashi/dat"
	"karashi/db"
	"karashi/inout"
	"karashi/inventory"
	"karashi/loader"
	"karashi/masteredit"
	"karashi/monthend" // 新しいパッケージを追加
	"karashi/transaction"
	"karashi/units"
	"karashi/updatemaster"
	"karashi/usage"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	_ "github.com/mattn/go-sqlite3"
)

func findInvalidCharacters(conn *sql.DB) {
	fmt.Println("--- Shift_JIS変換チェックを開始します ---")
	products, err := db.GetAllProductMasters(conn)
	if err != nil {
		log.Fatalf("製品マスターの取得に失敗: %v", err)
	}
	for _, p := range products {
		checkAndReport("製品マスター", p.ProductCode, "製品名", p.ProductName)
		checkAndReport("製品マスター", p.ProductCode, "カナ名", p.KanaName)
		checkAndReport("製品マスター", p.ProductCode, "メーカー名", p.MakerName)
		checkAndReport("製品マスター", p.ProductCode, "包装", p.PackageSpec)
	}
	clients, err := db.GetAllClients(conn)
	if err != nil {
		log.Fatalf("得意先マスターの取得に失敗: %v", err)
	}
	for _, c := range clients {
		checkAndReport("得意先マスター", c.Code, "得意先名", c.Name)
	}
	fmt.Println("--- チェックが完了しました ---")
}

func checkAndReport(table, code, field, value string) {
	encoder := japanese.ShiftJIS.NewEncoder()
	_, _, err := transform.String(encoder, value)
	if err != nil {
		fmt.Printf("【エラー】 テーブル「%s」のレコード「%s」の項目「%s」に変換できない文字が含まれています。\n  -> 値: %s\n", table, code, field, value)
	}
}

func main() {
	conn, err := sql.Open("sqlite3", "yamato.db")
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	defer conn.Close()

	if err := loader.InitDatabase(conn); err != nil {
		log.Fatalf("master init failed: %v", err)
	}
	// ▼▼▼ ここから追加 ▼▼▼
	// データベース内のYJコードの最大値からシーケンスを初期化
	if err := db.InitializeSequenceFromMaxYjCode(conn); err != nil {
		log.Fatalf("failed to initialize sequence from max yj_code: %v", err)
	}
	// ▲▲▲ ここまで追加 ▲▲▲
	if _, err := units.LoadTANIFile("SOU/TANI.CSV"); err != nil {
		log.Fatalf("tani init failed: %v", err)
	}
	log.Println("master init complete")

	findInvalidCharacters(conn)

	mux := http.NewServeMux()

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

	// Master edit handler
	mux.HandleFunc("/api/masters/editable", masteredit.GetEditableMastersHandler(conn))
	mux.HandleFunc("/api/master/update", masteredit.UpdateMasterHandler(conn))
	mux.HandleFunc("/api/master/update-jcshms", updatemaster.JCSHMSUpdateHandler(conn))

	// Handlers from dedicated packages
	mux.Handle("/api/dat/upload", dat.UploadDatHandler(conn))
	mux.Handle("/api/usage/upload", usage.UploadUsageHandler(conn))
	mux.Handle("/api/inout/save", inout.SaveInOutHandler(conn))
	mux.Handle("/api/inventory/upload", inventory.UploadInventoryHandler(conn))
	mux.HandleFunc("/api/receipts", transaction.GetReceiptsHandler(conn))
	mux.HandleFunc("/api/transaction/", transaction.GetTransactionHandler(conn))
	mux.HandleFunc("/api/transaction/delete/", transaction.DeleteTransactionHandler(conn))
	mux.HandleFunc("/api/clients/export", backup.ExportClientsHandler(conn))
	mux.HandleFunc("/api/clients/import", backup.ImportClientsHandler(conn))
	mux.HandleFunc("/api/products/export", backup.ExportProductsHandler(conn))
	mux.HandleFunc("/api/products/import", backup.ImportProductsHandler(conn))
	mux.HandleFunc("/api/aggregation", aggregation.GetAggregationHandler(conn))
	mux.HandleFunc("/api/transactions/reprocess", transaction.ReProcessTransactionsHandler(conn))
	mux.HandleFunc("/api/inventory/months", monthend.GetAvailableMonthsHandler(conn))
	mux.HandleFunc("/api/inventory/calculate", monthend.CalculateMonthEndInventoryHandler(conn))

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
