// File: db/jcshms.go (修正版)
package db

import (
	"database/sql"
	"fmt"
	"karashi/model"
	"log" // ★ ログ出力のためにインポート
	"strconv"
	"strings"
)

// (GetJcshmsByJan関数は変更なし)
func GetJcshmsByJan(conn *sql.DB, jan string) (*model.JCShms, error) {
	// ...
	return nil, nil
}

func GetJcshmsByCodesMap(conn *sql.DB, jans []string) (map[string]*model.JCShms, error) {
	if len(jans) == 0 {
		return make(map[string]*model.JCShms), nil
	}

	results := make(map[string]*model.JCShms)

	inClause := `(?` + strings.Repeat(",?", len(jans)-1) + `)`
	args := make([]interface{}, len(jans))
	for i, jan := range jans {
		args[i] = jan
		results[jan] = &model.JCShms{}
	}

	q1 := `SELECT JC000, JC009, JC018, JC022, JC030, JC037, JC039, JC044, JC050,
	              JC061, JC062, JC063, JC064, JC065, JC066
	       FROM jcshms WHERE JC000 IN ` + inClause

	rows1, err := conn.Query(q1, args...)
	if err != nil {
		return nil, fmt.Errorf("jcshms bulk search failed: %w", err)
	}
	defer rows1.Close()

	for rows1.Next() {
		var jan string
		var jcshmsPart model.JCShms
		var jc050 sql.NullString // ★★★ JC050を安全に読み込むため、一時的に文字列型で受け取る

		if err := rows1.Scan(&jan, &jcshmsPart.JC009, &jcshmsPart.JC018, &jcshmsPart.JC022, &jcshmsPart.JC030,
			&jcshmsPart.JC037, &jcshmsPart.JC039, &jcshmsPart.JC044, &jc050, // ★★★ スキャン先をjc050に変更
			&jcshmsPart.JC061, &jcshmsPart.JC062, &jcshmsPart.JC063, &jcshmsPart.JC064, &jcshmsPart.JC065, &jcshmsPart.JC066,
		); err != nil {
			return nil, err
		}

		res := results[jan]
		res.JC009 = jcshmsPart.JC009
		res.JC018 = jcshmsPart.JC018
		res.JC022 = jcshmsPart.JC022
		res.JC030 = jcshmsPart.JC030
		res.JC037 = jcshmsPart.JC037
		res.JC039 = jcshmsPart.JC039
		res.JC044 = jcshmsPart.JC044
		res.JC061 = jcshmsPart.JC061
		res.JC062 = jcshmsPart.JC062
		res.JC063 = jcshmsPart.JC063
		res.JC064 = jcshmsPart.JC064
		res.JC065 = jcshmsPart.JC065
		res.JC066 = jcshmsPart.JC066

		// ★★★ ここからがJC050を安全に数値へ変換する処理 ★★★
		val, err := strconv.ParseFloat(jc050.String, 64)
		if err != nil {
			// 変換に失敗した場合(空文字など)は0を設定し、ログを出力
			res.JC050 = 0
			if jc050.String != "" { // 完全な空文字はログに出さない場合
				log.Printf("[WARN] JC050のデータが不正なため0に変換しました。製品名: %s, 元の値: '%s'", res.JC018, jc050.String)
			}
		} else {
			// 変換に成功した場合はその値を設定
			res.JC050 = val
		}
		// ★★★ ここまで ★★★
	}

	q2 := `SELECT JA001, JA006, JA007, JA008 FROM jancode WHERE JA001 IN ` + inClause

	rows2, err := conn.Query(q2, args...)
	if err != nil {
		return nil, fmt.Errorf("jancode bulk search failed: %w", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var jan string
		var jaPart struct {
			JA006 sql.NullFloat64
			JA007 sql.NullString
			JA008 sql.NullFloat64
		}
		if err := rows2.Scan(&jan, &jaPart.JA006, &jaPart.JA007, &jaPart.JA008); err != nil {
			return nil, err
		}
		results[jan].JA006 = jaPart.JA006
		results[jan].JA007 = jaPart.JA007
		results[jan].JA008 = jaPart.JA008
	}

	return results, nil
}
