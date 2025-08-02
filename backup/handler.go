// File: backup/handler.go (Corrected)
package backup

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"karashi/db"
	"karashi/model"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ExportClientsHandler handles exporting the client master to a CSV file.
func ExportClientsHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clients, err := db.GetAllClients(conn)
		if err != nil {
			http.Error(w, "Failed to get clients", http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		sjisWriter := transform.NewWriter(&buf, japanese.ShiftJIS.NewEncoder())
		csvWriter := csv.NewWriter(sjisWriter)

		if err := csvWriter.Write([]string{"client_code", "client_name"}); err != nil {
			http.Error(w, "Failed to write CSV header", http.StatusInternalServerError)
			return
		}

		for _, client := range clients {
			if err := csvWriter.Write([]string{client.Code, client.Name}); err != nil {
				http.Error(w, "Failed to write CSV row", http.StatusInternalServerError)
				return
			}
		}

		csvWriter.Flush()

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", `attachment; filename="client_master.csv"`)
		w.Write(buf.Bytes())
	}
}

// ImportClientsHandler handles importing clients from a CSV file.
func ImportClientsHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "File upload error", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		sjisReader := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
		csvReader := csv.NewReader(sjisReader)
		records, err := csvReader.ReadAll()
		if err != nil {
			http.Error(w, "Failed to parse CSV file", http.StatusBadRequest)
			return
		}

		if len(records) < 1 {
			http.Error(w, "Empty CSV file", http.StatusBadRequest)
			return
		}

		tx, err := conn.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare("INSERT OR REPLACE INTO client_master (client_code, client_name) VALUES (?, ?)")
		if err != nil {
			http.Error(w, "Failed to prepare DB statement", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		var importedCount int
		for i := 1; i < len(records); i++ {
			row := records[i]
			if len(row) < 2 {
				continue
			}
			clientCode := row[0]
			clientName := row[1]
			if _, err := stmt.Exec(clientCode, clientName); err != nil {
				log.Printf("Failed to import row %d: %v", i+1, err)
				http.Error(w, fmt.Sprintf("Failed to import row %d", i+1), http.StatusInternalServerError)
				return
			}
			importedCount++
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("%d件の得意先をインポートしました。", importedCount),
		})
	}
}

// ExportProductsHandler handles exporting the product master to a CSV file.
func ExportProductsHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ★★★ CHANGED: Use GetEditableProductMasters to fetch only non-JCSHMS items ★★★
		products, err := db.GetEditableProductMasters(conn)
		if err != nil {
			http.Error(w, "Failed to get products", http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		sjisWriter := transform.NewWriter(&buf, japanese.ShiftJIS.NewEncoder())
		csvWriter := csv.NewWriter(sjisWriter)

		header := []string{
			"product_code", "yj_code", "product_name", "origin", "kana_name", "maker_name", "package_spec",
			"yj_unit_name", "yj_pack_unit_qty", "flag_poison", "flag_deleterious", "flag_narcotic",
			"flag_psychotropic", "flag_stimulant", "flag_stimulant_raw", "jan_pack_inner_qty",
			"jan_unit_code", "jan_pack_unit_qty", "reorder_point", "nhi_price",
		}
		if err := csvWriter.Write(header); err != nil {
			http.Error(w, "Failed to write CSV header", http.StatusInternalServerError)
			return
		}

		for _, p := range products {
			row := []string{
				p.ProductCode, p.YjCode, p.ProductName, p.Origin, p.KanaName, p.MakerName, p.PackageSpec,
				p.YjUnitName, fmt.Sprintf("%f", p.YjPackUnitQty), fmt.Sprintf("%d", p.FlagPoison),
				fmt.Sprintf("%d", p.FlagDeleterious), fmt.Sprintf("%d", p.FlagNarcotic),
				fmt.Sprintf("%d", p.FlagPsychotropic), fmt.Sprintf("%d", p.FlagStimulant),
				fmt.Sprintf("%d", p.FlagStimulantRaw), fmt.Sprintf("%f", p.JanPackInnerQty),
				fmt.Sprintf("%d", p.JanUnitCode), fmt.Sprintf("%f", p.JanPackUnitQty),
				fmt.Sprintf("%f", p.ReorderPoint), fmt.Sprintf("%f", p.NhiPrice),
			}
			if err := csvWriter.Write(row); err != nil {
				http.Error(w, "Failed to write CSV row", http.StatusInternalServerError)
				return
			}
		}
		csvWriter.Flush()

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", `attachment; filename="product_master.csv"`)
		w.Write(buf.Bytes())
	}
}

// ImportProductsHandler handles importing products from a CSV file.
func ImportProductsHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "File upload error", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		sjisReader := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
		csvReader := csv.NewReader(sjisReader)
		records, err := csvReader.ReadAll()
		if err != nil {
			http.Error(w, "Failed to parse CSV file", http.StatusBadRequest)
			return
		}

		if len(records) < 2 { // At least one header and one data row
			http.Error(w, "CSV file must have a header and at least one data row", http.StatusBadRequest)
			return
		}

		tx, err := conn.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var importedCount int
		for i := 1; i < len(records); i++ {
			row := records[i]
			if len(row) < 20 {
				continue
			}

			yjPackUnitQty, _ := strconv.ParseFloat(row[8], 64)
			flagPoison, _ := strconv.Atoi(row[9])
			flagDeleterious, _ := strconv.Atoi(row[10])
			flagNarcotic, _ := strconv.Atoi(row[11])
			flagPsychotropic, _ := strconv.Atoi(row[12])
			flagStimulant, _ := strconv.Atoi(row[13])
			flagStimulantRaw, _ := strconv.Atoi(row[14])
			janPackInnerQty, _ := strconv.ParseFloat(row[15], 64)
			janUnitCode, _ := strconv.Atoi(row[16])
			janPackUnitQty, _ := strconv.ParseFloat(row[17], 64)
			reorderPoint, _ := strconv.ParseFloat(row[18], 64)
			nhiPrice, _ := strconv.ParseFloat(row[19], 64)

			input := model.ProductMasterInput{
				ProductCode: row[0], YjCode: row[1], ProductName: row[2], Origin: row[3], KanaName: row[4],
				MakerName: row[5], PackageSpec: row[6], YjUnitName: row[7], YjPackUnitQty: yjPackUnitQty,
				FlagPoison: flagPoison, FlagDeleterious: flagDeleterious, FlagNarcotic: flagNarcotic,
				FlagPsychotropic: flagPsychotropic, FlagStimulant: flagStimulant, FlagStimulantRaw: flagStimulantRaw,
				JanPackInnerQty: janPackInnerQty, JanUnitCode: janUnitCode, JanPackUnitQty: janPackUnitQty,
				ReorderPoint: reorderPoint, NhiPrice: nhiPrice,
			}

			if err := db.UpsertProductMasterInTx(tx, input); err != nil {
				log.Printf("Failed to import product row %d: %v", i+1, err)
				http.Error(w, fmt.Sprintf("Failed to import product row %d", i+1), http.StatusInternalServerError)
				return
			}
			importedCount++
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("%d件の製品をインポートしました。", importedCount),
		})
	}
}
