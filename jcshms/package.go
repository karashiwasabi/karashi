// File: jcshms/package.go
package jcshms

import (
	"database/sql"
	"fmt"
)

// JCRecord は jcshms テーブルの 1 行を表します。
// 必要なフィールドのみ宣言してください。
type JCRecord struct {
	JC000 string // JANコード
	JC009 string // YJコード
	JC018 string // 商品名
	JC022 string // 商品名かなソートキー
	JC037 string // 包装形態
	JC039 string // 包装単位コード
	JC044 string // 包装総量
	JC061 string // 毒薬フラグ
	JC062 string // 劇薬フラグ
	JC063 string // 麻薬フラグ
	JC064 string // 向精神薬フラグ
}

// QueryByJan は jcshms テーブルから key=jan のレコードを返します。
func QueryByJan(db *sql.DB, jan string) (*JCRecord, error) {
	const sqlstr = `
SELECT
  JC000,
  JC009,
  JC018,
  JC022,
  JC037,
  JC039,
  JC044,
  JC061,
  JC062,
  JC063,
  JC064
FROM jcshms
WHERE JC000 = ?
`
	var r JCRecord
	err := db.QueryRow(sqlstr, jan).Scan(
		&r.JC000,
		&r.JC009,
		&r.JC018,
		&r.JC022,
		&r.JC037,
		&r.JC039,
		&r.JC044,
		&r.JC061,
		&r.JC062,
		&r.JC063,
		&r.JC064,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("jcshms lookup failed: %w", err)
	}
	return &r, nil
}
