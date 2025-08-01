// File: inout/handler.go (Corrected)
package inout

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"karashi/db"
	"karashi/model"
	"log"
	"net/http"
	"time"
)

type SavePayload struct {
	IsNewClient           bool                      `json:"isNewClient"`
	ClientCode            string                    `json:"clientCode"`
	ClientName            string                    `json:"clientName"`
	TransactionDate       string                    `json:"transactionDate"`
	TransactionType       string                    `json:"transactionType"`
	Records               []model.TransactionRecord `json:"records"`
	OriginalReceiptNumber string                    `json:"originalReceiptNumber"` // ★ ADDED
}

func SaveInOutHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload SavePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		tx, err := conn.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// --- Client Handling ---
		clientCode := payload.ClientCode
		if payload.IsNewClient {
			// ... (no changes here)
		}

		// --- Receipt Number Handling ---
		var receiptNumber string
		dateStr := payload.TransactionDate
		if dateStr == "" {
			dateStr = time.Now().Format("20060102")
		}

		// ★★★ ADDED: Check if this is an update or a new slip ★★★
		if payload.OriginalReceiptNumber != "" {
			// This is an UPDATE (delete and re-register)
			receiptNumber = payload.OriginalReceiptNumber
			if err := db.DeleteTransactionsByReceiptNumberInTx(tx, receiptNumber); err != nil {
				// We ignore the "no transaction found" error, as it's not critical
				if err.Error() != fmt.Sprintf("no transaction found with receipt number: %s", receiptNumber) {
					http.Error(w, "Failed to delete original slip for update", http.StatusInternalServerError)
					return
				}
			}
		} else {
			// This is a NEW slip, generate a new number
			var lastSeq int
			q := `SELECT CAST(SUBSTR(receipt_number, 11) AS INTEGER) FROM transaction_records 
				  WHERE receipt_number LIKE ? ORDER BY 1 DESC LIMIT 1`
			err = tx.QueryRow(q, "io"+dateStr+"%").Scan(&lastSeq)
			if err != nil && err != sql.ErrNoRows {
				http.Error(w, "Failed to get last receipt number sequence", http.StatusInternalServerError)
				return
			}
			newSeq := lastSeq + 1
			receiptNumber = fmt.Sprintf("io%s%03d", dateStr, newSeq)
		}

		// --- Record Saving ---
		flagMap := map[string]int{"入庫": 1, "出庫": 2}
		flag := flagMap[payload.TransactionType]

		for i := range payload.Records {
			rec := &payload.Records[i]
			rec.ClientCode = clientCode
			rec.ReceiptNumber = receiptNumber
			rec.TransactionDate = dateStr
			rec.Flag = flag
			rec.ProcessFlagMA = "COMPLETE"
			rec.ProcessingStatus = sql.NullString{String: "completed", Valid: true}

			if rec.JanCode != "" {
				master, err := db.GetProductMasterByCodeInTx(tx, rec.JanCode)
				if err != nil {
					http.Error(w, "Failed to check product master", http.StatusInternalServerError)
					return
				}
				if master == nil {
					// ... (create new master logic - no changes here)
				}
			}
		}

		if err := db.PersistTransactionRecordsInTx(tx, payload.Records); err != nil {
			log.Printf("Failed to persist records: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Database save failed."})
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message":       "Saved successfully",
			"receiptNumber": receiptNumber,
		})
	}
}
