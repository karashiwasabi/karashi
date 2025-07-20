// File: model/types.go
package model

// UploadedDAT は /uploadDat ハンドラが受け取る生行データを表します。
type UploadedDAT struct {
	Line string // CSV/Dat 一行分の生テキスト
}

// ParsedDAT は .dat ファイルを固定長分解して得られるレコードです。
// Usage ワークフローにも流用できるよう、YJQty/YJUnit を追加しています。
type ParsedDAT struct {
	SlipDate      string  // 処理日付 (YYYYMMDD)
	PartnerCode   string  // 伝票上の得意先コード
	ReceiptNumber string  // 伝票番号
	LineNumber    string  // 行番号
	Flag          int     // 種別フラグ
	JANCode       string  // JANコード
	ProductName   string  // 製品名
	DatQty        float64 // DAT由来の数量
	YJQty         float64 // YJ数量 (Usage/Branch用)
	YJUnit        string  // YJ単位 (Usage/Branch用)
	UnitPrice     float64 // 単価
	Subtotal      float64 // 小計
	ExpiryDate    string  // 有効期限 (YYYYMMDD)
	LotNumber     string  // ロット番号
}

// ParsedUsage は Usage CSV を Shift_JIS→UTF-8 で読み込んだあと得られるレコードです。
// CSV列順: Date, Jc, Yj, Pname, YjQty, YjUnitName
type ParsedUsage struct {
	Date       string  // CSV[0]: Date (YYYYMMDD)
	Jc         string  // CSV[1]: 元JANコード
	Yj         string  // CSV[2]: YJコード
	Pname      string  // CSV[3]: 品名
	YjQty      float64 // CSV[4]: YJ数量
	YjUnitName string  // CSV[5]: YJ単位名称
}

// ARInput は a_records 登録前の共通入力型です。
// Branch→MA→DA すべてで受け渡します。
type ARInput struct {
	Adate  string // adate       日付1
	Apcode string // apcode      得意先コード2
	Arpnum string // arpnum      伝票番号3
	Alnum  string // alnum       行番号4
	Aflag  int    // aflag       種別フラグ5

	Ajc string // ajc         JANコード6
	Ayj string // ayj         YJコード7

	Apname             string  // 品名8
	Akana              string  // 品名かな9
	Apkg               string  // 包装10
	Amaker             string  // メーカー11
	Adatqty            float64 // DAT数量12
	Ajanqty            float64 // JAN数量13
	Ajpu               string  // JAN単位コード14
	Ajanunitname       string  // JAN単位名称15
	Ajanunitcode       string  // JAN単位コード16
	Ayjqty             float64 // YJ数量17
	Ayjpu              string  // YJ単位コード18
	Ayjunitname        string  // YJ単位名称19
	Aunitprice         float64 // 単価20
	Asubtotal          float64 // 小計21
	Ataxamount         float64 // 税額22
	Ataxrate           string  // 税率23
	Aexpdate           string  // 有効期限24
	Alot               string  // ロット番号25
	Adokuyaku          int     // 毒薬フラグ26
	Agekiyaku          int     // 劇薬フラグ27
	Amayaku            int     // 麻薬フラグ28
	Akouseisinyaku     int     // 向精神薬フラグ29
	Akakuseizai        int     // 覚せい剤フラグ30
	Akakuseizaigenryou int     // 覚醒剤原料フラグ31

	Ama string // 1～6 のルート番号32
}

// ARResult は DAハンドラが返す結果エイリアスです。
type ARResult = ARInput

// UsageAflag は brusage フェーズで設定する aflag 値です。
const UsageAflag = 3
