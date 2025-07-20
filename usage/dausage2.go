// File: usage/dausage2.go
package usage

import (
	"database/sql"
	"fmt"
)

// PersistBranch2DA は Branch-2 用に生成した ARInput を
// a_records テーブルに一括登録します。
func PersistBranch2DA(db *sql.DB, ars []ARInput) error {
	const q = `
INSERT INTO a_records (
  adate, apcode, arpnum, alnum, aflag, ajc, ayj, apname
  -- 必要に応じて他のカラムを追加
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`

	for _, ar := range ars {
		if _, err := db.Exec(
			q,
			ar.Adate,
			ar.Apcode,
			ar.Arpnum,
			ar.Alnum,
			ar.Aflag,
			ar.Ajc,
			ar.Ayj,
			ar.Apname,
		); err != nil {
			return fmt.Errorf("PersistBranch2DA failed (Apcode=%s): %w", ar.Apcode, err)
		}
	}
	return nil
}
