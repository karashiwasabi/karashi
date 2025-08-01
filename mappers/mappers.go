// File: mappers/mappers.go (修正版)
package mappers

import (
	"database/sql"
	"karashi/db"
	"karashi/model"
	"karashi/units"
	"strconv"
)

// MapProductMasterToTransactionは、ProductMasterの情報をTransactionRecordにマッピングします。
func MapProductMasterToTransaction(ar *model.TransactionRecord, master *model.ProductMaster) {
	ar.YjCode = master.YjCode
	ar.ProductName = master.ProductName
	ar.KanaName = master.KanaName
	ar.PackageForm = master.PackageSpec
	ar.MakerName = master.MakerName
	ar.YjPackUnitQty = master.YjPackUnitQty
	ar.JanPackUnitQty = master.JanPackUnitQty
	ar.FlagPoison = master.FlagPoison
	ar.FlagDeleterious = master.FlagDeleterious
	ar.FlagNarcotic = master.FlagNarcotic
	ar.FlagPsychotropic = master.FlagPsychotropic
	ar.FlagStimulant = master.FlagStimulant
	ar.FlagStimulantRaw = master.FlagStimulantRaw
	ar.JanPackInnerQty = master.JanPackInnerQty

	yjUnitName := units.ResolveName(master.YjUnitName)
	janUnitCode := strconv.Itoa(master.JanUnitCode)
	var janUnitName string
	if janUnitCode == "0" || janUnitCode == "" {
		janUnitName = yjUnitName
	} else {
		janUnitName = units.ResolveName(janUnitCode)
	}
	ar.JanUnitName = janUnitName
	ar.YjUnitName = yjUnitName
	ar.JanUnitCode = janUnitCode

	// ProductMasterからJCShms相当の構造体を作成して関数を呼び出す
	tempJcshms := model.JCShms{
		JC037: master.PackageSpec,
		JC039: master.YjUnitName,
		JC044: master.YjPackUnitQty,
		JA006: sql.NullFloat64{Float64: master.JanPackInnerQty, Valid: true},
		JA008: sql.NullFloat64{Float64: master.JanPackUnitQty, Valid: true},
		JA007: sql.NullString{String: strconv.Itoa(master.JanUnitCode), Valid: true},
	}
	ar.PackageSpec = units.FormatPackageSpec(&tempJcshms)
}

// MapJcshmsToTransactionは、JCShmsの情報をTransactionRecordにマッピングします。
func MapJcshmsToTransaction(ar *model.TransactionRecord, jcshms *model.JCShms) {
	ar.ProductName = jcshms.JC018
	ar.KanaName = jcshms.JC022
	ar.PackageForm = jcshms.JC037
	ar.MakerName = jcshms.JC030
	ar.YjPackUnitQty = jcshms.JC044
	ar.JanPackInnerQty = jcshms.JA006.Float64
	ar.JanPackUnitQty = jcshms.JA008.Float64
	ar.FlagPoison = jcshms.JC061
	ar.FlagDeleterious = jcshms.JC062
	ar.FlagNarcotic = jcshms.JC063
	ar.FlagPsychotropic = jcshms.JC064
	ar.FlagStimulant = jcshms.JC065
	ar.FlagStimulantRaw = jcshms.JC066

	yjUnitName := units.ResolveName(jcshms.JC039)
	janUnitCode := jcshms.JA007.String
	var janUnitName string
	if janUnitCode == "0" || janUnitCode == "" {
		janUnitName = yjUnitName
	} else {
		janUnitName = units.ResolveName(janUnitCode)
	}
	ar.JanUnitName = janUnitName
	ar.YjUnitName = yjUnitName
	ar.JanUnitCode = janUnitCode

	// 新しい共通関数を呼び出して包装表記を生成
	ar.PackageSpec = units.FormatPackageSpec(jcshms)
}

// CreateMasterFromJcshmsInTxは、既存のトランザクション内でJCSHMSからマスターを作成します。
func CreateMasterFromJcshmsInTx(tx *sql.Tx, jan, yj string, jcshms *model.JCShms) error {
	var nhiPrice float64
	if jcshms.JC044 > 0 {
		nhiPrice = jcshms.JC050 / jcshms.JC044
	}
	masterInput := model.ProductMasterInput{
		ProductCode:      jan,
		YjCode:           yj,
		Origin:           "JCSHMS",
		ProductName:      jcshms.JC018,
		KanaName:         jcshms.JC022,
		MakerName:        jcshms.JC030,
		PackageSpec:      jcshms.JC037,
		YjUnitName:       jcshms.JC039,
		YjPackUnitQty:    jcshms.JC044,
		FlagPoison:       jcshms.JC061,
		FlagDeleterious:  jcshms.JC062,
		FlagNarcotic:     jcshms.JC063,
		FlagPsychotropic: jcshms.JC064,
		FlagStimulant:    jcshms.JC065,
		FlagStimulantRaw: jcshms.JC066,
		JanPackInnerQty:  jcshms.JA006.Float64,
		JanPackUnitQty:   jcshms.JA008.Float64,
		NhiPrice:         nhiPrice,
	}

	if jcshms.JA007.Valid {
		if val, err := strconv.Atoi(jcshms.JA007.String); err == nil {
			masterInput.JanUnitCode = val
		}
	}

	return db.CreateProductMasterInTx(tx, masterInput)
}
