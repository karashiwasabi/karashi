// File: usage/parser.go
package usage

import (
	"encoding/csv"
	"fmt"
	"io"
	"karashi/model"
	"strconv"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ParseUsageはUSAGE CSVを解析し、UnifiedInputRecordのスライスを返します。
func ParseUsage(r io.Reader) ([]model.UnifiedInputRecord, error) {
	reader := csv.NewReader(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))
	reader.FieldsPerRecord = -1

	var records []model.UnifiedInputRecord
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv read error: %w", err)
		}

		yjQty, _ := strconv.ParseFloat(rec[4], 64)

		unifiedRec := model.UnifiedInputRecord{
			Date:        rec[0],
			YjCode:      rec[1],
			JanCode:     rec[2],
			ProductName: rec[3],
			YjQuantity:  yjQty,
			YjUnitName:  rec[5],
		}
		records = append(records, unifiedRec)
	}
	return records, nil
}
