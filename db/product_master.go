// File: db/product_master.go
package db

import (
	"database/sql"
	"fmt"
	"karashi/model"
	"strings"
)

// ★★★ 全てのSELECT文で使う共通の列リストを定義 ★★★
const selectColumns = `
	product_code, yj_code, product_name, origin, kana_name, maker_name, package_spec, 
	yj_unit_name, yj_pack_unit_qty, flag_poison, flag_deleterious, flag_narcotic, 
	flag_psychotropic, flag_stimulant, flag_stimulant_raw, jan_pack_inner_qty, 
	jan_unit_code, jan_pack_unit_qty, reorder_point, nhi_price`

// ★★★ 共通のスキャン関数を定義 ★★★
func scanProductMaster(row interface{ Scan(...interface{}) error }) (*model.ProductMaster, error) {
	var m model.ProductMaster
	err := row.Scan(
		&m.ProductCode, &m.YjCode, &m.ProductName, &m.Origin, &m.KanaName, &m.MakerName, &m.PackageSpec,
		&m.YjUnitName, &m.YjPackUnitQty, &m.FlagPoison, &m.FlagDeleterious, &m.FlagNarcotic,
		&m.FlagPsychotropic, &m.FlagStimulant, &m.FlagStimulantRaw, &m.JanPackInnerQty,
		&m.JanUnitCode, &m.JanPackUnitQty, &m.ReorderPoint, &m.NhiPrice,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetProductMasterByCodeは、製品コードをキーにマスターを取得します。
func GetProductMasterByCode(conn *sql.DB, code string) (*model.ProductMaster, error) {
	q := `SELECT ` + selectColumns + ` FROM product_master WHERE product_code = ? LIMIT 1`
	m, err := scanProductMaster(conn.QueryRow(q, code))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetProductMasterByCode failed: %w", err)
	}
	return m, nil
}

// GetProductMastersByCodesMapは、複数の製品コードをキーにマスターのマップを返します。
func GetProductMastersByCodesMap(conn *sql.DB, codes []string) (map[string]*model.ProductMaster, error) {
	if len(codes) == 0 {
		return make(map[string]*model.ProductMaster), nil
	}
	q := `SELECT ` + selectColumns + ` FROM product_master WHERE product_code IN (?` + strings.Repeat(",?", len(codes)-1) + `)`
	args := make([]interface{}, len(codes))
	for i, code := range codes {
		args[i] = code
	}
	rows, err := conn.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("query for masters by codes failed: %w", err)
	}
	defer rows.Close()
	mastersMap := make(map[string]*model.ProductMaster)
	for rows.Next() {
		m, err := scanProductMaster(rows)
		if err != nil {
			return nil, err
		}
		mastersMap[m.ProductCode] = m
	}
	return mastersMap, nil
}

// GetProductMasterByNameは、製品名をキーにマスターを取得します。
func GetProductMasterByName(conn *sql.DB, name string) (*model.ProductMaster, error) {
	q := `SELECT ` + selectColumns + ` FROM product_master WHERE product_name = ? LIMIT 1`
	m, err := scanProductMaster(conn.QueryRow(q, name))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetProductMasterByName failed: %w", err)
	}
	return m, nil
}

