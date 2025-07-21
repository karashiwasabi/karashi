// File: db/ma_master.go
package db

import (
	"database/sql"
	"fmt"
	"karashi/model" // Import the new model package
)

// The MaMaster and MaMasterInput structs are now defined in model/types.go

// GetMaMasterByCode fetches a ma_master record by its primary key (MA000).
func GetMaMasterByCode(conn *sql.DB, code string) (*model.MaMaster, error) {
	const q = `SELECT MA000, MA009, MA018, MA022, MA030, MA037, MA039, MA044, MA061, MA062, MA063, MA064, MA065, MA066, MA131, MA132, MA133 FROM ma_master WHERE MA000 = ? LIMIT 1`
	var m model.MaMaster
	err := conn.QueryRow(q, code).Scan(
		&m.MA000, &m.MA009, &m.MA018, &m.MA022, &m.MA030, &m.MA037,
		&m.MA039, &m.MA044, &m.MA061, &m.MA062, &m.MA063, &m.MA064,
		&m.MA065, &m.MA066, &m.MA131, &m.MA132, &m.MA133,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// CreateMaMaster creates a new master record in the ma_master table.
func CreateMaMaster(conn *sql.DB, rec model.MaMasterInput) error {

	const q = `
INSERT INTO ma_master (
  MA000, MA009, MA018, MA022, MA030, MA037, MA039, MA044, 
  MA061, MA062, MA063, MA064, MA065, MA066,
  MA131, MA132, MA133
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := conn.Exec(q,
		rec.MA000, rec.MA009, rec.MA018, rec.MA022, rec.MA030, rec.MA037, rec.MA039, rec.MA044,
		rec.MA061, rec.MA062, rec.MA063, rec.MA064, rec.MA065, rec.MA066,
		rec.MA131, rec.MA132, rec.MA133,
	)
	if err != nil {
		return fmt.Errorf("CreateMaMaster failed: %w", err)
	}
	return nil
}
