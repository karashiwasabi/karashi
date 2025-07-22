// File: db/inventory.go
package db

import (
	"database/sql"
	"fmt"
	"karashi/model"
)

// PersistInventoryRecordsは、棚卸レコードのリストをDBに保存します。
func PersistInventoryRecords(conn *sql.DB, records []model.ARInput) error {
	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 同じ日付の古い棚卸データを一旦削除
	if len(records) > 0 {
		date := records[0].Adate
		if _, err := tx.Exec("DELETE FROM inventory WHERE inv_date = ?", date); err != nil {
			return fmt.Errorf("failed to delete old inventory data: %w", err)
		}
	}

	const q = `
INSERT OR IGNORE INTO inventory (
  inv_date, inv_jan_code, inv_yj_code, inv_product_name, inv_quantity
) VALUES (?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(q)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, rec := range records {
		_, err := stmt.Exec(
			rec.Adate,
			rec.Ajc,
			rec.Ayj,
			rec.Apname,
			rec.Ajanqty, // 在庫数は Ajanqty を使う
		)
		if err != nil {
			return fmt.Errorf("failed to exec statement for inventory (JAN: %s): %w", rec.Ajc, err)
		}
	}

	return tx.Commit()
}
