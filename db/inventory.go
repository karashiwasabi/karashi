// File: db/inventory.go
package db

import (
	"database/sql"
	"fmt"
	"karashi/model"
	"log"
)

// PersistInventoryRecordsInTxは、棚卸レコードを新しいinventory_recordsテーブルに保存します
func PersistInventoryRecordsInTx(tx *sql.Tx, records []model.TransactionRecord) error {
	const q = `
INSERT OR REPLACE INTO inventory_records (
    transaction_date, client_code, receipt_number, line_number, flag,
    jan_code, yj_code, product_name, kana_name, package_form, package_spec, maker_name,
    dat_quantity, jan_pack_inner_qty, jan_quantity, jan_pack_unit_qty, jan_unit_name, jan_unit_code,
    yj_quantity, yj_pack_unit_qty, yj_unit_name, unit_price, subtotal,
    tax_amount, tax_rate, expiry_date, lot_number, flag_poison,
    flag_deleterious, flag_narcotic, flag_psychotropic, flag_stimulant,
    flag_stimulant_raw, process_flag_ma, processing_status
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

	stmt, err := tx.Prepare(q)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for inventory_records: %w", err)
	}
	defer stmt.Close()

	for _, rec := range records {
		_, err := stmt.Exec(
			rec.TransactionDate, rec.ClientCode, rec.ReceiptNumber, rec.LineNumber, rec.Flag,
			rec.JanCode, rec.YjCode, rec.ProductName, rec.KanaName, rec.PackageForm, rec.PackageSpec, rec.MakerName,
			rec.DatQuantity, rec.JanPackInnerQty, rec.JanQuantity,
			rec.JanPackUnitQty,
			rec.JanUnitName, rec.JanUnitCode,
			rec.YjQuantity, rec.YjPackUnitQty, rec.YjUnitName, rec.UnitPrice, rec.Subtotal,
			rec.TaxAmount, rec.TaxRate, rec.ExpiryDate, rec.LotNumber, rec.FlagPoison,
			rec.FlagDeleterious, rec.FlagNarcotic, rec.FlagPsychotropic, rec.FlagStimulant,
			rec.FlagStimulantRaw, rec.ProcessFlagMA, rec.ProcessingStatus,
		)
		if err != nil {
			log.Printf("!!! FAILED to insert into inventory_records: JAN=%s, Error: %v", rec.JanCode, err)
			return fmt.Errorf("failed to exec statement for inventory_records (JAN: %s): %w", rec.JanCode, err)
		}
	}
	return nil
}

// ▼▼▼ ここから追記 ▼▼▼

// GetLatestInventoryRecord は指定日以前の最新の棚卸レコードを取得します。
func GetLatestInventoryRecord(conn *sql.DB, janCode, date string) (*model.TransactionRecord, error) {
	const q = `
		SELECT ` + TransactionColumns + ` FROM inventory_records
		WHERE jan_code = ? AND transaction_date <= ?
		ORDER BY transaction_date DESC
		LIMIT 1`

	row := conn.QueryRow(q, janCode, date)
	rec, err := ScanTransactionRecord(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 履歴がない場合はエラーではなくnilを返す
		}
		return nil, err
	}
	return rec, nil
}

// ▲▲▲ ここまで追記 ▲▲▲
