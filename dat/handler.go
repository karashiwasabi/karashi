// File: dat/handler.go
package dat

import (
	"database/sql"
	"encoding/json"
	"karashi/central"
	"karashi/db"
	"karashi/model"
	"log"
	"net/http"
)

func UploadDatHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "File upload error: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.MultipartForm.RemoveAll()

		var allParsedRecords []model.UnifiedInputRecord
		for _, fileHeader := range r.MultipartForm.File["file"] {
			file, err := fileHeader.Open()
			if err != nil {
				log.Printf("Failed to open file %s: %v", fileHeader.Filename, err)
				continue
			}
			defer file.Close()
			parsed, err := ParseDat(file)
			if err != nil {
				log.Printf("Failed to parse file %s: %v", fileHeader.Filename, err)
				continue
			}
			allParsedRecords = append(allParsedRecords, parsed...)
		}

		// ✨ ここからが新しい高速化処理 ✨
		// トランザクションを開始
		tx, err := conn.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// 新しい一括処理関数を呼び出す
		finalRecords, err := central.ProcessDatRecords(tx, conn, allParsedRecords)
		if err != nil {
			log.Printf("central.ProcessDatRecords failed: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if len(finalRecords) > 0 {
			if err := db.PersistTransactionRecordsInTx(tx, finalRecords); err != nil {
				log.Printf("db.PersistTransactionRecordsInTx error: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			log.Printf("transaction commit error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		// ✨ ここまで ✨

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Parsed and processed DAT files successfully",
			"records": finalRecords,
		})
	}
}
