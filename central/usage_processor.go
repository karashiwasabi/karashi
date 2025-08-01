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
		if rec.JanCode != "" {
			janSet[rec.JanCode] = struct{}{}
		}
		key := rec.JanCode
		if key == "" {
			key = fmt.Sprintf("9999999999999%s", rec.ProductName)
		}
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

		key := ar.JanCode
		isSyntheticKey := false
		if key == "" {
			key = fmt.Sprintf("9999999999999%s", ar.ProductName)
			isSyntheticKey = true
		}

		if master, ok := mastersMap[key]; ok {
			if master.Origin == "JCSHMS" {
				ar.ProcessFlagMA = FlagComplete
				ar.ProcessingStatus = sql.NullString{String: "completed", Valid: true}
			} else {
				ar.ProcessFlagMA = FlagProvisional
				ar.ProcessingStatus = sql.NullString{String: "provisional", Valid: true}
			}
			ar.JanQuantity = ar.YjQuantity // USAGEではYjQuantityをJanQuantityにコピー
			mappers.MapProductMasterToTransaction(&ar, master)
			ar.JanCode = master.ProductCode
		} else {
			if jcshms, ok := jcshmsMap[ar.JanCode]; ok && jcshms.JC018 != "" {
				ar.ProcessFlagMA = FlagComplete
				ar.ProcessingStatus = sql.NullString{String: "completed", Valid: true}
				yjCode := jcshms.JC009
				if yjCode == "" {
					// ★★★ 呼び出し方を修正 ★★★
					newYj, _ := db.NextSequenceInTx(tx, "MA2Y", "MA2Y", 8)
					yjCode = newYj
				}
				ar.YjCode = yjCode
				mappers.CreateMasterFromJcshmsInTx(tx, ar.JanCode, yjCode, jcshms)
				mappers.MapJcshmsToTransaction(&ar, jcshms)
				ar.JanQuantity = ar.YjQuantity
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
