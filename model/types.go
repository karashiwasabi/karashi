// File: model/types.go
package model

// ARInputは、最終的にa_recordsテーブルに登録されるレコードの完全な構造体です。
type ARInput struct {
	Adate              string
	Apcode             string
	Arpnum             string
	Alnum              string
	Aflag              int
	Ajc                string
	Ayj                string
	Apname             string
	Akana              string
	Apkg               string
	Amaker             string
	Adatqty            float64
	Ajanqty            float64
	Ajpu               float64
	Ajanunitnm         string
	Ajanunitcode       string
	Ayjqty             float64
	Ayjpu              float64
	Ayjunitnm          string
	Aunitprice         float64
	Asubtotal          float64
	Ataxamt            float64
	Ataxrate           float64
	Aexpdate           float64
	Alot               string
	Adokuyaku          int
	Agekiyaku          int
	Amayaku            int
	Akouseisinyaku     int
	Akakuseizai        int
	Akakuseizaigenryou int
	Ama                string
}

// ParsedUsageは、USAGE CSVファイルの1行からパースされた生データです。
type ParsedUsage struct {
	Date       string
	Jc         string
	Yj         string
	Pname      string
	YjQty      float64
	YjUnitName string
}

// MaMasterは、ma_masterテーブルの1レコードの完全なデータです。
type MaMaster struct {
	MA000 string  // JAN or 合成JAN
	MA009 string  // YJコード
	MA018 string  // 品名
	MA022 string  // 品名かな
	MA030 string  // メーカー
	MA037 string  // 包装
	MA039 string  // YJ側単位名
	MA044 float64 // YJ側数量文字列
	MA061 int     // 毒薬フラグ
	MA062 int     // 劇薬フラグ
	MA063 int     // 麻薬フラグ
	MA064 int     // 向精神薬フラグ
	MA065 int     // 覚せい剤フラグ
	MA066 int     // 覚醒剤原料フラグ
	MA131 float64 //
	MA132 int     // JAN単位コード
	MA133 float64 // JANあたり数量
}

// MaMasterInputは、新しいma_masterレコードを作成するために必要なデータ構造です。
type MaMasterInput struct {
	// vvv ここから下が修正箇所 vvv
	MA000, MA009, MA018, MA022, MA030, MA037, MA039 string
	MA044, MA131, MA133                             float64
	MA061, MA062, MA063, MA064, MA065, MA066, MA132 int
}
