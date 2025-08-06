package inventory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"karashi/central"
	"karashi/db"
	"karashi/model"
	"net/http"
)

// UploadInventoryHandlerは棚卸ファイルを受け取り、centralで処理した後、transaction_recordsテーブルに登録します
func UploadInventoryHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "File upload error", http.StatusBadRequest)
			return
		}
		defer file.Close()

		parsedData, err := ParseInventoryFile(file)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse file: %v", err), http.StatusBadRequest)
			return
		}
		date := parsedData.Date
		if date == "" {
			http.Error(w, "Inventory date not found in file's H record", http.StatusBadRequest)
			return
		}

		// ファイルから読み取ったデータをJANコード(または製品名)単位で集計
		aggregatedData := make(map[string]model.UnifiedInputRecord)
		for _, row := range parsedData.Records {
			key := row.JanCode
			if key == "" || key == "0000000000000" {
				key = fmt.Sprintf("9999999999999%s", row.ProductName)
			}
			rowYjQty := row.JanQuantity * row.JanPackInnerQty

			if data, ok := aggregatedData[key]; ok {
				data.YjQuantity += rowYjQty
				data.JanQuantity += row.JanQuantity
				aggregatedData[key] = data
			} else {
				row.YjQuantity = rowYjQty
				aggregatedData[key] = row
			}
		}
		var aggregatedRecords []model.UnifiedInputRecord
		for _, rec := range aggregatedData {
			aggregatedRecords = append(aggregatedRecords, rec)
		}

		tx, err := conn.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// centralの関数を呼び出し、マスター準備とトランザクション作成を委任
		finalRecords, err := central.ProcessInventoryRecords(tx, conn, aggregatedRecords)
		if err != nil {
			http.Error(w, "Failed to process inventory records", http.StatusInternalServerError)
			return
		}

		// 登録日付と伝票番号を最終レコードに付与
		receiptNumber := fmt.Sprintf("INV%s", date)
		for i := range finalRecords {
			finalRecords[i].TransactionDate = date
			finalRecords[i].ReceiptNumber = receiptNumber
			finalRecords[i].LineNumber = fmt.Sprintf("%d", i+1)

			// ▼▼▼ ここから修正 ▼▼▼
			// DatQuantityを設定していた不要なループを完全に削除
			// ▼▼▼ ここまで修正 ▼▼▼
		}

		// データベースに保存
		if len(finalRecords) > 0 {
			if err := db.PersistTransactionRecordsInTx(tx, finalRecords); err != nil {
				http.Error(w, "Failed to save inventory records to transaction", http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		// 結果を返す
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": fmt.Sprintf("%d件の棚卸データを登録しました。", len(finalRecords)),
			"details": finalRecords,
		})
	}
}
