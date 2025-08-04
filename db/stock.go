// File: db/stock.go
package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// CalculateYjStockByDateは、指定日時点のYJコードごとの理論在庫を計算します。(YJ単位)
func CalculateYjStockByDate(conn *sql.DB, date string, yjCodes []string) (map[string]float64, error) {
	if len(yjCodes) == 0 {
		return make(map[string]float64), nil
	}

	stockMap := make(map[string]float64)
	for _, code := range yjCodes {
		stockMap[code] = 0
	}

	q := `SELECT yj_code, flag, yj_quantity FROM transaction_records
          WHERE transaction_date <= ? AND yj_code IN (?` + strings.Repeat(",?", len(yjCodes)-1) + `)`

	args := []interface{}{date}
	for _, code := range yjCodes {
		args = append(args, code)
	}

	rows, err := conn.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions for YJ stock calculation: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var yjCode string
		var flag int
		var quantity float64
		if err := rows.Scan(&yjCode, &flag, &quantity); err != nil {
			return nil, err
		}

		// flag 1:入庫, 2:出庫, 3:処方, 4:棚卸増, 5:棚卸減
		switch flag {
		case 1, 4: // 在庫を増やす
			stockMap[yjCode] += quantity
		case 2, 3, 5: // 在庫を減らす
			stockMap[yjCode] -= quantity
		}
	}
	return stockMap, nil
}
