// File: usage/handler.go
package usage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"karashi/central"
	"karashi/db"
	"karashi/model"
	"log"
	"net/http"
)

func UploadUsageHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "File upload error: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.MultipartForm.RemoveAll()

		var allParsed []model.UnifiedInputRecord
		for _, fh := range r.MultipartForm.File["file"] {
			f, _ := fh.Open()
			defer f.Close()
			recs, _ := ParseUsage(f)
			allParsed = append(allParsed, recs...)
		}
		filtered := removeUsageDuplicates(allParsed)

		if len(filtered) == 0 {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(map[string]interface{}{"records": []model.TransactionRecord{}})
			return
		}

		tx, err := conn.Begin()
		if err != nil {
			log.Printf("Failed to begin transaction for usage: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// ✨ 修正点: tx を渡す
		finalRecords, err := central.ProcessUsageRecords(tx, conn, filtered)
		if err != nil {
			log.Printf("central.ProcessUsageRecords failed: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		minDate, maxDate := "99999999", "00000000"
		for _, rec := range filtered {
			if rec.Date < minDate {
				minDate = rec.Date
			}
			if rec.Date > maxDate {
				maxDate = rec.Date
			}
		}

		if err := db.DeleteUsageTransactionsInDateRange(tx, minDate, maxDate); err != nil {
			log.Printf("db.DeleteUsageTransactionsInDateRange error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if len(finalRecords) > 0 {
			if err := db.PersistTransactionRecordsInTx(tx, finalRecords); err != nil {
				log.Printf("PersistTransactionRecordsInTx error: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			log.Printf("transaction commit error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"records": finalRecords,
		})
	}
}

// removeUsageDuplicatesはUSAGE固有の重複排除ロジック
func removeUsageDuplicates(rs []model.UnifiedInputRecord) []model.UnifiedInputRecord {
	seen := make(map[string]struct{})
	var out []model.UnifiedInputRecord
	for _, r := range rs {
		key := fmt.Sprintf("%s|%s|%s|%s", r.Date, r.JanCode, r.YjCode, r.ProductName)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}
	return out
}
