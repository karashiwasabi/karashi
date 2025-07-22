// File: usage/branch.go
package usage

import (
	"database/sql"
	"fmt"
	"karashi/db"
	"karashi/model"
	"karashi/tani"
	"strconv"
)

// ExecuteBranching determines which logic branch to follow based on the parsed USAGE data.
func ExecuteBranching(conn *sql.DB, pu model.ParsedUsage) (model.ARInput, error) {
	key := pu.Jc
	isSyntheticKey := false
	if key == "" {
		key = fmt.Sprintf("9999999999999%s", pu.Pname)
		isSyntheticKey = true
	}

	master, err := db.GetMaMasterByCode(conn, key)
	if err != nil {
		return model.ARInput{}, fmt.Errorf("failed to get ma_master by key %s: %w", key, err)
	}

	if master != nil {
		return processBranch1(pu, master)
	}

	if isSyntheticKey {
		return processBranch2(conn, pu)
	}

	jcshms, err := db.GetJcshmsByJan(conn, pu.Jc)
	if err != nil {
		return model.ARInput{}, fmt.Errorf("failed to get jcshms by jan %s: %w", pu.Jc, err)
	}

	if jcshms != nil {
		if jcshms.JC009 != "" {
			return processBranch5(conn, pu, jcshms)
		} else {
			return processBranch4(conn, pu, jcshms)
		}
	} else {
		return processBranch6(conn, pu)
	}
}

// --- Helper Functions for Each Branch ---

// processBranch1: Use an existing ma_master record.
func processBranch1(pu model.ParsedUsage, master *model.MaMaster) (model.ARInput, error) {
	ayjunitnm := tani.ResolveName(master.MA039)
	ajanunitcode := strconv.Itoa(master.MA132)
	var ajanunitnm string
	if ajanunitcode == "0" {
		ajanunitnm = ayjunitnm
	} else {
		ajanunitnm = tani.ResolveName(ajanunitcode)
	}

	pkg := fmt.Sprintf("%s %g%s", master.MA037, master.MA044, ayjunitnm)
	if master.MA131 != 0 && master.MA133 != 0 {
		pkg += fmt.Sprintf(" (%g%s×%g%s)", master.MA131, ayjunitnm, master.MA133, ajanunitnm)
	}

	// vvv ここが修正箇所 vvv
	janQty := 0.0
	if master.MA131 != 0 { // MA133 から MA131 に変更
		janQty = pu.YjQty / master.MA131
	}
	// ^^^ ここまで ^^^

	return model.ARInput{
		Adate:              pu.Date,
		Aflag:              BranchAflag,
		Ajc:                master.MA000,
		Ayj:                master.MA009,
		Apname:             master.MA018,
		Akana:              master.MA022,
		Apkg:               pkg,
		Amaker:             master.MA030,
		Ajanqty:            janQty,
		Ajpu:               master.MA133,
		Ajanunitnm:         ajanunitnm,
		Ajanunitcode:       ajanunitcode,
		Ayjqty:             pu.YjQty,
		Ayjpu:              master.MA044,
		Ayjunitnm:          ayjunitnm,
		Adokuyaku:          master.MA061,
		Agekiyaku:          master.MA062,
		Amayaku:            master.MA063,
		Akouseisinyaku:     master.MA064,
		Akakuseizai:        master.MA065,
		Akakuseizaigenryou: master.MA066,
		Ama:                Ama1,
	}, nil
}

// processBranch2: Register a new item that has no JAN code.
func processBranch2(conn *sql.DB, pu model.ParsedUsage) (model.ARInput, error) {
	newYjCode, err := db.NextSequence(conn, "MA2Y")
	if err != nil {
		return model.ARInput{}, err
	}
	syntheticJan := fmt.Sprintf("9999999999999%s", pu.Pname)

	masterInput := model.MaMasterInput{
		MA000: syntheticJan,
		MA009: newYjCode,
		MA018: pu.Pname,
		MA039: tani.ResolveCode(pu.YjUnitName),
	}
	if err := db.CreateMaMaster(conn, masterInput); err != nil {
		return model.ARInput{}, err
	}

	return model.ARInput{
		Adate:     pu.Date,
		Aflag:     BranchAflag,
		Ajc:       syntheticJan,
		Ayj:       newYjCode,
		Apname:    pu.Pname,
		Ayjqty:    pu.YjQty,
		Ayjunitnm: pu.YjUnitName,
		Ama:       Ama2,
	}, nil
}

// processBranch4: Create a new ma_master record based on JCSHMS data and assign a new YJ code.
func processBranch4(conn *sql.DB, pu model.ParsedUsage, jcshms *db.JCShms) (model.ARInput, error) {
	newYjCode, err := db.NextSequence(conn, "MA2Y")
	if err != nil {
		return model.ARInput{}, err
	}
	if err := createMasterFromJcshms(conn, pu.Jc, newYjCode, jcshms); err != nil {
		return model.ARInput{}, err
	}
	return createARInputFromJcshms(pu, newYjCode, jcshms, Ama4)
}

