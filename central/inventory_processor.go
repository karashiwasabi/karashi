// File: central/inventory_processor.go
package central

import (
	"database/sql"
	"fmt"
	"karashi/db"
	"karashi/mappers"
	"karashi/model"
)

// ProcessInventoryRecordsは棚卸レコードを受け取り、マスターの照合・作成とトランザクションの完全化を行います。
func ProcessInventoryRecords(tx *sql.Tx, conn *sql.DB, records []model.UnifiedInputRecord) ([]model.TransactionRecord, error) {
	if len(records) == 0 {
		return []model.TransactionRecord{}, nil
	}

	// --- 1. キーの準備 ---
	keySet := make(map[string]struct{})
	janSet := make(map[string]struct{})
	for _, rec := range records {
		if rec.JanCode != "" && rec.JanCode != "0000000000000" {
			janSet[rec.JanCode] = struct{}{}
		}
		// ▼▼▼ キー生成ロジックを修正 ▼▼▼
		key := rec.JanCode
		if key == "" || key == "0000000000000" {
			key = fmt.Sprintf("9999999999999%s", rec.ProductName)
		}
		// ▲▲▲
		if key != "" {
			keySet[key] = struct{}{}
		}
	}
	keyList := make([]string, 0, len(keySet))
	for key := range keySet {
		keyList = append(keyList, key)
	}
	janList := make([]string, 0, len(janSet))
	for jan := range janSet {
		janList = append(janList, jan)
	}

	// --- 2. 必要なマスター情報を一括取得 ---
	mastersMap, err := db.GetProductMastersByCodesMap(conn, keyList)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk get product masters: %w", err)
	}
	jcshmsMap, err := db.GetJcshmsByCodesMap(conn, janList)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk get jcshms: %w", err)
	}

	var finalRecords []model.TransactionRecord
	for _, rec := range records {
		tr := model.TransactionRecord{
			Flag:        0,
			JanCode:     rec.JanCode,
			ProductName: rec.ProductName,
			YjQuantity:  rec.YjQuantity,
		}

		// ▼▼▼ キー生成ロジックを修正 ▼▼▼
		key := tr.JanCode
		isSyntheticKey := false
		if key == "" || key == "0000000000000" {
			key = fmt.Sprintf("9999999999999%s", tr.ProductName)
			isSyntheticKey = true
		}
		// ▲▲▲

		// --- 3. マスター照合と作成 ---
		if master, ok := mastersMap[key]; ok {
			if master.Origin == "JCSHMS" {
				tr.ProcessFlagMA = FlagComplete
				tr.ProcessingStatus = sql.NullString{String: "completed", Valid: true}
			} else {
				tr.ProcessFlagMA = FlagProvisional
				tr.ProcessingStatus = sql.NullString{String: "provisional", Valid: true}
			}
			if master.JanPackInnerQty > 0 {
				tr.JanQuantity = tr.YjQuantity / master.JanPackInnerQty
			}
			mappers.MapProductMasterToTransaction(&tr, master)

		} else {
			if jcshms, ok := jcshmsMap[tr.JanCode]; ok && tr.JanCode != "" && jcshms.JC018 != "" {
				tr.ProcessFlagMA = FlagComplete
				tr.ProcessingStatus = sql.NullString{String: "completed", Valid: true}
				yjCode := jcshms.JC009
				if yjCode == "" {
					newYj, _ := db.NextSequenceInTx(tx, "MA2Y", "MA2Y", 8)
					yjCode = newYj
				}
				tr.YjCode = yjCode
				mappers.CreateMasterFromJcshmsInTx(tx, tr.JanCode, yjCode, jcshms)
				mappers.MapJcshmsToTransaction(&tr, jcshms)
				if jcshms.JA006.Float64 > 0 {
					tr.JanQuantity = tr.YjQuantity / jcshms.JA006.Float64
				}

			} else {
				tr.ProcessFlagMA = FlagProvisional
				tr.ProcessingStatus = sql.NullString{String: "provisional", Valid: true}
				janForMaster := tr.JanCode
				if isSyntheticKey {
					janForMaster = ""
				}
				newYj, productCode, err := createProvisionalMaster(tx, key, janForMaster, tr.ProductName, mastersMap)
				if err != nil {
					return nil, err
				}
				tr.YjCode = newYj
				tr.JanCode = productCode
				tr.JanQuantity = tr.YjQuantity
			}
		}
		finalRecords = append(finalRecords, tr)
	}

	return finalRecords, nil
}
