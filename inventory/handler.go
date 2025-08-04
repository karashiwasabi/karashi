// File: inventory/handler.go
package inventory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"karashi/db"
	"karashi/model"
	"karashi/units"
	"log"
	"net/http"
)

const (
	FlagInvAdjIn  = 4
	FlagInvAdjOut = 5
)

// UploadInventoryHandlerは棚卸ファイルを受け取り、YJ単位で在庫調整を行います
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

		tx, err := conn.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// --- 1. ファイルから読み取った物理在庫をYJコード単位で集計 ---
		physicalYjStockMap := make(map[string]float64)
		uniqueYjCodes := make(map[string]struct{})
		fileDataByJan := make(map[string]FileRow)

		for _, row := range parsedData.Rows {
			master, _ := db.GetProductMasterByCode(conn, row.JanCode)

			var yjQty float64
			if master != nil { // 既存マスターがある場合
				yjQty = row.PhysicalJanQty * master.YjPackUnitQty
			} else { // 新規の場合
				yjQty = row.PhysicalJanQty * row.InnerPackQty

				// 新規マスターを作成
				newMasterInput := model.ProductMasterInput{
					ProductCode:     row.JanCode,
					YjCode:          row.YjCode,
					ProductName:     row.ProductName,
					Origin:          "PROVISIONAL_INV", // 棚卸由来の暫定マスター
					YjUnitName:      units.ResolveCode(row.YjUnitName),
					YjPackUnitQty:   row.InnerPackQty, // YJ包装数量 = 内包装数量と仮定
					JanPackInnerQty: row.InnerPackQty,
				}
				if err := db.CreateProductMasterInTx(tx, newMasterInput); err != nil {
					log.Printf("Failed to create provisional master from inventory for JAN %s: %v", row.JanCode, err)
					http.Error(w, "Failed to create new master", http.StatusInternalServerError)
					return
				}
			}
			physicalYjStockMap[row.YjCode] += yjQty
			uniqueYjCodes[row.YjCode] = struct{}{}
			fileDataByJan[row.JanCode] = row
		}

		var yjCodeList []string
		for yj := range uniqueYjCodes {
			yjCodeList = append(yjCodeList, yj)
		}

		// --- 2. DBから現在の理論在庫をYJコード単位で取得 ---
		systemYjStockMap, err := db.CalculateYjStockByDate(conn, date, yjCodeList)
		if err != nil {
			http.Error(w, "Failed to calculate system stock", http.StatusInternalServerError)
			return
		}

		// --- 3. 差分を計算し、調整トランザクションを生成 ---
		var adjustments []model.TransactionRecord
		receiptNumber := fmt.Sprintf("INV%s", date)
		lineNumber := 1

		for yjCode, physicalCount := range physicalYjStockMap {
			systemCount := systemYjStockMap[yjCode]
			variance := physicalCount - systemCount

			if variance == 0 {
				continue
			}

			// 代表のJANコードを検索して基本情報を設定
			var representativeJan string
			for jan, data := range fileDataByJan {
				if data.YjCode == yjCode {
					representativeJan = jan
					break
				}
			}

			adj := model.TransactionRecord{
				TransactionDate:  date,
				ReceiptNumber:    receiptNumber,
				LineNumber:       fmt.Sprintf("%d", lineNumber),
				YjCode:           yjCode,
				JanCode:          representativeJan,
				UnitPrice:        0,
				Subtotal:         0,
				ProcessingStatus: sql.NullString{String: "provisional", Valid: true},
			}

			if variance > 0 {
				adj.Flag = FlagInvAdjIn
				adj.YjQuantity = variance
			} else {
				adj.Flag = FlagInvAdjOut
				adj.YjQuantity = -variance
			}
			adjustments = append(adjustments, adj)
			lineNumber++
		}

		// --- 4. データベースに保存 ---
		if len(adjustments) > 0 {
			if err := db.PersistTransactionRecordsInTx(tx, adjustments); err != nil {
				http.Error(w, "Failed to save adjustments", http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": fmt.Sprintf("%d件のYJコードについて在庫調整を登録しました。", len(adjustments)),
			"details": adjustments,
		})
	}
}
