// File: db/sequence.go (修正版)
package db

import (
	"database/sql"
	"fmt"
)

// NextSequenceInTxは、指定されたシーケンスから次の番号を生成します。
// 例: (tx, "CL", "CL", 4) -> "CL0001"
func NextSequenceInTx(tx *sql.Tx, name, prefix string, padding int) (string, error) {
	var lastNo int

	err := tx.QueryRow("SELECT last_no FROM code_sequences WHERE name = ?", name).Scan(&lastNo)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("シーケンス '%s' が見つかりません", name)
		}
		return "", fmt.Errorf("シーケンス '%s' の取得に失敗しました: %w", name, err)
	}

	newNo := lastNo + 1

	_, err = tx.Exec("UPDATE code_sequences SET last_no = ? WHERE name = ?", newNo, name)
	if err != nil {
		return "", fmt.Errorf("シーケンス '%s' の更新に失敗しました: %w", name, err)
	}

	format := fmt.Sprintf("%s%%0%dd", prefix, padding)
	return fmt.Sprintf(format, newNo), nil
}
