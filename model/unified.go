package model

// UnifiedInputRecordは、すべての入力ソース（DAT,USAGE,INV,IOD）の
// 全項目を網羅したスーパーセットとなる構造体です。
// 各パーサーはこの構造体を生成し、中央処理関数に渡します。
type UnifiedInputRecord struct {
	// --- 伝票・日付情報 ---
	Date          string `json:"date"`
	ClientCode    string `json:"clientCode"`
	ReceiptNumber string `json:"receiptNumber"`
	LineNumber    string `json:"lineNumber"`
	Flag          int    `json:"flag"`
	ExpiryDate    string `json:"expiryDate"` // ← float64 から string に変更
	LotNumber     string `json:"lotNumber"`

	// --- 薬品コード・名称 ---
	JanCode     string `json:"janCode"`
	YjCode      string `json:"yjCode"`
	ProductName string `json:"productName"`
	KanaName    string `json:"kanaName"`
	PackageSpec string `json:"packageSpec"`
	MakerName   string `json:"makerName"`

	// --- 数量 ---
	DatQuantity     float64 `json:"datQuantity"`
	JanPackInnerQty float64 `json:"janPackInnerQty"` // ✨ この行を追加
	JanQuantity     float64 `json:"janQuantity"`
	JanPackUnitQty  float64 `json:"janPackUnitQty"`
	YjQuantity      float64 `json:"yjQuantity"`
	YjPackUnitQty   float64 `json:"yjPackUnitQty"`

	// --- 単位 ---
	JanUnitName string `json:"janUnitName"`
	JanUnitCode string `json:"janUnitCode"`
	YjUnitName  string `json:"yjUnitName"`

	// --- 金額 ---
	UnitPrice float64 `json:"unitPrice"`
	Subtotal  float64 `json:"subtotal"`
	TaxAmount float64 `json:"taxAmount"`
	TaxRate   float64 `json:"taxRate"`

	// --- 薬事区分 ---
	FlagPoison       int `json:"flagPoison"`
	FlagDeleterious  int `json:"flagDeleterious"`
	FlagNarcotic     int `json:"flagNarcotic"`
	FlagPsychotropic int `json:"flagPsychotropic"`
	FlagStimulant    int `json:"flagStimulant"`
	FlagStimulantRaw int `json:"flagStimulantRaw"`

	// --- 処理結果フラグ ---
	ProcessFlagMA string `json:"processFlagMA"`
}
