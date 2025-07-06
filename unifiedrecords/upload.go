// File: unifiedrecords/upload.go
package unifiedrecords

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// DB は main.go から渡される共有DB接続
var DB *sql.DB

// UploadDatHandler は POST でアップされた DAT ファイル群を受け取り、
// 解析・登録後に JSON を返却します。
func UploadDatHandler(w http.ResponseWriter, r *http.Request) {
	// 最大 10MB の multipart/form-data を解析
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "フォーム解釈失敗", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["datFileInput[]"]
	all := make([]Record, 0)

	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			log.Println("DATオープン失敗:", err)
			continue
		}

		records, err := ParseDATFile(f, DB)
		f.Close()
		if err != nil {
			log.Println("パース失敗:", err)
			continue
		}

		for _, rec := range records {
			if err := Insert(DB, rec); err != nil {
				log.Println("登録失敗:", err)
				continue
			}
			all = append(all, rec)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":   len(all),
		"records": all,
	})
}
