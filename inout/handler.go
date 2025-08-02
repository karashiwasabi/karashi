// File: inout/handler.go (修正版)
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

// SavePayload はフロントエンドから送られてくるJSONの構造体です
type SavePayload struct {
	IsNewClient           bool                      `json:"isNewClient"`
	ClientCode            string                    `json:"clientCode"`
	ClientName            string                    `json:"clientName"`
	TransactionDate       string                    `json:"transactionDate"`
	TransactionType       string                    `json:"transactionType"`
	Records               []model.TransactionRecord `json:"records"`
	OriginalReceiptNumber string                    `json:"originalReceiptNumber"`
}

// SaveInOutHandler は入出庫データを保存します
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

		// --- 得意先処理 ---
		clientCode := payload.ClientCode
		if payload.IsNewClient {
			var exists int
			err := tx.QueryRow("SELECT 1 FROM client_master WHERE client_name = ? LIMIT 1", payload.ClientName).Scan(&exists)
			if err != sql.ErrNoRows {
				if err == nil {
					w.WriteHeader(http.StatusConflict)
					json.NewEncoder(w).Encode(map[string]string{"message": fmt.Sprintf("得意先名 '%s' は既に使用されています。", payload.ClientName)})
				} else {
					http.Error(w, "Failed to check client existence", http.StatusInternalServerError)
				}
				return
			}

			newCode, err := db.NextSequenceInTx(tx, "CL", "CL", 4)
			if err != nil {
				http.Error(w, "Failed to generate new client code", http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec("INSERT INTO client_master (client_code, client_name) VALUES (?, ?)", newCode, payload.ClientName)
			if err != nil {
				http.Error(w, "Failed to create new client", http.StatusInternalServerError)
				return
			}
			clientCode = newCode
		}

		// --- 伝票番号処理 ---
		var receiptNumber string
		dateStr := payload.TransactionDate
		if dateStr == "" {
			dateStr = time.Now().Format("20060102")
		}

		if payload.OriginalReceiptNumber != "" {
			receiptNumber = payload.OriginalReceiptNumber
			if err := db.DeleteTransactionsByReceiptNumberInTx(tx, receiptNumber); err != nil {
				if err.Error() != fmt.Sprintf("no transaction found with receipt number: %s", receiptNumber) {
					http.Error(w, "Failed to delete original slip for update", http.StatusInternalServerError)
					return
				}
			}
		} else {
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

		// --- レコード保存 ---
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
				if master == nil { // マスターが存在しない場合は作成
					newMaster := model.ProductMasterInput{
						ProductCode:      rec.JanCode,
						YjCode:           rec.YjCode,
						ProductName:      rec.ProductName,
						Origin:           "JCSHMS", // ★★★ ここを "MANUAL_ENTRY" から "JCSHMS" に修正 ★★★
						KanaName:         rec.KanaName,
						MakerName:        rec.MakerName,
						PackageSpec:      rec.PackageForm,
						YjUnitName:       rec.YjUnitName,
						YjPackUnitQty:    rec.YjPackUnitQty,
						JanPackInnerQty:  rec.JanPackInnerQty,
						JanPackUnitQty:   rec.JanPackUnitQty,
						FlagPoison:       rec.FlagPoison,
						FlagDeleterious:  rec.FlagDeleterious,
						FlagNarcotic:     rec.FlagNarcotic,
						FlagPsychotropic: rec.FlagPsychotropic,
						FlagStimulant:    rec.FlagStimulant,
						FlagStimulantRaw: rec.FlagStimulantRaw,
					}
					if err := db.CreateProductMasterInTx(tx, newMaster); err != nil {
						http.Error(w, "Failed to create new product master", http.StatusInternalServerError)
						return
					}
				}
			}
		}

		if err := db.PersistTransactionRecordsInTx(tx, payload.Records); err != nil {
			log.Printf("Failed to persist records: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "データベースへの保存に失敗しました。"})
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
