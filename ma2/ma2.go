package ma2

import (
	"database/sql"
	"fmt"

	"karashi/jancode"
	"karashi/jcshms"
)

// Record は ma2 の軽量版レコードです。
type Record struct {
	ID    int64  // AUTOINCREMENT
	MA000 string // JANコード
	MA009 string // YJコード
	MA018 string // 商品名
	MA037 string // 包装形態（未使用）
	MA039 string // 包装単位（未使用）
	MA044 string // 包装総量（未使用）
	MA131 string // JAN数量
	MA132 string // JAN単位コード
	MA133 string // YJ換算数量
	MA134 string // YJ単位コード
}

// QueryByJan は ma2 を JAN で検索します。
func QueryByJan(db *sql.DB, jan string) (*Record, error) {
	const sqlstr = `
        SELECT
          id,
          MA000, MA009, MA018,
          MA037, MA039, MA044,
          MA131, MA132, MA133, MA134
        FROM ma2
        WHERE MA000 = ?`
	var r Record
	err := db.QueryRow(sqlstr, jan).Scan(
		&r.ID,
		&r.MA000, &r.MA009, &r.MA018,
		&r.MA037, &r.MA039, &r.MA044,
		&r.MA131, &r.MA132, &r.MA133, &r.MA134,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ma2.QueryByJan failed: %w", err)
	}
	return &r, nil
}

// InsertWithoutYJ は YJ以外のフィールドで INSERT。
// トリガーで MA009 が自動採番されます。
func InsertWithoutYJ(db *sql.DB, r *Record) error {
	const stmt = `
        INSERT INTO ma2 (
          MA000, MA018,
          MA037, MA039, MA044,
          MA131, MA132, MA133, MA134
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := db.Exec(stmt,
		r.MA000, r.MA018,
		r.MA037, r.MA039, r.MA044,
		r.MA131, r.MA132, r.MA133, r.MA134,
	); err != nil {
		return fmt.Errorf("ma2.InsertWithoutYJ failed: %w", err)
	}
	return nil
}

// CheckOrCreateByJan は既存レコードがなければ作成します。
// datName は DAT から渡された品名です。
func CheckOrCreateByJan(db *sql.DB, jan, datName string) (*Record, error) {
	// 1) 既存チェック
	if rec, err := QueryByJan(db, jan); err != nil {
		return nil, err
	} else if rec != nil {
		return rec, nil
	}

	// 2) マスタ参照
	jc, _ := jcshms.QueryByJan(db, jan)
	ja, _ := jancode.QueryByJan(db, jan)

	// 3) レコード構築
	rec := &Record{
		MA000: jan,
		MA018: datName,
	}
	if jc != nil {
		rec.MA018 = jc.JC018
	}
	if ja != nil {
		rec.MA131 = ja.JA006
		rec.MA132 = ja.JA007
		rec.MA133 = ja.JA008
		rec.MA134 = ja.JA009
	}

	// 4) INSERT → トリガーで MA009 自動採番
	if err := InsertWithoutYJ(db, rec); err != nil {
		return nil, err
	}

	// 5) 採番後再取得
	return QueryByJan(db, jan)
}
