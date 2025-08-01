// File: db/usage.go
package db

import (
	"database/sql"
	"fmt"
)

// DeleteUsageTransactionsInDateRangeは、指定された日付範囲内のUSAGEレコード(flag=3)を削除します。
func DeleteUsageTransactionsInDateRange(tx *sql.Tx, minDate, maxDate string) error {
	const q = `DELETE FROM transaction_records WHERE flag = 3 AND transaction_date BETWEEN ? AND ?`
	_, err := tx.Exec(q, minDate, maxDate)
	if err != nil {
		return fmt.Errorf("failed to delete usage transactions: %w", err)
	}
	return nil
}
