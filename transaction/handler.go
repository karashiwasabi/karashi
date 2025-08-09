// File: transaction/handler.go
package transaction

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"karashi/central"
	"karashi/db"
	"net/http"
	"strings"
)

// GetReceiptsHandlerは、指定された日付の伝票番号一覧を返します。
func GetReceiptsHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		if date == "" {
			http.Error(w, "Date parameter is required", http.StatusBadRequest)
			return
		}
		receipts, err := db.GetReceiptNumbersByDate(conn, date)
		if err != nil {
			http.Error(w, "Failed to get receipt numbers", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(receipts)
	}
}

// GetTransactionHandlerは、指定された伝票番号の全明細を返します。
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

// DeleteTransactionHandlerは、指定された伝票番号の全明細を削除します。
func DeleteTransactionHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(map[string]string{"message": "Deleted successfully"})
	}
}

// ReProcessTransactionsHandlerは、仮登録状態のレコードを最新マスターで更新します。
func ReProcessTransactionsHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tx, err := conn.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// 中央処理ロジックを呼び出し
		count, err := central.ReProcessProvisionalRecords(tx, conn)
		if err != nil {
			http.Error(w, "Failed to reprocess provisional records: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("%d 件の仮登録データを更新しました。", count),
		})
	}
}
