// File: monthend/handler.go (新規作成)
package monthend

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"karashi/db"
	"karashi/mappers"
	"karashi/model"
	"log"
	"net/http"
	"time"
)

// GetAvailableMonthsHandlerは、トランザクションが存在するユニークな年月のリストを返します。
func GetAvailableMonthsHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		months, err := db.GetUniqueTransactionMonths(conn)
		if err != nil {
			http.Error(w, "年月の取得に失敗しました", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(months)
	}
}

// CalculateMonthEndInventoryHandler は指定された年月の月末在庫を計算・保存します。
func CalculateMonthEndInventoryHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Month string `json:"month"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "リクエストボディが不正です", http.StatusBadRequest)
			return
		}
		targetYearMonth := payload.Month

		firstDayOfMonth, err := time.Parse("2006-01", targetYearMonth)
		if err != nil {
			http.Error(w, "month パラメータの形式が不正です", http.StatusBadRequest)
			return
		}
		monthEndDate := firstDayOfMonth.AddDate(0, 1, -1)
		monthEndDateStr := monthEndDate.Format("20060102")

		tx, err := conn.Begin()
		if err != nil {
			http.Error(w, "トランザクションの開始に失敗しました", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		if err := db.DeleteTransactionsByFlagAndDate(tx, 30, monthEndDateStr); err != nil {
			http.Error(w, "既存の月末在庫データの削除に失敗しました", http.StatusInternalServerError)
			return
		}

		monthEndRecords, err := processMonthEndInventory(conn, monthEndDateStr)
		if err != nil {
			log.Printf("月末在庫の計算処理中にエラー: %v", err)
			http.Error(w, "月末在庫の計算処理中にエラーが発生しました", http.StatusInternalServerError)
			return
		}

		if len(monthEndRecords) > 0 {
			if err := db.PersistTransactionRecordsInTx(tx, monthEndRecords); err != nil {
				log.Printf("月末在庫データの保存に失敗: %v", err)
				http.Error(w, "月末在庫データの保存に失敗しました", http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "トランザクションのコミットに失敗しました", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": fmt.Sprintf("%s の月末在庫 %d件を計算・保存しました。", targetYearMonth, len(monthEndRecords)),
			"records": monthEndRecords,
		})
	}
}

func processMonthEndInventory(conn *sql.DB, monthEndDateStr string) ([]model.TransactionRecord, error) {
	allProducts, err := db.GetAllProductMasters(conn)
	if err != nil {
		return nil, fmt.Errorf("全製品マスターの取得に失敗: %w", err)
	}

	var results []model.TransactionRecord
	receiptNumber := fmt.Sprintf("ZA%s", monthEndDateStr)

	for i, product := range allProducts {
		// ▼▼▼ 修正点: 参照先を transaction_records から最新の棚卸(flag=0)データを取得する関数に変更 ▼▼▼
		latestInv, err := db.GetLatestRecordByFlag(conn, product.ProductCode, monthEndDateStr, 0)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("WARN: 品目 %s の棚卸日取得に失敗: %v", product.ProductCode, err)
			continue
		}

		var startStock float64
		var startDate string
		if latestInv != nil {
			startStock = latestInv.YjQuantity
			startDate = latestInv.TransactionDate
		} else {
			startStock = 0
			startDate = "00000000"
		}

		netChange, err := db.GetTransactionSumForProduct(conn, product.ProductCode, startDate, monthEndDateStr)
		if err != nil {
			log.Printf("WARN: 品目 %s のトランザクション集計に失敗: %v", product.ProductCode, err)
			continue
		}

		monthEndStock := startStock + netChange

		tr := model.TransactionRecord{
			TransactionDate:  monthEndDateStr,
			ReceiptNumber:    receiptNumber,
			LineNumber:       fmt.Sprintf("%d", i+1),
			Flag:             30,
			JanCode:          product.ProductCode,
			YjQuantity:       monthEndStock,
			ProcessFlagMA:    "COMPLETE",
			ProcessingStatus: sql.NullString{String: "completed", Valid: true},
		}
		mappers.MapProductMasterToTransaction(&tr, product)
		results = append(results, tr)
	}

	return results, nil
}