// processBranch5: Create a new ma_master record based on JCSHMS data using the existing YJ code.
func processBranch5(conn *sql.DB, pu model.ParsedUsage, jcshms *db.JCShms) (model.ARInput, error) {
	if err := createMasterFromJcshms(conn, pu.Jc, jcshms.JC009, jcshms); err != nil {
		return model.ARInput{}, err
	}
	return createARInputFromJcshms(pu, jcshms.JC009, jcshms, Ama5)
}

// processBranch6: Register a new item based only on its JAN code.
func processBranch6(conn *sql.DB, pu model.ParsedUsage) (model.ARInput, error) {
	newYjCode, err := db.NextSequence(conn, "MA2Y")
	if err != nil {
		return model.ARInput{}, err
	}

	masterInput := model.MaMasterInput{
		MA000: pu.Jc,
		MA009: newYjCode,
		MA018: pu.Pname,
		MA039: tani.ResolveCode(pu.YjUnitName),
	}
	if err := db.CreateMaMaster(conn, masterInput); err != nil {
		return model.ARInput{}, err
	}

	return model.ARInput{
		Adate:     pu.Date,
		Aflag:     BranchAflag,
		Ajc:       pu.Jc,
		Ayj:       newYjCode,
		Apname:    pu.Pname,
		Ayjqty:    pu.YjQty,
		Ayjunitnm: pu.YjUnitName,
		Ama:       Ama6,
	}, nil
}

// --- Common Helpers for Branches 4 & 5 ---

// createMasterFromJcshms creates a MaMasterInput struct from JCSHMS data and saves it.
func createMasterFromJcshms(conn *sql.DB, jan, yj string, jcshms *db.JCShms) error {
	masterInput := model.MaMasterInput{
		MA000: jan,
		MA009: yj,
		MA018: jcshms.JC018,
		MA022: jcshms.JC022,
		MA030: jcshms.JC030,
		MA037: jcshms.JC037,
		MA039: jcshms.JC039,
		MA044: jcshms.JC044,
		MA061: jcshms.JC061,
		MA062: jcshms.JC062,
		MA063: jcshms.JC063,
		MA064: jcshms.JC064,
		MA065: jcshms.JC065,
		MA066: jcshms.JC066,
		MA131: jcshms.JA006.Float64,
		MA133: jcshms.JA008.Float64,
	}

	if jcshms.JA007.Valid {
		if val, err := strconv.Atoi(jcshms.JA007.String); err == nil {
			masterInput.MA132 = val
		}
	}

	return db.CreateMaMaster(conn, masterInput)
}

// createARInputFromJcshms creates the final ARInput record from JCSHMS data.
func createARInputFromJcshms(pu model.ParsedUsage, yj string, jcshms *db.JCShms, ama string) (model.ARInput, error) {
	ja006 := jcshms.JA006.Float64
	ja007Str := jcshms.JA007.String
	ja008 := jcshms.JA008.Float64

	ayjunitnm := tani.ResolveName(jcshms.JC039)
	var ajanunitnm string
	if ja007Str == "0" || ja007Str == "" {
		ajanunitnm = ayjunitnm
	} else {
		ajanunitnm = tani.ResolveName(ja007Str)
	}

	pkg := fmt.Sprintf("%s %g%s", jcshms.JC037, jcshms.JC044, ayjunitnm)
	if ja006 != 0 && ja008 != 0 {
		pkg += fmt.Sprintf(" (%g%s×%g%s)", ja006, ayjunitnm, ja008, ajanunitnm)
	}

	// ここの計算式はJA006を使うので正しいままです
	janQty := 0.0
	if jcshms.JA006.Valid && ja006 != 0 {
		janQty = pu.YjQty / ja006
	}

	return model.ARInput{
		Adate:              pu.Date,
		Aflag:              BranchAflag,
		Ajc:                pu.Jc,
		Ayj:                yj,
		Apname:             jcshms.JC018,
		Akana:              jcshms.JC022,
		Apkg:               pkg,
		Amaker:             jcshms.JC030,
		Ajanqty:            janQty,
		Ajpu:               ja008,
		Ajanunitnm:         ajanunitnm,
		Ajanunitcode:       ja007Str,
		Ayjqty:             pu.YjQty,
		Ayjpu:              jcshms.JC044,
		Ayjunitnm:          ayjunitnm,
		Adokuyaku:          jcshms.JC061,
		Agekiyaku:          jcshms.JC062,
		Amayaku:            jcshms.JC063,
		Akouseisinyaku:     jcshms.JC064,
		Akakuseizai:        jcshms.JC065,
		Akakuseizaigenryou: jcshms.JC066,
		Ama:                ama,
	}, nil
}
