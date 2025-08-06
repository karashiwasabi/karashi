// File: db/search.go (修正版)
package db

import (
	"database/sql"
	"fmt"
	"karashi/model"
	"karashi/units"
	"log" // ★ ログ出力のためにインポート
	"strconv"
)

func SearchJcshmsByName(conn *sql.DB, nameQuery string) ([]model.ProductMasterView, error) {
	const q = `
		SELECT
			j.JC000, j.JC009, j.JC018, j.JC022, j.JC030, j.JC037, j.JC039,
			j.JC044, j.JC050,
			ja.JA006, ja.JA008, ja.JA007
		FROM
			jcshms AS j
		LEFT JOIN
			jancode AS ja ON j.JC000 = ja.JA001
		WHERE
			j.JC018 LIKE ? OR j.JC022 LIKE ?
		ORDER BY
			j.JC022
		LIMIT 500`

	rows, err := conn.Query(q, "%"+nameQuery+"%", "%"+nameQuery+"%")
	if err != nil {
		return nil, fmt.Errorf("SearchJcshmsByName failed: %w", err)
	}
	defer rows.Close()

	var results []model.ProductMasterView
	for rows.Next() {
		var tempJcshms model.JCShms
		var jc000, jc009, jc018, jc022, jc030, jc037, jc039 sql.NullString
		var jc044 sql.NullFloat64
		var jc050 sql.NullString

		if err := rows.Scan(
			&jc000, &jc009, &jc018, &jc022, &jc030, &jc037, &jc039,
			&jc044, &jc050,
			&tempJcshms.JA006, &tempJcshms.JA008, &tempJcshms.JA007,
		); err != nil {
			return nil, err
		}

		tempJcshms.JC037 = jc037.String
		tempJcshms.JC039 = jc039.String
		tempJcshms.JC044 = jc044.Float64

		val, err := strconv.ParseFloat(jc050.String, 64)
		if err != nil {
			tempJcshms.JC050 = 0
			if jc050.String != "" {
				log.Printf("[WARN] JC050のデータが不正なため0に変換しました。製品名: %s, 元の値: '%s'", jc018.String, jc050.String)
			}
		} else {
			tempJcshms.JC050 = val
		}

		var nhiPrice float64
		if tempJcshms.JC044 > 0 {
			nhiPrice = tempJcshms.JC050 / tempJcshms.JC044
		}

		pkg := units.FormatPackageSpec(&tempJcshms)

		yjUnitName := units.ResolveName(jc039.String)
		janUnitCodeStr := tempJcshms.JA007.String
		var janUnitName string
		if janUnitCodeStr == "0" || janUnitCodeStr == "" {
			janUnitName = yjUnitName
		} else {
			janUnitName = units.ResolveName(janUnitCodeStr)
		}

		// ▼▼▼ 修正箇所 ▼▼▼
		// 文字列のJAN単位コードを整数(int)に変換して、ProductMaster構造体にセットする
		var janUnitCodeInt int
		if tempJcshms.JA007.Valid {
			val, err := strconv.Atoi(tempJcshms.JA007.String)
			if err == nil {
				janUnitCodeInt = val
			}
		}
		// ▲▲▲ ここまで ▲▲▲

		view := model.ProductMasterView{
			ProductMaster: model.ProductMaster{
				ProductCode:     jc000.String,
				YjCode:          jc009.String,
				ProductName:     jc018.String,
				KanaName:        jc022.String,
				MakerName:       jc030.String,
				PackageSpec:     jc037.String,
				YjUnitName:      yjUnitName,
				JanUnitName:     janUnitName,
				JanUnitCode:     janUnitCodeInt, // 修正した値をセット
				YjPackUnitQty:   jc044.Float64,
				JanPackInnerQty: tempJcshms.JA006.Float64,
				JanPackUnitQty:  tempJcshms.JA008.Float64,
				NhiPrice:        nhiPrice,
			},
			FormattedPackageSpec: pkg,
		}
		results = append(results, view)
	}
	return results, nil
}
