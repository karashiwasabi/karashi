package usage

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"

	"karashi/model"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ParseUsage は Shift_JIS の Usage CSV を読み込み、
// model.ParsedUsage のスライスを返します。
// フォーマット: Date,Jc,Yj,Pname,Yjqty,Yjunitname （ヘッダーなし）
func ParseUsage(r io.Reader) ([]model.ParsedUsage, error) {
	// Shift_JIS → UTF-8 変換
	decoder := japanese.ShiftJIS.NewDecoder()
	reader := csv.NewReader(transform.NewReader(r, decoder))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	var out []model.ParsedUsage
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("ParseUsage: read error: %v", err)
			continue
		}
		// 必要フィールド数を満たすかチェック
		if len(rec) < 6 {
			log.Printf("ParseUsage: unexpected field count: got %d, want ≥6", len(rec))
			continue
		}

		// YJ数量を float64 に変換
		qty, err := strconv.ParseFloat(rec[4], 64)
		if err != nil {
			log.Printf("ParseUsage: parse YjQty error for %q: %v", rec[4], err)
			qty = 0
		}

		u := model.ParsedUsage{
			Date:       rec[0],
			Jc:         rec[2],
			Yj:         rec[1],
			Pname:      rec[3],
			YjQty:      qty,
			YjUnitName: rec[5],
		}
		out = append(out, u)
	}

	return out, nil
}
