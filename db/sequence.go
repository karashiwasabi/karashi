// File: db/sequence.go (修正版)
package db

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
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

// InitializeSequenceFromMaxYjCodeは、product_masterテーブルのyj_codeの最大値から
// MA2Yシーケンスの初期値を設定します。DB復元後のYJコード重複を防ぎます。
func InitializeSequenceFromMaxYjCode(conn *sql.DB) error {
	rows, err := conn.Query("SELECT yj_code FROM product_master WHERE yj_code LIKE 'MA2Y%'")
	if err != nil {
		return fmt.Errorf("failed to query existing yj_codes: %w", err)
	}
	defer rows.Close()

	var maxNo int64 = 0
	prefix := "MA2Y"

	for rows.Next() {
		var yjCode string
		if err := rows.Scan(&yjCode); err != nil {
			// スキャンエラーはログに出力するが、処理は続行
			log.Printf("Warn: could not scan yj_code: %v", err)
			continue
		}

		if strings.HasPrefix(yjCode, prefix) {
			numPart := strings.TrimPrefix(yjCode, prefix)
			num, err := strconv.ParseInt(numPart, 10, 64)
			if err == nil {
				if num > maxNo {
					maxNo = num
				}
			}
		}
	}

	if maxNo > 0 {
		_, err = conn.Exec("UPDATE code_sequences SET last_no = ? WHERE name = ?", maxNo, "MA2Y")
		if err != nil {
			return fmt.Errorf("failed to update MA2Y sequence with max value %d: %w", maxNo, err)
		}
		log.Printf("MA2Y sequence initialized to %d based on existing product masters.", maxNo)
	}

	return nil
}

// ▼▼▼ ここから新しい関数を追加 ▼▼▼
// InitializeSequenceFromMaxClientCodeは、client_masterテーブルのclient_codeの最大値から
// CLシーケンスの初期値を設定します。DB復元後の得意先コード重複を防ぎます。
func InitializeSequenceFromMaxClientCode(conn *sql.DB) error {
	rows, err := conn.Query("SELECT client_code FROM client_master WHERE client_code LIKE 'CL%'")
	if err != nil {
		return fmt.Errorf("failed to query existing client_codes: %w", err)
	}
	defer rows.Close()

	var maxNo int64 = 0
	prefix := "CL"

	for rows.Next() {
		var clientCode string
		if err := rows.Scan(&clientCode); err != nil {
			log.Printf("Warn: could not scan client_code: %v", err)
			continue
		}

		if strings.HasPrefix(clientCode, prefix) {
			numPart := strings.TrimPrefix(clientCode, prefix)
			num, err := strconv.ParseInt(numPart, 10, 64)
			if err == nil {
				if num > maxNo {
					maxNo = num
				}
			}
		}
	}

	if maxNo > 0 {
		_, err = conn.Exec("UPDATE code_sequences SET last_no = ? WHERE name = ?", maxNo, "CL")
		if err != nil {
			return fmt.Errorf("failed to update CL sequence with max value %d: %w", maxNo, err)
		}
		log.Printf("CL sequence initialized to %d based on existing client masters.", maxNo)
	}

	return nil
}

// ▲▲▲ ここまで新しい関数を追加 ▲▲▲
