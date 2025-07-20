// File: usage/mausage2.go
package usage

import (
	"database/sql"
	"fmt"

	"karashi/db"
)

// HandleBranch2MA は Branch-2（JC 空・MA 未登録）のレコードを受け取り、
// model.ARInput のスライスを返します。
// 第1引数 conn は DB 接続、parsed はフィルタ済み ParsedUsage のリストです。
func HandleBranch2MA(conn *sql.DB, parsed []ParsedUsage) ([]ARInput, error) {
	var out []ARInput

	for i, rec := range parsed {
		// Branch-2 条件: JC が空 && MA 未登録
		if rec.Jc != "" {
			continue
		}
		exists, err := db.ExistsByJan(conn, rec.Jc)
		if err != nil {
			return nil, fmt.Errorf("ExistsByJan error: %w", err)
		}
		if exists {
			continue
		}

		// 合成 JAN コード ("9999999999999"+品名)
		syntheticJan := fmt.Sprintf("9999999999999%s", rec.Pname)

		// 自動採番 (例: code_sequences から次の値を取得)
		seq, err := db.NextSequence(conn, "MA2Y")
		if err != nil {
			return nil, fmt.Errorf("NextSequence error: %w", err)
		}

		// ARInput を組み立て
		ar := ARInput{
			Adate:  rec.Date,
			Apcode: seq,                    // 自動採番
			Arpnum: seq,                    // 伝票番号にも同じシーケンス
			Alnum:  fmt.Sprintf("%d", i+1), // 行番号
			Aflag:  BranchAflag,            // =3
			Ajc:    syntheticJan,           // JC フィールドに合成JAN
			Ayj:    rec.Yj,                 // YJ コード
			Apname: rec.Pname,              // 品名
			// 他フィールドはゼロ値 or 空文字で委ねる
		}
		out = append(out, ar)
	}

	return out, nil
}
