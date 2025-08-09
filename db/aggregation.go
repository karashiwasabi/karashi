// File: db/aggregation.go (Corrected and Finalized)
package db

import (
	"database/sql"
	"fmt"
	"karashi/model"
	"karashi/units"
	"sort"
	"strings"
)

// GetStockLedger はフィルターに基づいて在庫台帳と発注点情報を生成します
func GetStockLedger(conn *sql.DB, filters model.AggregationFilters) ([]model.StockLedgerYJGroup, error) {
	// --- Step 1: マスター関連のフィルターで対象となる製品リストを作成 ---
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
		return nil, fmt.Errorf("ledger master query failed: %w", err)
	}
	defer masterRows.Close()

	mastersByYjCode := make(map[string][]*model.ProductMaster)
	var productCodes []string
	productCodeSet := make(map[string]struct{})
	for masterRows.Next() {
		m, err := scanProductMaster(masterRows)
		if err != nil {
			return nil, err
		}
		if m.YjCode != "" {
			mastersByYjCode[m.YjCode] = append(mastersByYjCode[m.YjCode], m)
		}
		if _, ok := productCodeSet[m.ProductCode]; !ok {
			productCodeSet[m.ProductCode] = struct{}{}
			productCodes = append(productCodes, m.ProductCode)
		}
	}
	if len(productCodes) == 0 {
		return []model.StockLedgerYJGroup{}, nil
	}

	// --- Step 2: 期間フィルターで関連する全トランザクションを正しい順序で取得 ---
	transactionQuery := `SELECT ` + TransactionColumns + ` FROM transaction_records
		WHERE jan_code IN (?` + strings.Repeat(",?", len(productCodes)-1) + `)`
	txArgs := []interface{}{}
	for _, pc := range productCodes {
		txArgs = append(txArgs, pc)
	}
	if filters.StartDate != "" {
		transactionQuery += " AND transaction_date >= ? "
		txArgs = append(txArgs, filters.StartDate)
	}
	if filters.EndDate != "" {
		transactionQuery += " AND transaction_date <= ? "
		txArgs = append(txArgs, filters.EndDate)
	}
	transactionQuery += " ORDER BY transaction_date, flag, id"

	txRows, err := conn.Query(transactionQuery, txArgs...)
	if err != nil {
		return nil, fmt.Errorf("ledger transaction query failed: %w", err)
	}
	defer txRows.Close()

	transactionsByProductCode := make(map[string][]*model.TransactionRecord)
	for txRows.Next() {
		t, err := ScanTransactionRecord(txRows)
		if err != nil {
			return nil, err
		}
		transactionsByProductCode[t.JanCode] = append(transactionsByProductCode[t.JanCode], t)
	}

	// --- Step 3: グループ化と各種計算 ---
	var result []model.StockLedgerYJGroup
	for yjCode, masters := range mastersByYjCode {
		yjGroup := model.StockLedgerYJGroup{
			YjCode:         yjCode,
			ProductName:    masters[0].ProductName,
			YjUnitName:     units.ResolveName(masters[0].YjUnitName), // ★ 追加
			PackageLedgers: []model.StockLedgerPackageGroup{},
		}
		var yjTotalStart, yjTotalEnd, yjTotalReorderPoint float64
		var isYjReorderNeeded bool

		txsByPackage := make(map[string][]*model.TransactionRecord)

		for _, master := range masters {
			packageKey := fmt.Sprintf("%s|%f|%s", master.PackageSpec, master.JanPackInnerQty, master.YjUnitName)
			if txs, ok := transactionsByProductCode[master.ProductCode]; ok {
				txsByPackage[packageKey] = append(txsByPackage[packageKey], txs...)
			}
		}

		for _, txs := range txsByPackage {
			sort.Slice(txs, func(i, j int) bool {
				t1 := txs[i]
				t2 := txs[j]
				if t1.TransactionDate != t2.TransactionDate {
					return t1.TransactionDate < t2.TransactionDate
				}
				if t1.Flag != t2.Flag {
					return t1.Flag < t2.Flag
				}
				return t1.ID < t2.ID
			})
		}

		for key, txs := range txsByPackage {
			var janUnitName string
			if len(txs) > 0 {
				janUnitName = txs[0].JanUnitName
			}

			pkgLedger := model.StockLedgerPackageGroup{
				PackageKey:   key,
				JanUnitName:  janUnitName, // ★ 追加
				Transactions: []model.LedgerTransaction{},
			}
			var pkgStartingBalance, pkgEndingBalance, pkgNetChange, currentBalance float64

			if len(txs) > 0 {
				firstInventoryDate := ""
				firstInventoryIndex := -1
				for i, t := range txs {
					if t.Flag == 0 {
						firstInventoryDate = t.TransactionDate
						firstInventoryIndex = i
						break
					}
				}

				if firstInventoryIndex != -1 {
					inventorySum := 0.0
					for _, t := range txs {
						if t.TransactionDate == firstInventoryDate && t.Flag == 0 {
							inventorySum += t.YjQuantity
						}
					}
					pkgStartingBalance = inventorySum
					currentBalance = 0

					isAfterInventory := false
					for _, t := range txs {
						if !isAfterInventory && t.TransactionDate < firstInventoryDate {
							pkgLedger.Transactions = append(pkgLedger.Transactions, model.LedgerTransaction{TransactionRecord: *t, RunningBalance: 0})
							continue
						}

						if t.TransactionDate == firstInventoryDate && t.Flag == 0 {
							isAfterInventory = true
						}

						if isAfterInventory {
							if t.TransactionDate == firstInventoryDate && t.Flag == 0 {
								currentBalance += t.YjQuantity
							} else {
								var signedYjQty float64
								switch t.Flag {
								case 2, 3, 5, 12:
									signedYjQty = -t.YjQuantity
								case 1, 4, 11:
									signedYjQty = t.YjQuantity
								}
								currentBalance += signedYjQty
							}
						}
						pkgLedger.Transactions = append(pkgLedger.Transactions, model.LedgerTransaction{TransactionRecord: *t, RunningBalance: currentBalance})
					}
					pkgEndingBalance = currentBalance
				} else {
					firstTx := txs[0]
					var firstTxChange float64
					switch firstTx.Flag {
					case 2, 3, 5, 12:
						firstTxChange = -firstTx.YjQuantity
					case 1, 4, 11:
						firstTxChange = firstTx.YjQuantity
					}
					currentBalance = 0 - firstTxChange
					pkgStartingBalance = currentBalance

					for _, t := range txs {
						var signedYjQty float64
						switch t.Flag {
						case 2, 3, 5, 12:
							signedYjQty = -t.YjQuantity
						case 1, 4, 11:
							signedYjQty = t.YjQuantity
						}
						currentBalance += signedYjQty
						pkgLedger.Transactions = append(pkgLedger.Transactions, model.LedgerTransaction{TransactionRecord: *t, RunningBalance: currentBalance})
					}
					pkgEndingBalance = currentBalance
				}
				pkgNetChange = pkgEndingBalance - pkgStartingBalance
			}

			pkgLedger.StartingBalance, pkgLedger.NetChange, pkgLedger.EndingBalance = pkgStartingBalance, pkgNetChange, pkgEndingBalance

			var maxUsage float64
			for _, t := range txs {
				if t.Flag == 3 {
					if t.YjQuantity > maxUsage {
						maxUsage = t.YjQuantity
					}
				}
			}
			reorderPoint := maxUsage * filters.Coefficient
			pkgLedger.MaxUsage, pkgLedger.ReorderPoint = maxUsage, reorderPoint
			pkgLedger.IsReorderNeeded = pkgLedger.EndingBalance < reorderPoint && maxUsage > 0

			yjGroup.PackageLedgers = append(yjGroup.PackageLedgers, pkgLedger)
			yjTotalStart += pkgLedger.StartingBalance
			yjTotalEnd += pkgLedger.EndingBalance
			yjTotalReorderPoint += pkgLedger.ReorderPoint
			if pkgLedger.IsReorderNeeded {
				isYjReorderNeeded = true
			}
		}

		if len(yjGroup.PackageLedgers) > 0 {
			yjGroup.StartingBalance, yjGroup.EndingBalance, yjGroup.NetChange = yjTotalStart, yjTotalEnd, yjTotalEnd-yjTotalStart
			yjGroup.TotalReorderPoint = yjTotalReorderPoint
			yjGroup.IsReorderNeeded = isYjReorderNeeded
			result = append(result, yjGroup)
		}
	}
	return result, nil
}
