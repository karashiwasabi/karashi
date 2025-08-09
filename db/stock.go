// File: db/stock.go (Corrected)
package db

import (
	"database/sql"
	"fmt"
	"karashi/model"
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

		switch flag {
		case 1, 4, 11: // 在庫を増やす
			stockMap[yjCode] += quantity
		case 2, 3, 5, 12: // 在庫を減らす
			stockMap[yjCode] -= quantity
		}
	}
	return stockMap, nil
}

// CalculateCurrentStockForProduct は、棚卸を考慮した正確な現在庫を計算します
func CalculateCurrentStockForProduct(conn *sql.DB, janCode string) (float64, error) {
	var lastInventory model.TransactionRecord
	var hasInventory bool

	// 1. 最新の棚卸レコードを取得
	row := conn.QueryRow(`
		SELECT `+TransactionColumns+` FROM transaction_records
		WHERE jan_code = ? AND flag = 0
		ORDER BY transaction_date DESC, id DESC
		LIMIT 1`, janCode)

	// ScanTransactionRecordヘルパー関数を利用してスキャン
	rec, err := ScanTransactionRecord(row)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to get latest inventory for %s: %w", janCode, err)
	}
	if err == nil {
		hasInventory = true
		lastInventory = *rec
	}

	// 2. 棚卸以降の変動を計算
	var query string
	var args []interface{}

	if hasInventory {
		query = `
			SELECT
				SUM(CASE
					WHEN flag IN (1, 4, 11) THEN yj_quantity
					WHEN flag IN (2, 3, 5, 12) THEN -yj_quantity
					ELSE 0
				END)
			FROM transaction_records
			WHERE jan_code = ? AND (transaction_date > ? OR (transaction_date = ? AND id > ?))`
		args = []interface{}{janCode, lastInventory.TransactionDate, lastInventory.TransactionDate, lastInventory.ID}
	} else {
		query = `
			SELECT
				SUM(CASE
					WHEN flag IN (1, 4, 11) THEN yj_quantity
					WHEN flag IN (2, 3, 5, 12) THEN -yj_quantity
					ELSE 0
				END)
			FROM transaction_records
			WHERE jan_code = ?`
		args = []interface{}{janCode}
	}

	// ▼▼▼ ここから修正 ▼▼▼
	// NULLの可能性がある値を安全に受け取るためにsql.NullFloat64を使用
	var nullNetChange sql.NullFloat64
	err = conn.QueryRow(query, args...).Scan(&nullNetChange)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to calculate net change for %s: %w", janCode, err)
	}
	netChange := nullNetChange.Float64 // NULLの場合は0.0になる
	// ▲▲▲ ここまで修正 ▲▲▲

	// 3. 最終在庫を計算
	if hasInventory {
		return lastInventory.YjQuantity + netChange, nil
	}
	return netChange, nil
}
