// File: dat/types.go
package dat

// ParsedDatは、DATファイルのD行からパースされたレコードを表します。
type ParsedDat struct {
	WholesaleCode string
	DatDate       string
	DeliveryFlag  string
	ReceiptNumber string
	LineNumber    string
	JanCode       string
	ProductName   string
	Quantity      string
	UnitPrice     string
	Subtotal      string
	ExpiryDate    string
	LotNumber     string
}

// MarshalJSON は不要になったため削除しました。
