// File: db/jcshms.go
package db

import (
	"database/sql"
	"errors"
	"log"
)

// JCShms は jcshms + jancode 結合結果。
// フィールド名を SQL列名そのままに揃えています。
type JCShms struct {
	JC000 string // 元JAN
	JC009 string // YJコード
	JC018 string // 品名
	JC022 string // 品名かな
	JC030 string // メーカー
	JC037 string // 包装ベース

	JC039 string // YJ単位名
	JC044 string // YJあたり数量

	JC061 int // 毒薬フラグ
	JC062 int // 劇薬フラグ
	JC063 int // 麻薬フラグ
	JC064 int // 向精神薬フラグ
	JC065 int // 覚せい剤フラグ
	JC066 int // 覚醒剤原料フラグ

	JA006 string // JAN単位名
	JA007 string // JAN単位コード
	JA008 string // JANあたり数量
}

func GetJcshmsByJan(conn *sql.DB, jan string) (*JCShms, error) {
	const q = `
SELECT
  jc.JC000, jc.JC009, jc.JC018, jc.JC022, jc.JC030, jc.JC037,
  jc.JC039, jc.JC044,
  jc.JC061, jc.JC062, jc.JC063, jc.JC064, jc.JC065, jc.JC066,
  jn.JA006, jn.JA007, jn.JA008
FROM jcshms jc
LEFT JOIN jancode jn ON jc.JC000 = jn.JA001
WHERE jc.JC000 = ?`

	var r JCShms
	err := conn.QueryRow(q, jan).Scan(
		&r.JC000,
		&r.JC009,
		&r.JC018,
		&r.JC022,
		&r.JC030,
		&r.JC037,
		&r.JC039,
		&r.JC044,
		&r.JC061,
		&r.JC062,
		&r.JC063,
		&r.JC064,
		&r.JC065,
		&r.JC066,
		&r.JA006,
		&r.JA007,
		&r.JA008,
	)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[GetJcshmsByJan] no record for JAN=%q", jan)
		return nil, sql.ErrNoRows
	}
	if err != nil {
		log.Printf("[GetJcshmsByJan] query error for JAN=%q: %v", jan, err)
		return nil, err
	}
	return &r, nil
}
