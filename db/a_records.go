// File: db/a_records.go
package db

import (
	"database/sql"
	"fmt"
	"karashi/model"
)

// PersistARecordsは、処理済みのレコードリストをDBに保存します。
func PersistARecords(conn *sql.DB, ars []model.ARInput) error {
	// 複数のレコードを効率的かつ安全に登録するため、トランザクションを開始します。
	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// 途中でエラーが発生した場合は、すべての変更を元に戻します。
	defer tx.Rollback()

	// a_recordsテーブルの全32列に対応したINSERT文を準備します。
	// INSERT OR REPLACE を使うことで、主キーが重複した場合でも新しいデータで上書きします。
	const q = `
INSERT OR REPLACE INTO a_records (
  adate, apcode, arpnum, alnum, aflag, ajc, ayj, apname, akana, apkg, amaker,
  adatqty, ajanqty, ajpu, ajanunitname, ajanunitcode, ayjqty, ayjpu, ayjunitname,
  aunitprice, asubtotal, ataxamount, ataxrate, aexpdate, alot,
  adokuyaku, agekiyaku, amayaku, akouseisinyaku, akakuseizai, akakuseizaigenryou,
  ama
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(q)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// 受け取ったレコードのリストをループ処理で1件ずつデータベースに登録します。
	for _, ar := range ars {
		_, err := stmt.Exec(
			ar.Adate, ar.Apcode, ar.Arpnum, ar.Alnum, ar.Aflag, ar.Ajc, ar.Ayj, ar.Apname, ar.Akana, ar.Apkg, ar.Amaker,
			ar.Adatqty, ar.Ajanqty, ar.Ajpu, ar.Ajanunitnm, ar.Ajanunitcode, ar.Ayjqty, ar.Ayjpu, ar.Ayjunitnm,
			ar.Aunitprice, ar.Asubtotal, ar.Ataxamt, ar.Ataxrate, ar.Aexpdate, ar.Alot,
			ar.Adokuyaku, ar.Agekiyaku, ar.Amayaku, ar.Akouseisinyaku, ar.Akakuseizai, ar.Akakuseizaigenryou,
			ar.Ama,
		)
		if err != nil {
			// 1件でもエラーがあれば、全処理を中止してエラーを返します。
			return fmt.Errorf("failed to exec statement for ARInput (JAN: %s, Ama: %s): %w", ar.Ajc, ar.Ama, err)
		}
	}

	// すべての登録が成功した場合、トランザクションを確定します。
	return tx.Commit()
}
