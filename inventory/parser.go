// File: inventory/parser.go
package inventory

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// FileRowはファイルから読み込んだ行の構造体です
type FileRow struct {
	ProductName    string
	YjUnitName     string
	InnerPackQty   float64
	PhysicalJanQty float64
	YjCode         string
	JanCode        string
}

// ParsedInventoryFileはファイル全体の構造体です
type ParsedInventoryFile struct {
	Date string
	Rows []FileRow
}

// ParseInventoryFileは新しい形式の棚卸ファイルを解析します
func ParseInventoryFile(r io.Reader) (*ParsedInventoryFile, error) {
	reader := csv.NewReader(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))
	reader.FieldsPerRecord = -1 // 可変長カラムに対応

	var result ParsedInventoryFile
	var dataRows []FileRow

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv read all error: %w", err)
	}

	for _, row := range records {
		if len(row) == 0 {
			continue
		}

		rowType := strings.TrimSpace(row[0])
		switch rowType {
		case "H":
			if len(row) > 4 {
				result.Date = strings.TrimSpace(row[4])
			}
		case "R1":
			if len(row) > 45 {
				innerPackQty, _ := strconv.ParseFloat(strings.TrimSpace(row[17]), 64)
				physicalJanQty, _ := strconv.ParseFloat(strings.TrimSpace(row[21]), 64)

				dataRows = append(dataRows, FileRow{
					ProductName:    strings.TrimSpace(row[12]),
					YjUnitName:     strings.TrimSpace(row[16]),
					InnerPackQty:   innerPackQty,
					PhysicalJanQty: physicalJanQty,
					YjCode:         strings.TrimSpace(row[42]),
					JanCode:        strings.TrimSpace(row[45]),
				})
			}
		}
	}
	result.Rows = dataRows
	return &result, nil
}
