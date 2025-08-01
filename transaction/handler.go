// File: transaction/handler.go (New File)
package transaction

import (
	"database/sql"
	"encoding/json"
	"karashi/db"
	"net/http"
	"strings"
)

// GetReceiptsHandler handles fetching receipt numbers by date.
func GetReceiptsHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		if date == "" {
			http.Error(w, "Date parameter is required", http.StatusBadRequest)
			return
		}
		numbers, err := db.GetReceiptNumbersByDate(conn, date)
		if err != nil {
			http.Error(w, "Failed to get receipt numbers", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(numbers)
	}
}

// GetTransactionHandler handles fetching all details for a specific receipt number.
func GetTransactionHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		receiptNumber := strings.TrimPrefix(r.URL.Path, "/api/transaction/")
		if receiptNumber == "" {
			http.Error(w, "Receipt number is required", http.StatusBadRequest)
			return
		}
		records, err := db.GetTransactionsByReceiptNumber(conn, receiptNumber)
		if err != nil {
			http.Error(w, "Failed to get transaction details", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	}
}

// DeleteTransactionHandler handles deleting an entire transaction slip.
func DeleteTransactionHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		receiptNumber := strings.TrimPrefix(r.URL.Path, "/api/transaction/delete/")
		if receiptNumber == "" {
			http.Error(w, "Receipt number is required", http.StatusBadRequest)
			return
		}

		tx, err := conn.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		if err := db.DeleteTransactionsByReceiptNumberInTx(tx, receiptNumber); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Slip deleted successfully"})
	}
}
