// File: unifiedrecords/datparser.go
package unifiedrecords

import (
	"bufio"
	"bytes"
	"database/sql"
	"io"
	"log"
	"strconv"
	"strings"

	"karashi/ma0"
	"karashi/ma2"
	"karashi/tani"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ParseDATFile は DAT フォーマットを読み込み、
// MA0 → MA2 の順でマスタ補完しつつ Record を返します。
func ParseDATFile(r io.Reader, db *sql.DB) ([]Record, error) {
	var list []Record
	var partnerCode string
	reader := bufio.NewReader(r)

	for {
		rawLine, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		line := bytes.TrimRight(rawLine, "\r\n")
		if len(line) < 3 {
			continue
		}

		switch string(line[:3]) {
		case "S20":
			if len(line) >= 12 {
				partnerCode = strings.TrimSpace(string(line[3:12]))
			}

		case "D20":
			if len(line) < 121 {
				continue
			}
			field := func(start, end int) []byte {
				if len(line) >= end {
					return line[start:end]
				}
				return line[start:]
			}

			rawName := field(38, 78)
			name, _, err := transform.String(
				japanese.ShiftJIS.NewDecoder(),
				string(rawName),
			)
			if err != nil {
				name = string(rawName)
			}

			rec := Record{
				SlipDate:       strings.TrimSpace(string(field(4, 12))),
				Flag:           atoi(string(field(3, 4))),
				ReceiptNumber:  strings.TrimSpace(string(field(12, 22))),
				LineNumber:     strings.TrimSpace(string(field(22, 24))),
				JANCode:        strings.TrimSpace(string(field(25, 38))),
				ProductName:    strings.TrimSpace(name),
				DATQty:         atoi(string(field(78, 83))),
				UnitPrice:      atof(string(field(83, 92))),
				SubtotalAmount: atof(string(field(92, 101))),
				ExpiryDate:     strings.TrimSpace(string(field(109, 115))),
				LotNumber:      strings.TrimSpace(string(field(115, 121))),
				PartnerCode:    partnerCode,
			}

			if m0, err := ma0.CheckOrCreateMA0(db, rec.JANCode); err != nil {
				log.Println("ma0 lookup failed:", err)
			} else if m0 != nil {
				applyMaster0(&rec, m0)
			} else {
				if m2, err := ma2.CheckOrCreateByJan(db, rec.JANCode, rec.ProductName); err != nil {
					log.Println("ma2 lookup failed:", err)
				} else if m2 != nil {
					applyMaster2(&rec, m2)
				}
			}

			list = append(list, rec)
		}
	}
	return list, nil
}

func applyMaster0(rec *Record, m0 *ma0.MA0Full) {
	rec.YJCode = m0.MA009
	rec.ProductName = m0.MA018

	rec.JANQuantity = atoi(m0.MA131)
	rec.JANUnitCode = m0.MA132

	rec.YJQuantity = atof(m0.MA044)

	// YJ単位名称を必ず設定
	rec.YJUnitName = tani.ResolveName(m0.MA039)

	// JAN単位名称は常にコード→名称解決
	rec.JANUnitName = tani.ResolveName(rec.JANUnitCode)

	// JANUnitCode が空 or "0" の場合だけ名称をフォールバック
	if rec.JANUnitCode == "" || rec.JANUnitCode == "0" {
		rec.JANUnitName = rec.YJUnitName
	}

	rec.Packaging = buildPackagingString(m0)
}

func applyMaster2(rec *Record, m2 *ma2.Record) {
	rec.YJCode = m2.MA009
	rec.ProductName = m2.MA018

	rec.JANQuantity = atoi(m2.MA131)
	rec.JANUnitCode = m2.MA132

	rec.YJQuantity = atof(m2.MA044)

	// YJ単位名称を必ず設定
	rec.YJUnitName = tani.ResolveName(m2.MA039)

	// JAN単位名称は常にコード→名称解決
	rec.JANUnitName = tani.ResolveName(rec.JANUnitCode)

	// JANUnitCode が空 or "0" の場合だけ名称をフォールバック
	if rec.JANUnitCode == "" || rec.JANUnitCode == "0" {
		rec.JANUnitName = rec.YJUnitName
	}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

// buildPackagingString は MA0 の包装総量（MA044）＋包装単位（MA039）を組み立てます。
func buildPackagingString(m *ma0.MA0Full) string {
	qty := strings.TrimSpace(m.MA044)
	unit := tani.ResolveName(m.MA039)
	if qty == "" || unit == "" {
		return ""
	}
	return qty + unit
}
