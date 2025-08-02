// File: updatemaster/handler.go (修正版)
package updatemaster

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"karashi/db"
	"karashi/model"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// UpdatedProduct は更新された製品情報をフロントエンドに返すための構造体です
type UpdatedProduct struct {
	JanCode     string `json:"janCode"`
	ProductName string `json:"productName"`
}

// JCSHMSUpdateHandler は product_master を新しいJCSHMSデータで更新します
func JCSHMSUpdateHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 新しいJCSHMS/JANCODEをメモリに読み込む
		newJcshmsData, err := loadCSVToMap("SOU/JCSHMS.CSV", true)
		if err != nil {
			http.Error(w, "Failed to load new JCSHMS.CSV", http.StatusInternalServerError)
			return
		}
		newJancodeData, err := loadCSVToMap("SOU/JANCODE.CSV", true)
		if err != nil {
			http.Error(w, "Failed to load new JANCODE.CSV", http.StatusInternalServerError)
			return
		}

		// 2. DBから現在の product_master を全件取得
		productMasters, err := db.GetAllProductMasters(conn)
		if err != nil {
			http.Error(w, "Failed to get product masters", http.StatusInternalServerError)
			return
		}

		var updatedProducts []UpdatedProduct

		tx, err := conn.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// 3. product_masterをループし、新しいデータと照合・更新
		for _, master := range productMasters {
			janCode := master.ProductCode
			newJcshmsRow, jcshmsExists := newJcshmsData[janCode]

			if !jcshmsExists {
				continue // 新しいJCSHMSに存在しないものはスキップ
			}

			input, err := createInputFromCSV(newJcshmsRow, newJancodeData[janCode])
			if err != nil {
				log.Printf("[WARN] Skipping update for JAN %s due to data error: %v", janCode, err)
				continue
			}

			if err := db.UpsertProductMasterInTx(tx, input); err != nil {
				http.Error(w, fmt.Sprintf("Failed to update master for %s", janCode), http.StatusInternalServerError)
				return
			}

			updatedProducts = append(updatedProducts, UpdatedProduct{
				JanCode:     janCode,
				ProductName: input.ProductName,
			})
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":         fmt.Sprintf("%d件の製品マスターを更新しました。", len(updatedProducts)),
			"updatedProducts": updatedProducts,
		})
	}
}

// loadCSVToMapはCSVファイルを読み込み、最初の列をキーとしたマップを返します
func loadCSVToMap(filepath string, skipHeader bool) (map[string][]string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(transform.NewReader(f, japanese.ShiftJIS.NewDecoder()))
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	if skipHeader {
		if _, err := r.Read(); err != nil {
			return nil, err
		}
	}

	dataMap := make(map[string][]string)
	for {
		row, err := r.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		if len(row) > 0 {
			dataMap[row[0]] = row
		}
	}
	return dataMap, nil
}

// createInputFromCSVはCSVの行データからProductMasterInputを生成します
func createInputFromCSV(jcshmsRow, jancodeRow []string) (model.ProductMasterInput, error) {
	var input model.ProductMasterInput
	// CSVの列は0から始まる。JC000はインデックス0。
	if len(jcshmsRow) < 66 {
		return input, fmt.Errorf("invalid jcshms row length, expected at least 66 columns, got %d", len(jcshmsRow))
	}

	input.ProductCode = jcshmsRow[0]  // JC000 (JAN)
	input.YjCode = jcshmsRow[9]       // JC009
	input.ProductName = jcshmsRow[18] // JC018
	input.Origin = "JCSHMS"
	input.KanaName = jcshmsRow[22]                                 // JC022
	input.MakerName = jcshmsRow[30]                                // JC030
	input.PackageSpec = jcshmsRow[37]                              // JC037
	input.YjUnitName = jcshmsRow[39]                               // JC039
	input.YjPackUnitQty, _ = strconv.ParseFloat(jcshmsRow[44], 64) // JC044
	input.NhiPrice, _ = strconv.ParseFloat(jcshmsRow[50], 64)      // JC050
	input.FlagPoison, _ = strconv.Atoi(jcshmsRow[61])              // JC061
	input.FlagDeleterious, _ = strconv.Atoi(jcshmsRow[62])         // JC062
	input.FlagNarcotic, _ = strconv.Atoi(jcshmsRow[63])            // JC063
	input.FlagPsychotropic, _ = strconv.Atoi(jcshmsRow[64])        // JC064
	input.FlagStimulant, _ = strconv.Atoi(jcshmsRow[65])           // JC065
	input.FlagStimulantRaw, _ = strconv.Atoi(jcshmsRow[66])        // JC066

	if len(jancodeRow) >= 9 {
		input.JanPackInnerQty, _ = strconv.ParseFloat(jancodeRow[6], 64) // JA006
		input.JanUnitCode, _ = strconv.Atoi(jancodeRow[7])               // JA007
		input.JanPackUnitQty, _ = strconv.ParseFloat(jancodeRow[8], 64)  // JA008
	}

	return input, nil
}
