// File: inventory/branch.go
package inventory

import (
	"database/sql"
	"fmt"
	"karashi/db"
	"karashi/model"
	"karashi/tani"
	"strconv"
)

// ExecuteInventoryBranchingは、棚卸レコードを正しい分岐ロジックで処理します。
func ExecuteInventoryBranching(conn *sql.DB, rec model.ARInput) (model.ARInput, error) {
	// --- ケース1: JANコードが存在する場合 ---
	if rec.Ajc != "" {
		master, err := db.GetMaMasterByCode(conn, rec.Ajc)
		if err != nil {
			return model.ARInput{}, err
		}
		if master != nil {
			return processInvBranch1(rec, master)
		}

		jcshms, err := db.GetJcshmsByJan(conn, rec.Ajc)
		if err != nil {
			return model.ARInput{}, err
		}
		if jcshms != nil {
			if jcshms.JC009 != "" {
				return processInvBranch2(conn, rec, jcshms) // Branch 2
			}
			return processInvBranch3(conn, rec, jcshms) // Branch 3
		}

		return processInvBranch4(conn, rec) // Branch 4
	}

	// --- ケース2: JANコードが無く、YJコードが存在する場合 ---
	if rec.Ayj != "" {
		return processInvBranch5(conn, rec) // Branch 5
	}

	return rec, nil
}

// --- 各分岐のヘルパー関数 ---

// Branch 1: JANあり, ma_masterあり
func processInvBranch1(rec model.ARInput, master *model.MaMaster) (model.ARInput, error) {
	rec.Ayj = master.MA009
	rec.Apname = master.MA018
	rec.Amaker = master.MA030
	rec.Ayjpu = master.MA044
	rec.Ajpu = master.MA133

	// vvv 包装文字列と単位名のロジックを追加 vvv
	ayjunitnm := tani.ResolveName(master.MA039)
	ajanunitcode := strconv.Itoa(master.MA132)
	var ajanunitnm string
	if ajanunitcode == "0" {
		ajanunitnm = ayjunitnm
	} else {
		ajanunitnm = tani.ResolveName(ajanunitcode)
	}
	rec.Ajanunitnm = ajanunitnm
	rec.Ayjunitnm = ayjunitnm

	pkg := fmt.Sprintf("%s %g%s", master.MA037, master.MA044, ayjunitnm)
	if master.MA131 != 0 && master.MA133 != 0 {
		pkg += fmt.Sprintf(" (%g%s×%g%s)", master.MA131, ayjunitnm, master.MA133, ajanunitnm)
	}
	rec.Apkg = pkg
	// ^^^ ここまで ^^^

	return rec, nil
}

// Branch 2: JANあり, ma_masterなし, JCSHMSあり (YJあり)
func processInvBranch2(conn *sql.DB, rec model.ARInput, jcshms *db.JCShms) (model.ARInput, error) {
	rec.Ayj = jcshms.JC009
	rec.Apname = jcshms.JC018
	rec.Amaker = jcshms.JC030
	rec.Ayjpu = jcshms.JC044
	rec.Ajpu = jcshms.JA008.Float64

	// vvv 包装文字列と単位名のロジックを追加 vvv
	ayjunitnm := tani.ResolveName(jcshms.JC039)
	ajanunitcode := jcshms.JA007.String
	var ajanunitnm string
	if ajanunitcode == "0" || ajanunitcode == "" {
		ajanunitnm = ayjunitnm
	} else {
		ajanunitnm = tani.ResolveName(ajanunitcode)
	}
	rec.Ajanunitnm = ajanunitnm
	rec.Ayjunitnm = ayjunitnm

	pkg := fmt.Sprintf("%s %g%s", jcshms.JC037, jcshms.JC044, ayjunitnm)
	if jcshms.JA006.Valid && jcshms.JA008.Valid {
		pkg += fmt.Sprintf(" (%g%s×%g%s)", jcshms.JA006.Float64, ayjunitnm, jcshms.JA008.Float64, ajanunitnm)
	}
	rec.Apkg = pkg
	// ^^^ ここまで ^^^

	masterInput := model.MaMasterInput{
		MA000: rec.Ajc, MA009: jcshms.JC009, MA018: jcshms.JC018, MA022: jcshms.JC022, MA030: jcshms.JC030,
		MA037: jcshms.JC037, MA039: jcshms.JC039, MA044: jcshms.JC044, MA061: jcshms.JC061,
		MA062: jcshms.JC062, MA063: jcshms.JC063, MA064: jcshms.JC064, MA065: jcshms.JC065,
		MA066: jcshms.JC066, MA131: jcshms.JA006.Float64, MA133: jcshms.JA008.Float64,
	}
	if jcshms.JA007.Valid {
		if val, err := strconv.Atoi(jcshms.JA007.String); err == nil {
			masterInput.MA132 = val
		}
	}
	if err := db.CreateMaMaster(conn, masterInput); err != nil {
		return model.ARInput{}, err
	}
	return rec, nil
}

