// File: inout/handler.go (修正後)
package inout

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

// SaveRecordInput はフロントエンドから送られる個々の明細データです。
// マスター作成に必要な情報も含むように拡張します。
type SaveRecordInput struct {
	// 基本的な入力情報
	JanQuantity float64 `json:"janQuantity"`
	ExpiryDate  string  `json:"expiryDate"`
	LotNumber   string  `json:"lotNumber"`

	// マスター情報 (製品マスターそのもの)
	// 'productCode' は ProductMaster 内の ProductCode を使います。
	model.ProductMaster
}

// SavePayload はフロントエンドから送られるJSON全体の構造体です
type SavePayload struct {
	IsNewClient           bool              `json:"isNewClient"`
	ClientCode            string            `json:"clientCode"`
	ClientName            string            `json:"clientName"`
	TransactionDate       string            `json:"transactionDate"`
	TransactionType       string            `json:"transactionType"`
	Records               []SaveRecordInput `json:"records"`
	OriginalReceiptNumber string            `json:"originalReceiptNumber"`
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

		// --- 得意先・伝票番号処理 (ここは変更なし) ---
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

		// --- ▼▼▼ レコード保存処理を全面的に修正 ▼▼▼ ---
		var finalRecords []model.TransactionRecord
		flagMap := map[string]int{"入庫": 11, "出庫": 12}
		flag := flagMap[payload.TransactionType]

		for i, rec := range payload.Records {
			// 製品コードが存在しない場合はスキップ
			if rec.ProductCode == "" {
				continue
			}

			// 1. マスター存在確認、なければ正規のデータで新規作成
			master, err := db.GetProductMasterByCodeInTx(tx, rec.ProductCode)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get product master for %s", rec.ProductCode), http.StatusInternalServerError)
				return
			}

			if master == nil {
				// マスターが存在しない場合、フロントから送られた完全な製品情報で新規作成する
				newMasterInput := model.ProductMasterInput{
					ProductCode:      rec.ProductCode,
					YjCode:           rec.YjCode,
					ProductName:      rec.ProductName,
					Origin:           "JCSHMS", // モーダル由来はJCSHMSとする
					KanaName:         rec.KanaName,
					MakerName:        rec.MakerName,
					PackageSpec:      rec.PackageSpec,
					YjUnitName:       rec.YjUnitName,
					YjPackUnitQty:    rec.YjPackUnitQty,
					FlagPoison:       rec.FlagPoison,
					FlagDeleterious:  rec.FlagDeleterious,
					FlagNarcotic:     rec.FlagNarcotic,
					FlagPsychotropic: rec.FlagPsychotropic,
					FlagStimulant:    rec.FlagStimulant,
					FlagStimulantRaw: rec.FlagStimulantRaw,
					JanPackInnerQty:  rec.JanPackInnerQty,
					JanUnitCode:      rec.JanUnitCode,
					JanPackUnitQty:   rec.JanPackUnitQty,
					NhiPrice:         rec.NhiPrice,
				}
				if err := db.CreateProductMasterInTx(tx, newMasterInput); err != nil {
					http.Error(w, fmt.Sprintf("Failed to create new product master for %s", rec.ProductCode), http.StatusInternalServerError)
					return
				}
				// 作成したマスターを再取得して利用する
				master, err = db.GetProductMasterByCodeInTx(tx, rec.ProductCode)
				if err != nil || master == nil {
					http.Error(w, "Failed to retrieve newly created product master", http.StatusInternalServerError)
					return
				}
			}

			// 2. ご指示の計算式でトランザクションレコードを作成
			yjQuantity := rec.JanQuantity * master.JanPackInnerQty
			subtotal := yjQuantity * master.NhiPrice

			// 完全な TransactionRecord を組み立てる
			tr := model.TransactionRecord{
				TransactionDate:  dateStr,
				ClientCode:       clientCode,
				ReceiptNumber:    receiptNumber,
				LineNumber:       fmt.Sprintf("%d", i+1),
				Flag:             flag,
				JanCode:          rec.ProductCode,
				JanQuantity:      rec.JanQuantity,
				YjQuantity:       yjQuantity,
				UnitPrice:        master.NhiPrice, // 単価にはマスターの薬価を設定
				Subtotal:         subtotal,
				ExpiryDate:       rec.ExpiryDate,
				LotNumber:        rec.LotNumber,
				ProcessFlagMA:    "COMPLETE",
				ProcessingStatus: sql.NullString{String: "completed", Valid: true},
			}

			// マッパーでマスターのその他情報をトランザクションに反映
			mappers.MapProductMasterToTransaction(&tr, master)

			finalRecords = append(finalRecords, tr)
		}

		if len(finalRecords) > 0 {
			if err := db.PersistTransactionRecordsInTx(tx, finalRecords); err != nil {
				log.Printf("Failed to persist records: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"message": "データベースへの保存に失敗しました。"})
				return
			}
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
