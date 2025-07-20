package usage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// UploadUsageHandler は Usage CSV を受け取り、パース→重複・期間フィルタ→分岐判定→
// Branch-2 の MA2・DA 登録→分岐結果を JSON で返却するハンドラです。
//
//	db   : *sql.DB
//	from : 期間フィルタ開始日 (YYYYMMDD)
//	to   : 期間フィルタ終了日 (YYYYMMDD)
func UploadUsageHandler(db *sql.DB, from, to string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.MultipartForm.RemoveAll()

		// 1. CSV → []ParsedUsage
		var all []ParsedUsage
		for _, fh := range r.MultipartForm.File["file"] {
			f, err := fh.Open()
			if err != nil {
				log.Printf("UploadUsageHandler: open file error: %v", err)
				continue
			}
			recs, err := ParseUsage(f)
			f.Close()
			if err != nil {
				log.Printf("UploadUsageHandler: parse error: %v", err)
				continue
			}
			all = append(all, recs...)
		}

		// 2. 重複＆期間フィルタ
		filtered := RemoveDupAndFilterByPeriod(all, from, to)

		// 3. 分岐判定
		branchResults := branchUsage(filtered)

		// 4. Branch-2 のみ抽出 → MA2 生成 & DA 登録
		var bs2 []ParsedUsage
		for _, br := range branchResults {
			if br.Ama == "2" {
				bs2 = append(bs2, br.Parsed)
			}
		}
		arInputs, err := HandleBranch2MA(db, bs2)
		if err != nil {
			log.Printf("UploadUsageHandler: HandleBranch2MA error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if err := PersistBranch2DA(db, arInputs); err != nil {
			log.Printf("UploadUsageHandler: PersistBranch2DA error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// 5. JSON レスポンス
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(branchResults); err != nil {
			log.Printf("UploadUsageHandler: json encode error: %v", err)
		}
	}
}

// RemoveDupAndFilterByPeriod は重複レコードを排除し、
// Date が from～to 範囲外のものを除外します。
func RemoveDupAndFilterByPeriod(
	rs []ParsedUsage, from, to string,
) []ParsedUsage {
	seen := make(map[string]struct{}, len(rs))
	var out []ParsedUsage
	for _, r := range rs {
		if r.Date < from || r.Date > to {
			continue
		}
		key := fmt.Sprintf("%s|%s|%s|%s", r.Date, r.Jc, r.Yj, r.Pname)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}
	return out
}
