// File: usage/upusage.go
package usage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"karashi/db"
	"karashi/model"
	"log"
	"net/http"
)

func UploadUsageHandler(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// --- 1. ファイルパース ---
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, fmt.Sprintf("parse form error: %v", err), http.StatusBadRequest)
			return
		}
		defer r.MultipartForm.RemoveAll()

		var allParsed []model.ParsedUsage
		files := r.MultipartForm.File["file"]
		if len(files) == 0 {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(map[string]interface{}{"records": []model.ARInput{}})
			return
		}

		for _, fh := range files {
			f, err := fh.Open()
			if err != nil {
				log.Printf("open file error: %v", err)
				continue
			}
			defer f.Close()

			recs, err := ParseUsage(f)
			if err != nil {
				log.Printf("parse error: %v", err)
				continue
			}
			allParsed = append(allParsed, recs...)
		}

		// --- 2. 前処理 (重複フィルタ) ---
		filtered := RemoveDuplicates(allParsed)

		// --- 3. 分岐処理 ---
		var finalARs []model.ARInput
		for _, pu := range filtered {
			// ★★★ ここが修正点 ★★★
			// 不要になった行番号の引数 'i' を削除
			ar, err := ExecuteBranching(conn, pu)
			if err != nil {
				log.Printf("ExecuteBranching failed for JAN %q: %v", pu.Jc, err)
				continue
			}
			finalARs = append(finalARs, ar)
		}

		// --- 4. データベースへの一括保存 ---
		if len(finalARs) > 0 {
			if err := db.PersistARecords(conn, finalARs); err != nil {
				log.Printf("PersistARecords error: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		// --- 5. JSONレスポンス ---
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		response := map[string]interface{}{
			"records": finalARs,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("json encode error: %v", err)
		}
	}
}

// RemoveDuplicatesは重複排除のみを行います。
func RemoveDuplicates(rs []model.ParsedUsage) []model.ParsedUsage {
	seen := make(map[string]struct{}, len(rs))
	var out []model.ParsedUsage
	for _, r := range rs {
		key := fmt.Sprintf("%s|%s|%s|%s", r.Date, r.Jc, r.Yj, r.Pname)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}
	return out
}
