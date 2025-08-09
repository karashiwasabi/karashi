// File: model/types.go (Corrected)
package model

import "database/sql"

// StockLedgerYJGroupは、在庫台帳のYJコードごとのグループです。
type StockLedgerYJGroup struct {
	YjCode            string                    `json:"yjCode"`
	ProductName       string                    `json:"productName"`
	YjUnitName        string                    `json:"yjUnitName"` // ★ 追加
	PackageLedgers    []StockLedgerPackageGroup `json:"packageLedgers"`
	StartingBalance   float64                   `json:"startingBalance"`
	NetChange         float64                   `json:"netChange"`
	EndingBalance     float64                   `json:"endingBalance"`
	TotalReorderPoint float64                   `json:"totalReorderPoint"`
	IsReorderNeeded   bool                      `json:"isReorderNeeded"`
}

// StockLedgerPackageGroupは、包装ごとの在庫台帳です。
type StockLedgerPackageGroup struct {
	PackageKey      string              `json:"packageKey"`
	JanUnitName     string              `json:"janUnitName"` // ★ 追加
	StartingBalance float64             `json:"startingBalance"`
	Transactions    []LedgerTransaction `json:"transactions"`
	NetChange       float64             `json:"netChange"`
	EndingBalance   float64             `json:"endingBalance"`
	MaxUsage        float64             `json:"maxUsage"`
	ReorderPoint    float64             `json:"reorderPoint"`
	IsReorderNeeded bool                `json:"isReorderNeeded"`
}

// LedgerTransactionは、在庫推移計算後の個々のトランザクションです。
type LedgerTransaction struct {
	TransactionRecord
	RunningBalance float64 `json:"runningBalance"`
}

// AggregationFilters は集計時のフィルター条件を保持します
type AggregationFilters struct {
	StartDate   string
	EndDate     string
	KanaName    string
	DrugTypes   []string
	NoMovement  bool
	Coefficient float64
}

// YJGroup はYJコードごとの集計結果です
type YJGroup struct {
	YjCode        string         `json:"yjCode"`
	ProductName   string         `json:"productName"`
	TotalJanQty   float64        `json:"totalJanQty"`
	TotalYjQty    float64        `json:"totalYjQty"`
	MaxUsageYjQty float64        `json:"maxUsageYjQty"`
	PackageGroups []PackageGroup `json:"packageGroups"`
}

// PackageGroup は包装ごとの集計結果です
type PackageGroup struct {
	PackageKey     string              `json:"packageKey"`
	TotalJanQty    float64             `json:"totalJanQty"`
	MaxUsageJanQty float64             `json:"maxUsageJanQty"`
	TotalYjQty     float64             `json:"totalYjQty"`
	MaxUsageYjQty  float64             `json:"maxUsageYjQty"`
	Transactions   []TransactionRecord `json:"transactions"`
}

type JCShms struct {
	JC009 string
	JC018 string
	JC022 string
	JC030 string
	JC037 string
	JC039 string
	JC044 float64
	JC050 float64
	JC061 int
	JC062 int
	JC063 int
	JC064 int
	JC065 int
	JC066 int
	JA006 sql.NullFloat64
	JA007 sql.NullString
	JA008 sql.NullFloat64
}

