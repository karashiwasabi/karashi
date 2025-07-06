// File: ma0/quick.go
package ma0

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"karashi/jancode"
	"karashi/jcshms"
)

// CheckOrCreateMA0 は MA0 テーブルを検索し、
// 見つかればそのまま返却。なければ JC+JA マスタ情報を
// MA0Full にフルマッピングして INSERT → 返却します。
func CheckOrCreateMA0(db *sql.DB, jan string) (*MA0Full, error) {
	// 空 JAN はスキップ
	if jan == "" || strings.Trim(jan, "0") == "" {
		return nil, nil
	}

	// 1) 既存 MA0 検索 (SELECT *)
	const sel = "SELECT * FROM ma0 WHERE MA000 = ?"
	var full MA0Full
	v := reflect.ValueOf(&full).Elem()
	scanArgs := make([]interface{}, v.NumField())
	for i := range scanArgs {
		scanArgs[i] = v.Field(i).Addr().Interface()
	}
	err := db.QueryRow(sel, jan).Scan(scanArgs...)
	if err == nil {
		// 既に登録済み
		return &full, nil
	}
	if err != sql.ErrNoRows {
		// DB エラー
		return nil, fmt.Errorf("ma0 lookup failed: %w", err)
	}

	// 2) JC/JA マスタ取得
	jc, err := jcshms.QueryByJan(db, jan)
	if err != nil {
		return nil, fmt.Errorf("jcshms lookup failed: %w", err)
	}
	ja, err := jancode.QueryByJan(db, jan)
	if err != nil {
		return nil, fmt.Errorf("jancode lookup failed: %w", err)
	}

	// 3) マッピング (どちらも存在し、YJ が取れる場合のみ登録)
	if jc != nil && jc.JC009 != "" && ja != nil && ja.JA009 != "" {
		// JC側フィールド
		full.MA000 = jc.JC000
		full.MA009 = jc.JC009
		full.MA018 = jc.JC018
		full.MA022 = jc.JC022

		full.MA037 = jc.JC037
		full.MA039 = jc.JC039
		full.MA044 = jc.JC044

		full.MA061 = jc.JC061
		full.MA062 = jc.JC062
		full.MA063 = jc.JC063
		full.MA064 = jc.JC064

		// JA側フィールド
		full.MA125 = ja.JA000
		full.MA131 = ja.JA006
		full.MA132 = ja.JA007
		full.MA133 = ja.JA008
		full.MA134 = ja.JA009

		// 4) INSERT OR IGNORE INTO ma0
		ph := make([]string, v.NumField())
		for i := range ph {
			ph[i] = "?"
		}
		insertSQL := fmt.Sprintf("INSERT OR IGNORE INTO ma0 VALUES(%s)", strings.Join(ph, ","))
		args := make([]interface{}, v.NumField())
		for i := range args {
			args[i] = v.Field(i).Interface()
		}
		if _, err := db.Exec(insertSQL, args...); err != nil {
			return nil, fmt.Errorf("ma0 insert failed: %w", err)
		}
		return &full, nil
	}

	// 5) マスタ情報不足 → 登録せず nil
	return nil, nil
}
