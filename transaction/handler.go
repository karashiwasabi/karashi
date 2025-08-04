// File: transaction/handler.go
package transaction

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"karashi/db"
	"karashi/mappers" // mappersをインポート
	"log"
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

// ▼▼▼ このハンドラを新規追加 ▼▼▼
// ReProcessTransactionsHandlerは仮トランザクションを再処理します
func ReProcessTransactionsHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		provisionalRecords, err := db.GetProvisionalTransactions(conn)
		if err != nil {
			log.Printf("Failed to get provisional transactions: %v", err)
			http.Error(w, "Failed to get provisional transactions", http.StatusInternalServerError)
			return
		}

		if len(provisionalRecords) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "更新対象のデータはありませんでした。"})
			return
		}

		tx, err := conn.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		updatedCount := 0
		for _, record := range provisionalRecords {
			if record.JanCode == "" {
				continue
			}
			master, err := db.GetProductMasterByCode(conn, record.JanCode)
			if err != nil || master == nil {
				continue // マスターが見つからない場合はスキップ
			}

			recCopy := record

			// マッパーを呼び出してトランザクション情報を更新
			mappers.MapProductMasterToTransaction(&recCopy, master)

			// マスターの由来に基づいてステータスを決定
			if master.Origin == "JCSHMS" {
				recCopy.ProcessFlagMA = "COMPLETE"
				recCopy.ProcessingStatus = sql.NullString{String: "completed", Valid: true}
			} else {
				recCopy.ProcessFlagMA = "PROVISIONAL"
				recCopy.ProcessingStatus = sql.NullString{String: "provisional", Valid: true} // provisionalのまま
			}

			// 更新されたトランザクション情報をDBに保存
			if err := db.UpdateFullTransactionInTx(tx, &recCopy); err != nil {
				log.Printf("Failed to reprocess transaction ID %d: %v", record.ID, err)
				continue
			}
			updatedCount++
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("%d件の取引データを更新しました。", updatedCount),
		})
	}
}
