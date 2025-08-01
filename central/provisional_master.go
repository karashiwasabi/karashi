// File: central/provisional_master.go (修正版)
package central

import (
	"database/sql"
	"fmt"
	"karashi/db"
	"karashi/model"
)

// createProvisionalMasterは、暫定マスターを作成し、メモリ上のマップを更新する共通関数です。
func createProvisionalMaster(
	tx *sql.Tx,
	key string,
	janCode string,
	productName string,
	mastersMap map[string]*model.ProductMaster,
) (string, string, error) { // yjCodeとproductCodeを返します

	// ★★★ 呼び出し方を修正: プレフィックス"MA2Y"とパディング8桁を追加 ★★★
	newYj, err := db.NextSequenceInTx(tx, "MA2Y", "MA2Y", 8)
	if err != nil {
		return "", "", fmt.Errorf("failed to get next sequence for provisional master: %w", err)
	}

	productCode := janCode
	if key != janCode {
		productCode = key
	}

	minMasterInput := model.ProductMasterInput{
		ProductCode: productCode,
		YjCode:      newYj,
		ProductName: productName,
		Origin:      "PROVISIONAL",
	}

	if err := db.CreateProductMasterInTx(tx, minMasterInput); err != nil {
		return "", "", fmt.Errorf("failed to create provisional master in tx for key %s: %w", key, err)
	}

	newMasterForMap := &model.ProductMaster{
		ProductCode: minMasterInput.ProductCode,
		YjCode:      minMasterInput.YjCode,
		ProductName: minMasterInput.ProductName,
		Origin:      minMasterInput.Origin,
	}
	mastersMap[key] = newMasterForMap

	return newYj, productCode, nil
}
