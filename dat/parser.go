// File: dat/parser.go
package dat

import (
	"bufio"
	"io"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ParseDatは、固定長のDATファイルからレコードを抽出します。
func ParseDat(r io.Reader) ([]ParsedDat, error) {
	scanner := bufio.NewScanner(r)
	var records []ParsedDat
	var currentWholesale string

	for scanner.Scan() {
		line := scanner.Text() // Shift_JISのままの行データ
		if len(line) == 0 {
			continue
		}

		switch line[0:1] {
		case "S":
			if len(line) >= 13 {
				currentWholesale = strings.TrimSpace(line[2:13])
			}
		case "D":
			if len(line) < 121 {
				line += strings.Repeat(" ", 121-len(line))
			}

			// 製品名フィールドをShift_JISのまま切り出し、UTF-8にデコード
			productNameSJIS := line[38:78]
			utf8Bytes, _, _ := transform.Bytes(japanese.ShiftJIS.NewDecoder(), []byte(productNameSJIS))
			productNameUTF8 := strings.TrimSpace(string(utf8Bytes))

			rec := ParsedDat{
				WholesaleCode: currentWholesale,
				// vvv ここからが修正箇所 vvv
				DeliveryFlag: strings.TrimSpace(line[3:4]),  // 4文字目をフラグとして取得
				DatDate:      strings.TrimSpace(line[4:12]), // 5文字目からを日付として取得
				// ^^^ ここまで ^^^
				ReceiptNumber: strings.TrimSpace(line[12:22]),
				LineNumber:    strings.TrimSpace(line[22:24]),
				JanCode:       strings.TrimSpace(line[25:38]),
				ProductName:   productNameUTF8,
				Quantity:      strings.TrimSpace(line[78:83]),
				UnitPrice:     strings.TrimSpace(line[83:92]),
				Subtotal:      strings.TrimSpace(line[92:101]),
				ExpiryDate:    strings.TrimSpace(line[109:115]),
				LotNumber:     strings.TrimSpace(line[115:121]),
			}
			records = append(records, rec)
		}
	}
	return records, scanner.Err()
}
