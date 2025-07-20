package db

import (
	"database/sql"
	"fmt"
)

// NextSequence は code_sequences テーブルをトランザクション内でロックして
// last_no をインクリメントし、"MA2Y00000001" のような文字列を返します。
func NextSequence(conn *sql.DB, name string) (string, error) {
	tx, err := conn.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var last int
	err = tx.QueryRow(
		"SELECT last_no FROM code_sequences WHERE name = ?",
		name,
	).Scan(&last)
	if err != nil {
		return "", err
	}

	last++
	_, err = tx.Exec(
		"UPDATE code_sequences SET last_no = ? WHERE name = ?",
		last, name,
	)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	// フォーマットは "MA2Y"+8桁ゼロ埋め
	return fmt.Sprintf("%s%08d", name, last), nil
}

// ExistsMA2ByJan は ma_master テーブルに MA000=key のレコードがあるかを返します。
func ExistsMA2ByJan(conn *sql.DB, key string) (bool, error) {
	const q = `
SELECT 1
  FROM ma_master
 WHERE MA000 = ?
 LIMIT 1`
	var dummy int
	err := conn.QueryRow(q, key).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
