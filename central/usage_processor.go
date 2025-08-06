// File: central/usage_processor.go (修正版)
package central

import (
	"database/sql"
	"fmt"
	"karashi/db"
	"karashi/mappers"
	"karashi/model"
)

func ProcessUsageRecords(tx *sql.Tx, conn *sql.DB, records []model.UnifiedInputRecord) ([]model.TransactionRecord, error) {
	if len(records) == 0 {
		return []model.TransactionRecord{}, nil
	}

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
		ar := model.TransactionRecord{
			TransactionDate: rec.Date,
			Flag:            3,
			JanCode:         rec.JanCode,
			YjCode:          rec.YjCode,
			ProductName:     rec.ProductName,
			YjQuantity:      rec.YjQuantity,
			YjUnitName:      rec.YjUnitName,
		}

		// ▼▼▼ キー生成ロジックを修正 ▼▼▼
		key := ar.JanCode
		isSyntheticKey := false
		if key == "" || key == "0000000000000" {
			key = fmt.Sprintf("9999999999999%s", ar.ProductName)
			isSyntheticKey = true
		}
		// ▲▲▲

		if master, ok := mastersMap[key]; ok {
			if master.Origin == "JCSHMS" {
				ar.ProcessFlagMA = FlagComplete
				ar.ProcessingStatus = sql.NullString{String: "completed", Valid: true}
			} else {
				ar.ProcessFlagMA = FlagProvisional
				ar.ProcessingStatus = sql.NullString{String: "provisional", Valid: true}
			}

			// ▼▼▼ 修正点: masterのProductCodeをar.JanCodeに設定する ▼▼▼
			ar.JanCode = master.ProductCode

			if master.JanPackInnerQty > 0 {
				ar.JanQuantity = ar.YjQuantity / master.JanPackInnerQty
			} else {
				ar.JanQuantity = ar.YjQuantity
			}

			mappers.MapProductMasterToTransaction(&ar, master)
			// YJコードが空の既存マスターを参照した場合、ここで上書きする
			if ar.YjCode == "" && master.YjCode != "" {
				ar.YjCode = master.YjCode
			}

		} else {
			if jcshms, ok := jcshmsMap[ar.JanCode]; ok && ar.JanCode != "" && jcshms.JC018 != "" {
				ar.ProcessFlagMA = FlagComplete
				ar.ProcessingStatus = sql.NullString{String: "completed", Valid: true}
				yjCode := jcshms.JC009
				if yjCode == "" {
					newYj, _ := db.NextSequenceInTx(tx, "MA2Y", "MA2Y", 8)
					yjCode = newYj
				}
				ar.YjCode = yjCode
				mappers.CreateMasterFromJcshmsInTx(tx, ar.JanCode, yjCode, jcshms)
				mappers.MapJcshmsToTransaction(&ar, jcshms)

				if jcshms.JA006.Float64 > 0 {
					ar.JanQuantity = ar.YjQuantity / jcshms.JA006.Float64
				} else {
					ar.JanQuantity = ar.YjQuantity
				}

			} else {
				ar.ProcessFlagMA = FlagProvisional
				ar.ProcessingStatus = sql.NullString{String: "provisional", Valid: true}

				janForMaster := ar.JanCode
				if isSyntheticKey {
					janForMaster = ""
				}
				newYj, productCode, err := createProvisionalMaster(tx, key, janForMaster, ar.ProductName, mastersMap)
				if err != nil {
					return nil, err
				}
				ar.YjCode = newYj
				ar.JanCode = productCode
				ar.JanQuantity = ar.YjQuantity
			}
		}
		finalRecords = append(finalRecords, ar)
	}

	return finalRecords, nil
}
