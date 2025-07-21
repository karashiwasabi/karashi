// File: usage/pausage.go
package usage

import (
	"encoding/csv"
	"fmt"
	"io"
	"karashi/model" // Import the new model package
	"strconv"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ParseUsage parses the USAGE CSV file.
func ParseUsage(r io.Reader) ([]model.ParsedUsage, error) {
	reader := csv.NewReader(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))
	reader.FieldsPerRecord = -1

	var out []model.ParsedUsage
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv read error: %w", err)
		}

		qty, _ := strconv.ParseFloat(rec[4], 64)

		out = append(out, model.ParsedUsage{
			Date:       rec[0],
			Yj:         rec[1],
			Jc:         rec[2],
			Pname:      rec[3],
			YjQty:      qty,
			YjUnitName: rec[5],
		})
	}
	return out, nil
}
