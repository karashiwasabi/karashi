// File: central/constants.go
package central

// ProcessFlagMA の値を定数として定義
const (
	FlagComplete    = "COMPLETE"    // データ完了（既存マスタ、またはJCSHMS由来）
	FlagProvisional = "PROVISIONAL" // 暫定データ（最小情報からの自動採番）、継続的な更新対象
)