type TransactionRecord struct {
	ID               int            `json:"id"`
	TransactionDate  string         `json:"transactionDate"`
	ClientCode       string         `json:"clientCode"`
	ReceiptNumber    string         `json:"receiptNumber"`
	LineNumber       string         `json:"lineNumber"`
	Flag             int            `json:"flag"`
	JanCode          string         `json:"janCode"`
	YjCode           string         `json:"yjCode"`
	ProductName      string         `json:"productName"`
	KanaName         string         `json:"kanaName"`
	PackageForm      string         `json:"packageForm"`
	PackageSpec      string         `json:"packageSpec"`
	MakerName        string         `json:"makerName"`
	DatQuantity      float64        `json:"datQuantity"`
	JanPackInnerQty  float64        `json:"janPackInnerQty"`
	JanQuantity      float64        `json:"janQuantity"`
	JanPackUnitQty   float64        `json:"janPackUnitQty"`
	JanUnitName      string         `json:"janUnitName"`
	JanUnitCode      string         `json:"janUnitCode"`
	YjQuantity       float64        `json:"yjQuantity"`
	YjPackUnitQty    float64        `json:"yjPackUnitQty"`
	YjUnitName       string         `json:"yjUnitName"`
	UnitPrice        float64        `json:"unitPrice"`
	Subtotal         float64        `json:"subtotal"`
	TaxAmount        float64        `json:"taxAmount"`
	TaxRate          float64        `json:"taxRate"`
	ExpiryDate       string         `json:"expiryDate"`
	LotNumber        string         `json:"lotNumber"`
	FlagPoison       int            `json:"flagPoison"`
	FlagDeleterious  int            `json:"flagDeleterious"`
	FlagNarcotic     int            `json:"flagNarcotic"`
	FlagPsychotropic int            `json:"flagPsychotropic"`
	FlagStimulant    int            `json:"flagStimulant"`
	FlagStimulantRaw int            `json:"flagStimulantRaw"`
	ProcessFlagMA    string         `json:"processFlagMA"`
	ProcessingStatus sql.NullString `json:"processingStatus"`
}

type ProductMaster struct {
	ProductCode      string  `json:"productCode"`
	YjCode           string  `json:"yjCode"`
	ProductName      string  `json:"productName"`
	Origin           string  `json:"origin"`
	KanaName         string  `json:"kanaName"`
	MakerName        string  `json:"makerName"`
	PackageSpec      string  `json:"packageSpec"`
	YjUnitName       string  `json:"yjUnitName"`
	YjPackUnitQty    float64 `json:"yjPackUnitQty"`
	FlagPoison       int     `json:"flagPoison"`
	FlagDeleterious  int     `json:"flagDeleterious"`
	FlagNarcotic     int     `json:"flagNarcotic"`
	FlagPsychotropic int     `json:"flagPsychotropic"`
	FlagStimulant    int     `json:"flagStimulant"`
	FlagStimulantRaw int     `json:"flagStimulantRaw"`
	JanPackInnerQty  float64 `json:"janPackInnerQty"`
	JanUnitCode      int     `json:"janUnitCode"`
	JanPackUnitQty   float64 `json:"janPackUnitQty"`
	JanUnitName      string  `json:"janUnitName"`
	ReorderPoint     float64 `json:"reorderPoint"`
	NhiPrice         float64 `json:"nhiPrice"`
}

type ProductMasterInput struct {
	ProductCode      string  `json:"productCode"`
	YjCode           string  `json:"yjCode"`
	ProductName      string  `json:"productName"`
	Origin           string  `json:"origin"`
	KanaName         string  `json:"kanaName"`
	MakerName        string  `json:"makerName"`
	PackageSpec      string  `json:"packageSpec"`
	YjUnitName       string  `json:"yjUnitName"`
	YjPackUnitQty    float64 `json:"yjPackUnitQty"`
	FlagPoison       int     `json:"flagPoison"`
	FlagDeleterious  int     `json:"flagDeleterious"`
	FlagNarcotic     int     `json:"flagNarcotic"`
	FlagPsychotropic int     `json:"flagPsychotropic"`
	FlagStimulant    int     `json:"flagStimulant"`
	FlagStimulantRaw int     `json:"flagStimulantRaw"`
	JanPackInnerQty  float64 `json:"janPackInnerQty"`
	JanUnitCode      int     `json:"janUnitCode"`
	JanPackUnitQty   float64 `json:"janPackUnitQty"`
	JanUnitName      string  `json:"janUnitName"`
	ReorderPoint     float64 `json:"reorderPoint"`
	NhiPrice         float64 `json:"nhiPrice"`
}

type ProductMasterView struct {
	ProductMaster
	FormattedPackageSpec string `json:"formattedPackageSpec"`
}

type Client struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
