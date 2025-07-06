// File: ma0/record.go
package ma0

import "database/sql"

// Record は日常処理で使う主要9フィールドだけを持つ軽量版 struct です。
// フィールド名はテーブルのカラム名と完全一致しています。
type Record struct {
	MA000 string // JANコード
	MA009 string // YJコード
	MA018 string // 商品名
	MA037 string // 包装形態
	MA039 string // 包装単位
	MA044 string // 包装総量
	MA131 string // JAN数量
	MA132 string // JAN単位コード
	MA133 string // YJ換算数量
}

// QueryMA0 は指定された JAN をキーに Record を取得します。
func QueryMA0(db *sql.DB, jan string) (*Record, error) {
	const sqlstr = `
SELECT
  MA000,MA009,MA018,MA037,MA039,MA044,MA131,MA132,MA133
FROM ma0
WHERE MA000 = ?
`
	var r Record
	err := db.QueryRow(sqlstr, jan).Scan(
		&r.MA000,
		&r.MA009,
		&r.MA018,
		&r.MA037,
		&r.MA039,
		&r.MA044,
		&r.MA131,
		&r.MA132,
		&r.MA133,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}