// Branch 3: JANあり, ma_masterなし, JCSHMSあり (YJなし)
func processInvBranch3(conn *sql.DB, rec model.ARInput, jcshms *db.JCShms) (model.ARInput, error) {
	newYj, err := db.NextSequence(conn, "MA2Y")
	if err != nil {
		return model.ARInput{}, err
	}
	rec.Ayj = newYj
	rec.Apname = jcshms.JC018
	rec.Amaker = jcshms.JC030
	rec.Ayjpu = jcshms.JC044
	rec.Ajpu = jcshms.JA008.Float64

	// vvv 包装文字列と単位名のロジックを追加 vvv
	ayjunitnm := tani.ResolveName(jcshms.JC039)
	ajanunitcode := jcshms.JA007.String
	var ajanunitnm string
	if ajanunitcode == "0" || ajanunitcode == "" {
		ajanunitnm = ayjunitnm
	} else {
		ajanunitnm = tani.ResolveName(ajanunitcode)
	}
	rec.Ajanunitnm = ajanunitnm
	rec.Ayjunitnm = ayjunitnm

	pkg := fmt.Sprintf("%s %g%s", jcshms.JC037, jcshms.JC044, ayjunitnm)
	if jcshms.JA006.Valid && jcshms.JA008.Valid {
		pkg += fmt.Sprintf(" (%g%s×%g%s)", jcshms.JA006.Float64, ayjunitnm, jcshms.JA008.Float64, ajanunitnm)
	}
	rec.Apkg = pkg
	// ^^^ ここまで ^^^

	masterInput := model.MaMasterInput{
		MA000: rec.Ajc, MA009: newYj, MA018: jcshms.JC018, MA022: jcshms.JC022, MA030: jcshms.JC030,
		MA037: jcshms.JC037, MA039: jcshms.JC039, MA044: jcshms.JC044, MA061: jcshms.JC061,
		MA062: jcshms.JC062, MA063: jcshms.JC063, MA064: jcshms.JC064, MA065: jcshms.JC065,
		MA066: jcshms.JC066, MA131: jcshms.JA006.Float64, MA133: jcshms.JA008.Float64,
	}
	if jcshms.JA007.Valid {
		if val, err := strconv.Atoi(jcshms.JA007.String); err == nil {
			masterInput.MA132 = val
		}
	}
	if err := db.CreateMaMaster(conn, masterInput); err != nil {
		return model.ARInput{}, err
	}
	return rec, nil
}

// Branch 4: JANあり, どのマスターにもない
func processInvBranch4(conn *sql.DB, rec model.ARInput) (model.ARInput, error) {
	newYj, err := db.NextSequence(conn, "MA2Y")
	if err != nil {
		return model.ARInput{}, err
	}
	rec.Ayj = newYj

	masterInput := model.MaMasterInput{MA000: rec.Ajc, MA009: newYj, MA018: rec.Apname}
	if err := db.CreateMaMaster(conn, masterInput); err != nil {
		return model.ARInput{}, err
	}
	return rec, nil
}

// Branch 5: JANなし, YJあり
func processInvBranch5(conn *sql.DB, rec model.ARInput) (model.ARInput, error) {
	newJan := fmt.Sprintf("9999999999999%s", rec.Apname)
	rec.Ajc = newJan

	existingMaster, err := db.GetMaMasterByCode(conn, newJan)
	if err != nil {
		return model.ARInput{}, err
	}
	if existingMaster != nil {
		rec.Apname = existingMaster.MA018
		rec.Apkg = fmt.Sprintf("%s %g%s", existingMaster.MA037, existingMaster.MA044, tani.ResolveName(existingMaster.MA039))
		rec.Amaker = existingMaster.MA030
		rec.Ayjpu = existingMaster.MA044
		rec.Ajpu = existingMaster.MA133
		return rec, nil
	}

	masterInput := model.MaMasterInput{MA000: newJan, MA009: rec.Ayj, MA018: rec.Apname}
	if err := db.CreateMaMaster(conn, masterInput); err != nil {
		return model.ARInput{}, err
	}
	return rec, nil
}
