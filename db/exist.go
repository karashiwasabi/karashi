// File: db/exist.go
package db

import (
	"database/sql"
)

// ExistsByJan checks if ma_master.MA000 (JANコード) exists.
func ExistsByJan(conn *sql.DB, jan string) (bool, error) {
	const q = `
SELECT 1
  FROM ma_master
 WHERE MA000 = ?
 LIMIT 1`
	var dummy int
	err := conn.QueryRow(q, jan).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// ExistsByYj checks if ma_master.MA009 (YJコード) exists.
func ExistsByYj(conn *sql.DB, yj string) (bool, error) {
	const q = `
SELECT 1
  FROM ma_master
 WHERE MA009 = ?
 LIMIT 1`
	var dummy int
	err := conn.QueryRow(q, yj).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
