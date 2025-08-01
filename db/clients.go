// File: db/clients.go (新規作成)
package db

import (
	"database/sql"
	"fmt"
	"karashi/model"
)

// GetAllClientsは、得意先マスターから全件取得します。
func GetAllClients(conn *sql.DB) ([]model.Client, error) {
	rows, err := conn.Query("SELECT client_code, client_name FROM client_master ORDER BY client_code")
	if err != nil {
		return nil, fmt.Errorf("failed to get all clients: %w", err)
	}
	defer rows.Close()

	var clients []model.Client
	for rows.Next() {
		var c model.Client
		if err := rows.Scan(&c.Code, &c.Name); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, nil
}
