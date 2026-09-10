# 參戰者初始智力轉接修復

日期：2026-09-08。狀態：已修正，完整測試與建置通過。

正常建場 `NewQuest` 已複製力量、快捷等屬性，卻遺漏隊員 Member.Intellect
與敵人 Monster.Intellect。`restoreStats` 只有在法術倒數到期時才複製智力，
不是每次 BeginPlanning 都會補回；因此初始值零會影響 EnemyIntellectCheck 等既有規則。
城鎮轉接已在 cmd/kol/main.go 複製角色智力，缺口發生在 battle 內部建場。

修正：建立參戰者時複製兩方的既有智力資料，不更改智力公式、AI 分支或法術規則。
測試從 NewQuest 建場，核對隊員與敵人的初始值，以及尚未受法術時的智力判定
與法術到期恢復後相同，避免只驗單獨欄位而漏掉初始化／恢復的不對稱。

這是既有 typed data／規則轉接修復，不代表尚為 DRAFT 的 SPEC-020 已完成。
規劃門檻生命週期、命令面板、正常原版戰鬥及 AppImage 重建仍未完成。

驗證使用 `kol-package:20260907` 及既有 `workplace/spec-fixes-20260907/run.sh`，
`go test ./internal/... ./cmd/kol -count=1` 與建置通過。產物位於
`workplace/intellect-fix-20260908/`，執行檔 SHA-256：
`43626d1648627dcd4f275d29452694310dec218549bc955b8b3760100e441a83`。
測試退出碼零，檔案擁有權 1000:1000。未提交／推送，容器自動移除。
