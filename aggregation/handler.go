// File: aggregation/handler.go (Corrected)
package aggregation

import (
	"database/sql"
	"encoding/json"
	"karashi/db"
	"karashi/model"
	"net/http"
	"strconv"
	"strings"
)

func GetAggregationHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		coefficient, err := strconv.ParseFloat(q.Get("coefficient"), 64)
		if err != nil {
			coefficient = 1.5 // デフォルト値
		}

		filters := model.AggregationFilters{
			StartDate:   q.Get("startDate"),
			EndDate:     q.Get("endDate"),
			KanaName:    q.Get("kanaName"),
			DrugTypes:   strings.Split(q.Get("drugTypes"), ","),
			NoMovement:  q.Get("noMovement") == "true",
			Coefficient: coefficient,
		}

		results, err := db.GetStockLedger(conn, filters)
		if err != nil {
			http.Error(w, "Failed to get aggregated data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}
