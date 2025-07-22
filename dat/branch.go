// File: dat/branch.go
package dat

import (
	"database/sql"
	"fmt"
	"karashi/db"
	"karashi/model"
	"karashi/tani"
	"strconv"
)

// ExecuteDatBranching is the main branching logic for DAT records.
func ExecuteDatBranching(conn *sql.DB, prec model.ARInput) (model.ARInput, error) {
	key := prec.Ajc
	if key == "0000000000000" {
		key = fmt.Sprintf("9999999999999%s", prec.Apname)
	}

	master, err := db.GetMaMasterByCode(conn, key)
	if err != nil {
		return model.ARInput{}, err
	}
	if master != nil {
		if prec.Ajc == "0000000000000" {
			return processDatBranch1(prec, master)
		}
		return processDatBranch3(prec, master)
	}

	jcshms, err := db.GetJcshmsByJan(conn, prec.Ajc)
	if err != nil {
		return model.ARInput{}, err
	}
	if jcshms != nil {
		if jcshms.JC009 != "" {
			return processDatBranch4(conn, prec, jcshms)
		}
		return processDatBranch5(conn, prec, jcshms)
	}

	if prec.Ajc == "0000000000000" {
		return processDatBranch2(conn, prec)
	}
	return processDatBranch6(conn, prec)
}

// --- Helper Functions for Each Branch ---

// mrdat1: No JAN, ma_master exists
func processDatBranch1(prec model.ARInput, master *model.MaMaster) (model.ARInput, error) {
	prec.Ayj = master.MA009
	prec.Apname = master.MA018 // <-- Fix: Add product name
	prec.Ama = "1"
	prec.Apkg = fmt.Sprintf("%s %g%s", master.MA037, master.MA044, tani.ResolveName(master.MA039))
	prec.Amaker = master.MA030
	return prec, nil
}

// mrdat2: No JAN, no ma_master
func processDatBranch2(conn *sql.DB, prec model.ARInput) (model.ARInput, error) {
	newYj, err := db.NextSequence(conn, "MA2Y")
	if err != nil {
		return model.ARInput{}, err
	}
	prec.Ajc = fmt.Sprintf("9999999999999%s", prec.Apname)
	prec.Ayj = newYj
	prec.Ama = "2"

	masterInput := model.MaMasterInput{MA000: prec.Ajc, MA009: prec.Ayj, MA018: prec.Apname}
	if err := db.CreateMaMaster(conn, masterInput); err != nil {
		return model.ARInput{}, err
	}
	return prec, nil
}

// mrdat3: JAN exists, ma_master exists
func processDatBranch3(prec model.ARInput, master *model.MaMaster) (model.ARInput, error) {
	prec.Ayj = master.MA009
	prec.Apname = master.MA018 // <-- Fix: Add product name
	prec.Akana = master.MA022
	prec.Amaker = master.MA030
	prec.Ayjpu = master.MA044
	prec.Ajpu = master.MA133
	prec.Ayjunitnm = tani.ResolveName(master.MA039)
	prec.Ajanqty = prec.Adatqty * master.MA133
	prec.Ayjqty = prec.Adatqty * prec.Ayjpu

	ajanunitcode := strconv.Itoa(master.MA132)
	var ajanunitnm string
	if ajanunitcode == "0" {
		ajanunitnm = prec.Ayjunitnm
	} else {
		ajanunitnm = tani.ResolveName(ajanunitcode)
	}
	prec.Ajanunitcode = ajanunitcode
	prec.Ajanunitnm = ajanunitnm

	pkg := fmt.Sprintf("%s %g%s", master.MA037, master.MA044, prec.Ayjunitnm)
	if master.MA131 != 0 && master.MA133 != 0 {
		pkg += fmt.Sprintf(" (%g%s×%g%s)",
			master.MA131, prec.Ayjunitnm,
			master.MA133, ajanunitnm)
	}
	prec.Apkg = pkg

	prec.Ama = "3"
	return prec, nil
}