// CreateProductMasterInTxは、既存のトランザクション内でマスターを作成します。
func CreateProductMasterInTx(tx *sql.Tx, rec model.ProductMasterInput) error {
	const q = `INSERT INTO product_master (
			product_code, yj_code, product_name, origin, kana_name, maker_name, package_spec, 
			yj_unit_name, yj_pack_unit_qty, flag_poison, flag_deleterious, flag_narcotic, 
			flag_psychotropic, flag_stimulant, flag_stimulant_raw, jan_pack_inner_qty, 
			jan_unit_code, jan_pack_unit_qty, reorder_point, nhi_price
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.Exec(q,
		rec.ProductCode, rec.YjCode, rec.ProductName, rec.Origin, rec.KanaName, rec.MakerName, rec.PackageSpec,
		rec.YjUnitName, rec.YjPackUnitQty, rec.FlagPoison, rec.FlagDeleterious, rec.FlagNarcotic,
		rec.FlagPsychotropic, rec.FlagStimulant, rec.FlagStimulantRaw, rec.JanPackInnerQty,
		rec.JanUnitCode, rec.JanPackUnitQty, rec.ReorderPoint, rec.NhiPrice,
	)
	if err != nil {
		return fmt.Errorf("CreateProductMasterInTx failed: %w", err)
	}
	return nil
}

// UpdateProductMasterInTxは、既存のトランザクション内でマスターを更新します。
func UpdateProductMasterInTx(tx *sql.Tx, rec model.ProductMasterInput) error {
	const q = `UPDATE product_master SET
            yj_code = ?, product_name = ?, origin = ?, kana_name = ?, maker_name = ?, package_spec = ?, 
            yj_unit_name = ?, yj_pack_unit_qty = ?, flag_poison = ?, flag_deleterious = ?, 
            flag_narcotic = ?, flag_psychotropic = ?, flag_stimulant = ?, flag_stimulant_raw = ?, 
            jan_pack_inner_qty = ?, jan_unit_code = ?, jan_pack_unit_qty = ?, reorder_point = ?, nhi_price = ?
			WHERE product_code = ?`
	_, err := tx.Exec(q,
		rec.YjCode, rec.ProductName, rec.Origin, rec.KanaName, rec.MakerName, rec.PackageSpec,
		rec.YjUnitName, rec.YjPackUnitQty, rec.FlagPoison, rec.FlagDeleterious,
		rec.FlagNarcotic, rec.FlagPsychotropic, rec.FlagStimulant, rec.FlagStimulantRaw,
		rec.JanPackInnerQty, rec.JanUnitCode, rec.JanPackUnitQty, rec.ReorderPoint, rec.NhiPrice,
		rec.ProductCode,
	)
	if err != nil {
		return fmt.Errorf("UpdateProductMasterInTx failed: %w", err)
	}
	return nil
}

// GetProductMasterByCodeInTxは、既存のトランザクション内で製品コードをキーにマスターを取得します。
func GetProductMasterByCodeInTx(tx *sql.Tx, code string) (*model.ProductMaster, error) {
	q := `SELECT ` + selectColumns + ` FROM product_master WHERE product_code = ? LIMIT 1`
	m, err := scanProductMaster(tx.QueryRow(q, code))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetProductMasterByCodeInTx failed: %w", err)
	}
	return m, nil
}

// GetAllProductMastersは、製品マスターを全件取得します。
func GetAllProductMasters(conn *sql.DB) ([]*model.ProductMaster, error) {
	q := `SELECT ` + selectColumns + ` FROM product_master ORDER BY kana_name`

	rows, err := conn.Query(q)
	if err != nil {
		return nil, fmt.Errorf("GetAllProductMasters failed: %w", err)
	}
	defer rows.Close()

	var masters []*model.ProductMaster
	for rows.Next() {
		m, err := scanProductMaster(rows)
		if err != nil {
			return nil, err
		}
		masters = append(masters, m)
	}
	return masters, nil
}

// SearchProductMastersByNameは、製品名の一部に一致するマスターのリストを返します。
func SearchProductMastersByName(conn *sql.DB, nameQuery string) ([]*model.ProductMaster, error) {
	q := `SELECT ` + selectColumns + ` FROM product_master WHERE product_name LIKE ? OR kana_name LIKE ? ORDER BY kana_name LIMIT 100`

	rows, err := conn.Query(q, "%"+nameQuery+"%", "%"+nameQuery+"%")
	if err != nil {
		return nil, fmt.Errorf("SearchProductMastersByName failed: %w", err)
	}
	defer rows.Close()

	var masters []*model.ProductMaster
	for rows.Next() {
		m, err := scanProductMaster(rows)
		if err != nil {
			return nil, err
		}
		masters = append(masters, m)
	}
	return masters, nil
}
