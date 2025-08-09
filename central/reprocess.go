// File: central/reprocess.go (Corrected)
package central

import (
	"database/sql"
	"fmt"
	"karashi/db"
	"karashi/mappers"
)

// ReProcessProvisionalRecords finds all provisional transactions, re-matches them
// against the latest master data, and updates them if a match is found.
func ReProcessProvisionalRecords(tx *sql.Tx, conn *sql.DB) (int, error) {
	// 1. Get all provisional records
	provisionalRecords, err := db.GetProvisionalTransactions(conn)
	if err != nil {
		return 0, fmt.Errorf("failed to get provisional transactions: %w", err)
	}

	if len(provisionalRecords) == 0 {
		return 0, nil
	}

	// 2. Collect all unique JAN codes from the provisional records
	janSet := make(map[string]struct{})
	for _, rec := range provisionalRecords {
		if rec.JanCode != "" {
			janSet[rec.JanCode] = struct{}{}
		}
	}
	janList := make([]string, 0, len(janSet))
	for jan := range janSet {
		janList = append(janList, jan)
	}

	if len(janList) == 0 {
		return 0, nil
	}

	// 3. Get all necessary master data in bulk for efficiency
	mastersMap, err := db.GetProductMastersByCodesMap(conn, janList)
	if err != nil {
		return 0, fmt.Errorf("failed to bulk get product masters for reprocessing: %w", err)
	}
	jcshmsMap, err := db.GetJcshmsByCodesMap(conn, janList)
	if err != nil {
		return 0, fmt.Errorf("failed to bulk get jcshms for reprocessing: %w", err)
	}

	updatedCount := 0
	// 4. Loop through each provisional record and try to update it
	for _, rec := range provisionalRecords {
		// Attempt to match with an existing, complete master record first
		if master, ok := mastersMap[rec.JanCode]; ok {
			mappers.MapProductMasterToTransaction(&rec, master)

			// ▼▼▼ ここから修正 ▼▼▼
			// JCSHMS由来のマスターに紐付いた場合のみ、ステータスを完了にする
			if master.Origin == "JCSHMS" {
				rec.ProcessFlagMA = FlagComplete
				rec.ProcessingStatus = sql.NullString{String: "completed", Valid: true}
			} else {
				// JCSHMS由来でないマスター(手入力など)に紐付いた場合は、
				// MAフラグも処理ステータスも PROVISIONAL のままにする
				rec.ProcessFlagMA = FlagProvisional
				// rec.ProcessingStatus は変更しない
			}
			// ▲▲▲ ここまで修正 ▲▲▲

		} else if jcshms, ok := jcshmsMap[rec.JanCode]; ok && rec.JanCode != "" && jcshms.JC018 != "" {
			// If no product_master exists, try to match with JCSHMS data
			// and create a new product_master record from it.
			yjCode := jcshms.JC009
			if yjCode == "" {
				newYj, _ := db.NextSequenceInTx(tx, "MA2Y", "MA2Y", 8)
				yjCode = newYj
			}
			rec.YjCode = yjCode
			mappers.CreateMasterFromJcshmsInTx(tx, rec.JanCode, yjCode, jcshms)
			mappers.MapJcshmsToTransaction(&rec, jcshms)
			rec.ProcessingStatus = sql.NullString{String: "completed", Valid: true}
			rec.ProcessFlagMA = FlagComplete
		} else {
			// If still no match is found, skip this record
			continue
		}

		// 5. Save the updated record to the database
		if err := db.UpdateFullTransactionInTx(tx, &rec); err != nil {
			return 0, fmt.Errorf("failed to update reprocessed transaction ID %d: %w", rec.ID, err)
		}
		updatedCount++
	}

	return updatedCount, nil
}
