// 統合明細構造体とDB登録処理
package unifiedrecords

import (
	"database/sql"
)

// Record は unifiedrecords テーブルに対応する1行
type Record struct {
	SlipDate       string  `json:"slipdate"`
	JANCode        string  `json:"jancode"`
	YJCode         string  `json:"yjcode"`
	ProductName    string  `json:"productname"`
	Packaging      string  `json:"packaging"`
	DATQty         int     `json:"datqty"`
	JANQuantity    int     `json:"janquantity"`
	JANUnitName    string  `json:"janunitname"`
	JANUnitCode    string  `json:"janunitcode"`
	YJQuantity     float64 `json:"yjquantity"`
	YJUnitName     string  `json:"yjunitname"`
	UnitPrice      float64 `json:"unitprice"`
	SubtotalAmount float64 `json:"subtotalamount"`
	TaxAmount      float64 `json:"taxamount"`
	TaxRate        float64 `json:"taxrate"`
	ExpiryDate     string  `json:"expirydate"`
	LotNumber      string  `json:"lotnumber"`
	ReceiptNumber  string  `json:"receiptnumber"`
	LineNumber     string  `json:"linenumber"`
	Flag           int     `json:"flag"`
	PartnerCode    string  `json:"partnercode"`
}

// Insert は Record を unifiedrecords テーブルに登録する
func Insert(db *sql.DB, rec Record) error {
	const stmt = `
	INSERT INTO unifiedrecords (
		slipdate, jancode, yjcode, productname, packaging,
		datqty, janquantity, janunitname, janunitcode,
		yjquantity, yjunitname,
		unitprice, subtotalamount, taxamount, taxrate,
		expirydate, lotnumber, receiptnumber,
		linenumber, flag, partnercode
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,  ?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(stmt,
		rec.SlipDate, rec.JANCode, rec.YJCode, rec.ProductName, rec.Packaging,
		rec.DATQty, rec.JANQuantity, rec.JANUnitName, rec.JANUnitCode,
		rec.YJQuantity, rec.YJUnitName,
		rec.UnitPrice, rec.SubtotalAmount, rec.TaxAmount, rec.TaxRate,
		rec.ExpiryDate, rec.LotNumber, rec.ReceiptNumber,
		rec.LineNumber, rec.Flag, rec.PartnerCode,
	)
	return err
}
