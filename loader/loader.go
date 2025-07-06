// File: loader/loader.go
package loader

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func InitDatabase(db *sql.DB) error {
	if err := applySchema(db); err != nil {
		return fmt.Errorf("schema.sql 読み込み失敗: %w", err)
	}
	if err := loadCSV(db, "SOU/JCSHMS.CSV", "jcshms", 125, false); err != nil {
		return fmt.Errorf("JCSHMS 読み込み失敗: %w", err)
	}
	if err := loadCSV(db, "SOU/JANCODE.CSV", "jancode", 30, true); err != nil {
		return fmt.Errorf("JANCODE 読み込み失敗: %w", err)
	}
	return nil
}

func applySchema(db *sql.DB) error {
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(schema))
	return err
}

func loadCSV(db *sql.DB, filepath, tablename string, columns int, skipHeader bool) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(transform.NewReader(f, japanese.ShiftJIS.NewDecoder()))
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	if skipHeader {
		_, _ = r.Read()
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ph := make([]string, columns)
	for i := range ph {
		ph[i] = "?"
	}
	stmt, err := tx.Prepare(fmt.Sprintf(
		"INSERT OR REPLACE INTO %s VALUES (%s)",
		tablename, strings.Join(ph, ","),
	))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) != columns {
			continue
		}
		args := make([]interface{}, columns)
		for i := range args {
			args[i] = row[i]
		}
		if _, err := stmt.Exec(args...); err != nil {
			continue
		}
	}
	return tx.Commit()
}
