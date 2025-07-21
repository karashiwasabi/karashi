// File: db/jcshms.go
package db

import (
	"database/sql"
	"fmt"
)

// JCShmsは、jcshmsとjancodeテーブルを結合した結果を保持する構造体です。
type JCShms struct {
	JC009 string
	JC018 string
	JC022 string
	JC030 string
	JC037 string
	JC039 string
	JC044 float64 // vvv 修正点: stringからfloat64へ
	JC061 int
	JC062 int
	JC063 int
	JC064 int
	JC065 int
	JC066 int
	// vvv ここから下がjancode由来の修正点 vvv
	JA006 sql.NullFloat64 // JAN側 包装内数量
	JA007 sql.NullString  // JAN側 単位コード (文字列のまま)
	JA008 sql.NullFloat64 // JAN側 包装単位での数量
}

// GetJcshmsByJanは、JANコードを元にjcshmsとjancodeのデータをそれぞれ検索し、結果を結合します。
func GetJcshmsByJan(conn *sql.DB, jan string) (*JCShms, error) {
	// 最終的に返すための構造体
	finalRec := &JCShms{}
	found := false // どちらかのテーブルでデータが見つかったか

	// --- 1. jcshmsテーブルをJC000で検索 ---
	var jcshmsPart struct {
		JC009, JC018, JC022, JC030, JC037, JC039 string
		JC044                                    float64
		JC061, JC062, JC063, JC064, JC065, JC066 int
	}
	const q1 = `SELECT JC009, JC018, JC022, JC030, JC037, JC039, JC044, JC061, JC062, JC063, JC064, JC065, JC066
	              FROM jcshms WHERE JC000 = ? LIMIT 1`
	err1 := conn.QueryRow(q1, jan).Scan(
		&jcshmsPart.JC009, &jcshmsPart.JC018, &jcshmsPart.JC022, &jcshmsPart.JC030,
		&jcshmsPart.JC037, &jcshmsPart.JC039, &jcshmsPart.JC044, &jcshmsPart.JC061,
		&jcshmsPart.JC062, &jcshmsPart.JC063, &jcshmsPart.JC064, &jcshmsPart.JC065, &jcshmsPart.JC066,
	)
	if err1 != nil && err1 != sql.ErrNoRows {
		return nil, fmt.Errorf("jcshms search failed: %w", err1)
	}
	if err1 == nil {
		found = true
		finalRec.JC009 = jcshmsPart.JC009
		finalRec.JC018 = jcshmsPart.JC018
		finalRec.JC022 = jcshmsPart.JC022
		finalRec.JC030 = jcshmsPart.JC030
		finalRec.JC037 = jcshmsPart.JC037
		finalRec.JC039 = jcshmsPart.JC039
		finalRec.JC044 = jcshmsPart.JC044
		finalRec.JC061 = jcshmsPart.JC061
		finalRec.JC062 = jcshmsPart.JC062
		finalRec.JC063 = jcshmsPart.JC063
		finalRec.JC064 = jcshmsPart.JC064
		finalRec.JC065 = jcshmsPart.JC065
		finalRec.JC066 = jcshmsPart.JC066
	}

	// --- 2. jancodeテーブルをJA001で検索 ---
	const q2 = `SELECT JA006, JA007, JA008 FROM jancode WHERE JA001 = ? LIMIT 1`
	err2 := conn.QueryRow(q2, jan).Scan(&finalRec.JA006, &finalRec.JA007, &finalRec.JA008)
	if err2 != nil && err2 != sql.ErrNoRows {
		return nil, fmt.Errorf("jancode search failed: %w", err2)
	}
	if err2 == nil {
		found = true
	}

	// --- 3. どちらの検索でも見つからなければデータなしと判断 ---
	if !found {
		return nil, nil // Not found
	}

	return finalRec, nil
}
