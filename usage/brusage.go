package usage

import (
	"log"
)

// BranchUsage は現段階では ParsedUsage の JC 有無だけをログに出力します。
// 「JC が空白 → group=12」「JC あり → group=3456」を示します。
func branchUsage(parsed []ParsedUsage) {
	for _, rec := range parsed {
		group := "12"
		if rec.Jc != "" {
			group = "3456"
		}
		log.Printf("JC: %q    branch group: %s", rec.Jc, group)
	}
}
