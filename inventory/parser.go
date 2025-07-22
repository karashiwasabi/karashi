// File: inventory/parser.go
package inventory

import (
	"encoding/csv"
	"fmt"
	"io"
	"karashi/model"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func ParseInventory(r io.Reader) ([]model.ARInput, error) {
	reader := csv.NewReader(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))
	reader.FieldsPerRecord = -1

	headerRow, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header row: %w", err)
	}
	date := ""
	if len(headerRow) > 4 {
		date = strings.Trim(headerRow[4], `"' `)
	}

	if _, err := reader.Read(); err != nil && err != io.EOF {
		// Skip second row
	}

	var records []model.ARInput
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if len(row) < 46 {
			continue
		}

		// 0-indexed columns
		janQtyStr := strings.Trim(row[21], `"' `)
		coeffStr := strings.Trim(row[17], `"' `)

		janQty, _ := strconv.ParseFloat(janQtyStr, 64)
		coeff, _ := strconv.ParseFloat(coeffStr, 64)

		yjQty := janQty * coeff // Calculate YJ Quantity

		rec := model.ARInput{
			Adate:      date,
			Aflag:      4, // vvv 「種別」に4をセット vvv
			Apname:     strings.Trim(row[12], `"' `),
			Ajanqty:    janQty,
			Ajpu:       coeff,
			Ajanunitnm: strings.Trim(row[23], `"' `),
			Ayjqty:     yjQty,
			Ayjunitnm:  strings.Trim(row[16], `"' `),
			Ayj:        strings.Trim(row[42], `"' `),
			Ajc:        strings.Trim(row[45], `"' `),
		}
		records = append(records, rec)
	}
	return records, nil
}