// mrdat4: JAN exists, no ma_master, JCSHMS exists (with YJ)
func processDatBranch4(conn *sql.DB, prec model.ARInput, jcshms *db.JCShms) (model.ARInput, error) {
	masterInput := model.MaMasterInput{
		MA000: prec.Ajc, MA009: jcshms.JC009, MA018: jcshms.JC018, MA022: jcshms.JC022, MA030: jcshms.JC030,
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

	prec.Ayj = jcshms.JC009
	prec.Apname = jcshms.JC018 // <-- Fix: Add product name
	prec.Akana = jcshms.JC022
	prec.Amaker = jcshms.JC030
	prec.Ayjpu = jcshms.JC044
	prec.Ajpu = jcshms.JA008.Float64
	prec.Ayjunitnm = tani.ResolveName(jcshms.JC039)
	if jcshms.JA008.Valid {
		prec.Ajanqty = prec.Adatqty * jcshms.JA008.Float64
	}
	prec.Ayjqty = prec.Adatqty * prec.Ayjpu

	ajanunitcode := jcshms.JA007.String
	var ajanunitnm string
	if ajanunitcode == "0" || ajanunitcode == "" {
		ajanunitnm = prec.Ayjunitnm
	} else {
		ajanunitnm = tani.ResolveName(ajanunitcode)
	}
	prec.Ajanunitcode = ajanunitcode
	prec.Ajanunitnm = ajanunitnm

	pkg := fmt.Sprintf("%s %g%s", jcshms.JC037, jcshms.JC044, prec.Ayjunitnm)
	if jcshms.JA006.Valid && jcshms.JA008.Valid {
		pkg += fmt.Sprintf(" (%g%s×%g%s)",
			jcshms.JA006.Float64, prec.Ayjunitnm,
			jcshms.JA008.Float64, ajanunitnm)
	}
	prec.Apkg = pkg

	prec.Ama = "4"
	return prec, nil
}

// mrdat5: JAN exists, no ma_master, JCSHMS exists (no YJ)
func processDatBranch5(conn *sql.DB, prec model.ARInput, jcshms *db.JCShms) (model.ARInput, error) {
	newYj, err := db.NextSequence(conn, "MA2Y")
	if err != nil {
		return model.ARInput{}, err
	}
	prec.Ayj = newYj
	prec.Apname = jcshms.JC018 // <-- Fix: Add product name
	prec.Akana = jcshms.JC022
	prec.Amaker = jcshms.JC030
	prec.Ayjpu = jcshms.JC044
	prec.Ajpu = jcshms.JA008.Float64
	prec.Ayjunitnm = tani.ResolveName(jcshms.JC039)
	if jcshms.JA008.Valid {
		prec.Ajanqty = prec.Adatqty * jcshms.JA008.Float64
	}
	prec.Ayjqty = prec.Adatqty * prec.Ayjpu

	ajanunitcode := jcshms.JA007.String
	var ajanunitnm string
	if ajanunitcode == "0" || ajanunitcode == "" {
		ajanunitnm = prec.Ayjunitnm
	} else {
		ajanunitnm = tani.ResolveName(ajanunitcode)
	}
	prec.Ajanunitcode = ajanunitcode
	prec.Ajanunitnm = ajanunitnm

	pkg := fmt.Sprintf("%s %g%s", jcshms.JC037, jcshms.JC044, prec.Ayjunitnm)
	if jcshms.JA006.Valid && jcshms.JA008.Valid {
		pkg += fmt.Sprintf(" (%g%s×%g%s)",
			jcshms.JA006.Float64, prec.Ayjunitnm,
			jcshms.JA008.Float64, ajanunitnm)
	}
	prec.Apkg = pkg

	masterInput := model.MaMasterInput{
		MA000: prec.Ajc, MA009: prec.Ayj, MA018: jcshms.JC018, MA022: jcshms.JC022, MA030: jcshms.JC030,
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

	prec.Ama = "5"
	return prec, nil
}

// mrdat6: JAN exists, no ma_master, no JCSHMS
func processDatBranch6(conn *sql.DB, prec model.ARInput) (model.ARInput, error) {
	newYj, err := db.NextSequence(conn, "MA2Y")
	if err != nil {
		return model.ARInput{}, err
	}
	prec.Ayj = newYj
	prec.Ama = "6"

	masterInput := model.MaMasterInput{MA000: prec.Ajc, MA009: prec.Ayj, MA018: prec.Apname}
	if err := db.CreateMaMaster(conn, masterInput); err != nil {
		return model.ARInput{}, err
	}
	return prec, nil
}
