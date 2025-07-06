// File: jancode/package.go
package jancode

import (
	"database/sql"
	"fmt"
)

// JARecord は jancode テーブルの 1 行を表します。
// 必要なフィールドのみ宣言してください。
type JARecord struct {
	JA000 string // 管理コード
	JA006 string // JAN数量
	JA007 string // JAN単位コード
	JA008 string // YJ換算数量
	JA009 string // YJ換算単位コード
}

// QueryByJan は jancode テーブルから key=jan のレコードを返します。
func QueryByJan(db *sql.DB, jan string) (*JARecord, error) {
	const sqlstr = `
        SELECT
            JA000,   -- 管理コード
            JA006,   -- JAN数量
            JA007,   -- JAN単位コード
            JA008,   -- YJ換算数量
            JA009    -- YJ換算単位コード
        FROM jancode
        WHERE JA001 = ?  -- 主キー JA001（JANコード）で検索
    `
	var r JARecord
	err := db.QueryRow(sqlstr, jan).Scan(
		&r.JA000,
		&r.JA006,
		&r.JA007,
		&r.JA008,
		&r.JA009,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("jancode lookup failed: %w", err)
	}
	return &r, nil
}
