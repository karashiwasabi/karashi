// File: db/product_master.go (Corrected)
package db

import (
	"database/sql"
	"fmt"
	"karashi/model"
	"karashi/units"
	"strings"
)

// selectColumns defines the common column list for reuse.
const selectColumns = `
	product_code, yj_code, product_name, origin, kana_name, maker_name, package_spec, 
	yj_unit_name, yj_pack_unit_qty, flag_poison, flag_deleterious, flag_narcotic, 
	flag_psychotropic, flag_stimulant, flag_stimulant_raw, jan_pack_inner_qty, 
	jan_unit_code, jan_pack_unit_qty, reorder_point, nhi_price`

// scanProductMaster is a helper function to scan a row into a ProductMaster struct.
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

// GetProductMasterByCode retrieves a master record by its product code.
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

// GetProductMastersByCodesMap returns a map of master records for multiple product codes.
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

// GetProductMasterByName retrieves a master record by its product name.
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

// CreateProductMasterInTx creates a master record within an existing transaction.
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

// ★★★ THIS FUNCTION IS NEW/REPLACES UpdateProductMasterInTx ★★★
// UpsertProductMasterInTx updates a master record or inserts it if it doesn't exist.
func UpsertProductMasterInTx(tx *sql.Tx, rec model.ProductMasterInput) error {
	const q = `INSERT INTO product_master (
            product_code, yj_code, product_name, origin, kana_name, maker_name, package_spec, 
            yj_unit_name, yj_pack_unit_qty, flag_poison, flag_deleterious, flag_narcotic, 
            flag_psychotropic, flag_stimulant, flag_stimulant_raw, jan_pack_inner_qty, 
            jan_unit_code, jan_pack_unit_qty, reorder_point, nhi_price
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(product_code) DO UPDATE SET
            yj_code=excluded.yj_code, product_name=excluded.product_name, origin=excluded.origin, 
            kana_name=excluded.kana_name, maker_name=excluded.maker_name, package_spec=excluded.package_spec, 
            yj_unit_name=excluded.yj_unit_name, yj_pack_unit_qty=excluded.yj_pack_unit_qty, 
            flag_poison=excluded.flag_poison, flag_deleterious=excluded.flag_deleterious, 
            flag_narcotic=excluded.flag_narcotic, flag_psychotropic=excluded.flag_psychotropic, 
            flag_stimulant=excluded.flag_stimulant, flag_stimulant_raw=excluded.flag_stimulant_raw, 
            jan_pack_inner_qty=excluded.jan_pack_inner_qty, jan_unit_code=excluded.jan_unit_code, 
            jan_pack_unit_qty=excluded.jan_pack_unit_qty, reorder_point=excluded.reorder_point, 
            nhi_price=excluded.nhi_price`

	_, err := tx.Exec(q,
		rec.ProductCode, rec.YjCode, rec.ProductName, rec.Origin, rec.KanaName, rec.MakerName, rec.PackageSpec,
		rec.YjUnitName, rec.YjPackUnitQty, rec.FlagPoison, rec.FlagDeleterious, rec.FlagNarcotic,
		rec.FlagPsychotropic, rec.FlagStimulant, rec.FlagStimulantRaw, rec.JanPackInnerQty,
		rec.JanUnitCode, rec.JanPackUnitQty, rec.ReorderPoint, rec.NhiPrice,
	)
	if err != nil {
		return fmt.Errorf("UpsertProductMasterInTx failed: %w", err)
	}
	return nil
}

// GetProductMasterByCodeInTx retrieves a master record by its product code within an existing transaction.
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

// GetAllProductMasters retrieves all product master records.
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

// SearchProductMastersByName returns a list of master records that match a partial product name.
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

// GetEditableProductMasters fetches all non-JCSHMS product masters.
func GetEditableProductMasters(conn *sql.DB) ([]model.ProductMasterView, error) {
	q := `SELECT ` + selectColumns + ` FROM product_master WHERE origin != 'JCSHMS' ORDER BY kana_name`

	rows, err := conn.Query(q)
	if err != nil {
		return nil, fmt.Errorf("GetEditableProductMasters failed: %w", err)
	}
	defer rows.Close()

	var mastersView []model.ProductMasterView
	for rows.Next() {
		m, err := scanProductMaster(rows)
		if err != nil {
			return nil, err
		}

		tempJcshms := model.JCShms{
			JC037: m.PackageSpec,
			JC039: m.YjUnitName,
			JC044: m.YjPackUnitQty,
			JA006: sql.NullFloat64{Float64: m.JanPackInnerQty, Valid: true},
			JA008: sql.NullFloat64{Float64: m.JanPackUnitQty, Valid: true},
			JA007: sql.NullString{String: fmt.Sprintf("%d", m.JanUnitCode), Valid: true},
		}
		formattedSpec := units.FormatPackageSpec(&tempJcshms)

		mastersView = append(mastersView, model.ProductMasterView{
			ProductMaster:        *m,
			FormattedPackageSpec: formattedSpec,
		})
	}
	return mastersView, nil
}
