// File: db/aggregation.go (修正後の完全なコード)
package db

import (
	"database/sql"
	"fmt"
	"karashi/model"
	"strings"
)

// GetAggregatedTransactions はフィルターに基づいて集計結果を返します
func GetAggregatedTransactions(conn *sql.DB, filters model.AggregationFilters) ([]model.YJGroup, error) {
	// --- Step 1: マスター関連のフィルターで製品リストを作成 ---
	masterQuery := `SELECT ` + selectColumns + ` FROM product_master p WHERE 1=1 `
	masterArgs := []interface{}{}

	if filters.KanaName != "" {
		masterQuery += " AND p.kana_name LIKE ? "
		masterArgs = append(masterArgs, "%"+filters.KanaName+"%")
	}
	if len(filters.DrugTypes) > 0 && filters.DrugTypes[0] != "" {
		var conditions []string
		flagMap := map[string]string{
			"poison": "p.flag_poison = 1", "deleterious": "p.flag_deleterious = 1", "narcotic": "p.flag_narcotic = 1",
			"psychotropic1": "p.flag_psychotropic = 1", "psychotropic2": "p.flag_psychotropic = 2", "psychotropic3": "p.flag_psychotropic = 3",
			"stimulant": "p.flag_stimulant = 1", "stimulant_raw": "p.flag_stimulant_raw = 1",
		}
		for _, dt := range filters.DrugTypes {
			if condition, ok := flagMap[dt]; ok {
				conditions = append(conditions, condition)
			}
		}
		if len(conditions) > 0 {
			masterQuery += " AND (" + strings.Join(conditions, " OR ") + ")"
		}
	}

	masterRows, err := conn.Query(masterQuery, masterArgs...)
	if err != nil {
		return nil, fmt.Errorf("aggregation master query failed: %w", err)
	}
	defer masterRows.Close()

	var allMasters []*model.ProductMaster
	productCodes := []string{}
	productCodeSet := make(map[string]struct{})

	for masterRows.Next() {
		m, err := scanProductMaster(masterRows)
		if err != nil {
			return nil, err
		}
		allMasters = append(allMasters, m)
		if _, ok := productCodeSet[m.ProductCode]; !ok {
			productCodeSet[m.ProductCode] = struct{}{}
			productCodes = append(productCodes, m.ProductCode)
		}
	}
	if len(allMasters) == 0 {
		return []model.YJGroup{}, nil
	}

	// --- Step 2: 期間フィルターでトランザクションを取得し、製品コードごとにまとめる ---
	transactionsByProductCode := make(map[string][]*model.TransactionRecord)
	if len(productCodes) > 0 {
		transactionQuery := `SELECT ` + TransactionColumns + ` FROM transaction_records
			WHERE jan_code IN (?` + strings.Repeat(",?", len(productCodes)-1) + `)`

		txArgs := []interface{}{}
		for _, pc := range productCodes {
			txArgs = append(txArgs, pc)
		}

		if filters.StartDate != "" {
			transactionQuery += " AND transaction_date >= ?"
			txArgs = append(txArgs, filters.StartDate)
		}
		if filters.EndDate != "" {
			transactionQuery += " AND transaction_date <= ?"
			txArgs = append(txArgs, filters.EndDate)
		}

		transactionQuery += " ORDER BY transaction_date"

		txRows, err := conn.Query(transactionQuery, txArgs...)
		if err != nil {
			return nil, fmt.Errorf("aggregation transaction query failed: %w", err)
		}
		defer txRows.Close()

		for txRows.Next() {
			t, err := ScanTransactionRecord(txRows)
			if err != nil {
				return nil, err
			}
			transactionsByProductCode[t.JanCode] = append(transactionsByProductCode[t.JanCode], t)
		}
	}

	// --- Step 3: YJコード -> 包装表記 の階層でグループを正しく統合 ---
	yjGroupMap := make(map[string]*model.YJGroup)

	for _, m := range allMasters {
		if m.YjCode == "" {
			continue
		}

		if _, ok := yjGroupMap[m.YjCode]; !ok {
			yjGroupMap[m.YjCode] = &model.YJGroup{
				YjCode:        m.YjCode,
				ProductName:   m.ProductName,
				PackageGroups: []model.PackageGroup{},
			}
		}

		packageKey := fmt.Sprintf("%s %.2f%s", m.PackageSpec, m.JanPackInnerQty, m.YjUnitName)
		var targetPkg *model.PackageGroup

		for i := range yjGroupMap[m.YjCode].PackageGroups {
			if yjGroupMap[m.YjCode].PackageGroups[i].PackageKey == packageKey {
				targetPkg = &yjGroupMap[m.YjCode].PackageGroups[i]
				break
			}
		}

		if targetPkg == nil {
			newPkg := model.PackageGroup{
				PackageKey:   packageKey,
				Transactions: []model.TransactionRecord{},
			}
			yjGroupMap[m.YjCode].PackageGroups = append(yjGroupMap[m.YjCode].PackageGroups, newPkg)
			targetPkg = &yjGroupMap[m.YjCode].PackageGroups[len(yjGroupMap[m.YjCode].PackageGroups)-1]
		}

		if transactions, ok := transactionsByProductCode[m.ProductCode]; ok {
			for _, t := range transactions {
				targetPkg.Transactions = append(targetPkg.Transactions, *t)
			}
		}
	}

	// --- Step 4: 集計計算 ---
	var result []model.YJGroup
	for _, yjGroup := range yjGroupMap {
		// まず包装(PackageGroup)ごとの小計を計算
		for i := range yjGroup.PackageGroups {
			pkgGroup := &yjGroup.PackageGroups[i]
			for _, t := range pkgGroup.Transactions {
				var signedJanQty, signedYjQty float64

				if t.Flag == 2 || t.Flag == 3 || t.Flag == 12 {
					signedJanQty = -t.JanQuantity
					signedYjQty = -t.YjQuantity
				} else if t.Flag == 1 || t.Flag == 11 || t.Flag == 4 {
					signedJanQty = t.JanQuantity
					signedYjQty = t.YjQuantity
				} else if t.Flag == 5 {
					signedJanQty = -t.JanQuantity
					signedYjQty = -t.YjQuantity
				}

				pkgGroup.TotalJanQty += signedJanQty
				pkgGroup.TotalYjQty += signedYjQty

				if t.Flag == 3 {
					if -signedYjQty > pkgGroup.MaxUsageYjQty {
						pkgGroup.MaxUsageYjQty = -signedYjQty
					}
					if -t.JanQuantity > pkgGroup.MaxUsageJanQty {
						pkgGroup.MaxUsageJanQty = -t.JanQuantity
					}
				}
			}
		}

		// ▼▼▼ 修正点: YJGroup(大グループ)の合計値を計算する処理を追加 ▼▼▼
		// 包装ごとの小計を、大グループの合計に足し上げていく
		for _, pg := range yjGroup.PackageGroups {
			yjGroup.TotalJanQty += pg.TotalJanQty
			yjGroup.TotalYjQty += pg.TotalYjQty
			if pg.MaxUsageYjQty > yjGroup.MaxUsageYjQty {
				yjGroup.MaxUsageYjQty = pg.MaxUsageYjQty
			}
		}
		// ▲▲▲ ここまで ▲▲▲

		result = append(result, *yjGroup)
	}

	// --- Step 5: 「動きのない品目」フィルターを適用 ---
	var filteredResult []model.YJGroup
	for _, yjGroup := range result {
		hasPrescription := false
		totalTransactions := 0
		for _, pg := range yjGroup.PackageGroups {
			totalTransactions += len(pg.Transactions)
			for _, t := range pg.Transactions {
				if t.Flag == 3 {
					hasPrescription = true
					break
				}
			}
			if hasPrescription {
				break
			}
		}

		if filters.NoMovement {
			if !hasPrescription {
				filteredResult = append(filteredResult, yjGroup)
			}
		} else {
			if totalTransactions > 0 {
				filteredResult = append(filteredResult, yjGroup)
			}
		}
	}

	return filteredResult, nil
}
