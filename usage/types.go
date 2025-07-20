// File: usage/types.go
package usage

// ―――――――――――――――――――
// Usage CSV の列インデックス定義
//
// フォーマット（ヘッダーなし）:
//   {Date,Yj,Jc,Pname,YjQty,YjUnitName}
const (
	UsageDateIndex       = iota // CSV[0]
	UsageYjIndex                // CSV[1]
	UsageJcIndex                // CSV[2]
	UsagePnameIndex             // CSV[3]
	UsageYjQtyIndex             // CSV[4]
	UsageYjUnitNameIndex        // CSV[5]
)

// ParsedUsage は Shift_JIS→UTF-8 変換後の 1 行分レコードです。
// フィールド順: {Date,Jc,Yj,Pname,YjQty,YjUnitName}
type ParsedUsage struct {
	Date       string  // CSV[UsageDateIndex]
	Jc         string  // CSV[UsageJcIndex]
	Yj         string  // CSV[UsageYjIndex]
	Pname      string  // CSV[UsagePnameIndex]
	YjQty      float64 // CSV[UsageYjQtyIndex]
	YjUnitName string  // CSV[UsageYjUnitNameIndex]
}

// ――― brusage フェーズ用定数 ―――
const BranchAflag = 3

// Ama（グループ番号）定義
const (
	Ama1 = "1"
	Ama2 = "2"
	Ama3 = "3"
	Ama4 = "4"
	Ama5 = "5"
	Ama6 = "6"
)

// BrUsage は BranchUsage の出力型です。
// フィールド順: {Date,3,Jc,Yj,Pname,YjQty,YjUnitName,Ama}
type BrUsage struct {
	Date       string  // CSV[UsageDateIndex]
	Aflag      int     // = BranchAflag (3)
	Jc         string  // CSV[UsageJcIndex]
	Yj         string  // CSV[UsageYjIndex]
	Pname      string  // CSV[UsagePnameIndex]
	YjQty      float64 // CSV[UsageYjQtyIndex]
	YjUnitName string  // CSV[UsageYjUnitNameIndex]
	Ama        string  // "1"～"6"
}

// ARInput は a_records 登録前の共通入力型です。
// フィールド順: [
//   Adate,Apcode,Arpnum,Alnum,Aflag,Ajc,Ayj,Apname,Akana,Apkg,Amaker,
//   Adatqty,Ajanqty,Ajpu,Ajanunitnm,Ajanunitcd,Ayjqty,Ayjpu,Ayjunitnm,
//   Aunitprice,Asubtotal,Ataxamt,Ataxrate,Aexpdate,Alot,
//   Adokuyaku,Agekiyaku,Amayaku,Akouseisinyaku,Akakuseizai,Akakuseizaigenryou,Ama
// ]
type ARInput struct {
	Adate              string  // adate               日付1
	Apcode             string  // apcode              得意先コード2
	Arpnum             string  // arpnum              伝票番号3
	Alnum              string  // alnum               行番号4
	Aflag              int     // aflag               種別フラグ5
	Ajc                string  // ajc                 JANコード6
	Ayj                string  // ayj                 YJコード7
	Apname             string  // apname              品名8
	Akana              string  // akana               品名かな9
	Apkg               string  // apkg                包装10
	Amaker             string  // amaker              メーカー11
	Adatqty            float64 // adatqty             DAT数量12
	Ajanqty            float64 // ajanqty             JAN数量13
	Ajpu               string  // ajpu                JAN単位コード14
	Ajanunitnm         string  // ajanunitname        JAN単位名称15
	Ajanunitcd         string  // ajanunitcode        JAN単位コード16
	Ayjqty             float64 // ayjqty              YJ数量17
	Ayjpu              string  // ayjpu               YJ単位コード18
	Ayjunitnm          string  // ayjunitname         YJ単位名称19
	Aunitprice         float64 // aunitprice          単価20
	Asubtotal          float64 // asubtotal           小計21
	Ataxamt            float64 // ataxamount          税額22
	Ataxrate           string  // ataxrate            税率23
	Aexpdate           string  // aexpdate            有効期限24
	Alot               string  // alot                ロット番号25
	Adokuyaku          int     // adokuyaku           毒薬フラグ26
	Agekiyaku          int     // agekiyaku           劇薬フラグ27
	Amayaku            int     // amayaku             麻薬フラグ28
	Akouseisinyaku     int     // akouseisinyaku      向精神薬フラグ29
	Akakuseizai        int     // akakuseizai         覚せい剤フラグ30
	Akakuseizaigenryou int     // akakuseizaigenryou  覚醒剤原料フラグ31
	Ama                string  // ama                 ルート番号32
}

// MaMaster は ma_master テーブルの 1 レコードを表現します。
// フィールド順: [MA000,MA009,MA018,MA022,MA030,MA037,MA039,MA044,
//               MA061,MA062,MA063,MA064,MA065,MA066,MA131,MA132,MA133]
type MaMaster struct {
	MA000 string `db:"MA000" json:"MA000"`
	MA009 string `db:"MA009" json:"MA009"`
	MA018 string `db:"MA018" json:"MA018"`
	MA022 string `db:"MA022" json:"MA022"`
	MA030 string `db:"MA030" json:"MA030"`
	MA037 string `db:"MA037" json:"MA037"`
	MA039 string `db:"MA039" json:"MA039"`
	MA044 string `db:"MA044" json:"MA044"`
	MA061 string `db:"MA061" json:"MA061"`
	MA062 string `db:"MA062" json:"MA062"`
	MA063 string `db:"MA063" json:"MA063"`
	MA064 string `db:"MA064" json:"MA064"`
	MA065 string `db:"MA065" json:"MA065"`
	MA066 string `db:"MA066" json:"MA066"`
	MA131 string `db:"MA131" json:"MA131"`
	MA132 string `db:"MA132" json:"MA132"`
	MA133 string `db:"MA133" json:"MA133"`
}
