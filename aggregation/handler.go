// File: aggregation/handler.go (最終修正版)
package aggregation

import (
	"database/sql"
	"encoding/json"
	"karashi/db"
	"karashi/model"
	"net/http"
	"strings"
)

// GetAggregationHandler returns the aggregated results.
func GetAggregationHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		filters := model.AggregationFilters{
			StartDate: q.Get("startDate"),
			EndDate:   q.Get("endDate"),
			KanaName:  q.Get("kanaName"),
			DrugTypes: strings.Split(q.Get("drugTypes"), ","),
			// ▼▼▼ 修正点: NoMovementパラメーターの読み込みを追加 ▼▼▼
			NoMovement: q.Get("noMovement") == "true",
		}

		results, err := db.GetAggregatedTransactions(conn, filters)
		if err != nil {
			http.Error(w, "Failed to get aggregated data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}
