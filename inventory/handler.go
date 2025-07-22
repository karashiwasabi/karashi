// File: inventory/handler.go
package inventory

import (
	"database/sql"
	"encoding/json"
	"karashi/db"
	"karashi/model"
	"log"
	"net/http"
)

func UploadInventoryHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "File upload error", http.StatusBadRequest)
			return
		}
		defer file.Close()

		parsedRecords, err := ParseInventory(file)
		if err != nil {
			http.Error(w, "Inventory CSV parse error", http.StatusInternalServerError)
			log.Printf("inventory.ParseInventory error: %v", err)
			return
		}

		var finalRecords []model.ARInput
		for _, rec := range parsedRecords {
			processedRec, err := ExecuteInventoryBranching(conn, rec)
			if err != nil {
				log.Printf("ExecuteInventoryBranching failed for JAN %s: %v", rec.Ajc, err)
				continue
			}
			finalRecords = append(finalRecords, processedRec)
		}

		// Save the processed data to the inventory and a_records tables
		if err := db.PersistInventoryRecords(conn, finalRecords); err != nil {
			log.Printf("PersistInventoryRecords error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if err := db.PersistARecords(conn, finalRecords); err != nil {
			log.Printf("PersistARecords error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Processed and saved Inventory CSV successfully",
			"records": finalRecords,
		})
	}
}
