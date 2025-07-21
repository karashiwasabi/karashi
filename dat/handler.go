// File: dat/handler.go
package dat

import (
	"database/sql"
	"encoding/json"
	"karashi/db"
	"karashi/model"
	"log"
	"net/http"
	"strconv"
)

func UploadDatHandler(conn *sql.DB) http.HandlerFunc {
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

		parsedRecords, err := ParseDat(file)
		if err != nil {
			http.Error(w, "DAT parse error", http.StatusInternalServerError)
			log.Printf("dat.ParseDat error: %v", err)
			return
		}

		var finalRecords []model.ARInput
		for _, rec := range parsedRecords {
			// 1. Convert ParsedDat to the common ARInput struct
			datqty, _ := strconv.ParseFloat(rec.Quantity, 64)
			unitprice, _ := strconv.ParseFloat(rec.UnitPrice, 64)
			subtotal, _ := strconv.ParseFloat(rec.Subtotal, 64)
			expdate, _ := strconv.ParseFloat(rec.ExpiryDate, 64)
			flag, _ := strconv.Atoi(rec.DeliveryFlag)

			ar := model.ARInput{
				Adate:      rec.DatDate,
				Apcode:     rec.WholesaleCode,
				Arpnum:     rec.ReceiptNumber,
				Alnum:      rec.LineNumber,
				Aflag:      flag,
				Ajc:        rec.JanCode,
				Apname:     rec.ProductName,
				Adatqty:    datqty,
				Aunitprice: unitprice,
				Asubtotal:  subtotal,
				Aexpdate:   expdate,
				Alot:       rec.LotNumber,
			}

			// 2. Call the branching logic
			processedAr, err := ExecuteDatBranching(conn, ar)
			if err != nil {
				log.Printf("ExecuteDatBranching failed for JAN %s: %v", ar.Ajc, err)
				continue
			}
			finalRecords = append(finalRecords, processedAr)
		}

		// 3. Persist the final records to the a_records table
		if len(finalRecords) > 0 {
			if err := db.PersistARecords(conn, finalRecords); err != nil {
				log.Printf("PersistARecords error: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		// Return the final, processed data as JSON to the frontend
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Parsed and processed DAT file successfully",
			"records": finalRecords,
		})
	}
}
