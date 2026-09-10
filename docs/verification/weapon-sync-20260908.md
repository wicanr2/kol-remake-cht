# 戰鬥武器數值同步修復

2026-09-08，依 SPEC-010 已 READY 的完整物品搬移契約修復。
稽核發現 TakeItem 更新 Actor.Weapon，但未更新 Member.Weapon；指令列及資訊框
讀取後者，因此拾取不同類武器後可能仍提供舊武器指令，或顯示錯誤彈藥種類。
DropItem、Knockback 及背包滿時 TakeSpoil 亦有同類遺漏。

本次同步四條既有路徑：丟棄／擊落清除 Member.Weapon；拾取／戰利品交換帶入
地面物品的完整 Weapon 值。沒有新增規則、資料格式或持久庫存。

新增回歸測試直接驗證：丟劍後撿 Elf Bow，指令列出現射擊並移除近戰武器指令，
資訊框顯示 Elf Bow／ARROWS:20；背包滿時換入弩，兩份武器資料一致，原劍完整留在地面。
完整 `go test ./internal/... ./cmd/... -count=1 -timeout 120s` 與建置通過。
這是 remake 資料鏈回歸，未新增 dosgolem 收據，也未完成正常 GUI 拾取異種武器驗收。

產物：`workplace/weapon-sync-20260908/kol`，SHA-256
`c43ca217329d4fc96bab5997757b3a8ec208936bfcae76bd874343ccecb91bd1`。
同目錄保存 tests.log、inputs.sha256、kol.sha256；修改檔執行前後雜湊相同，
所有輸出擁有權 1000:1000。

使用既有 kol-package:20260907，來源唯讀、斷網、4 GiB、2 CPU、256 程序上限，
外層逾時 240 秒；一次性容器 --rm，Xvfb 由 trap 清理。結束後專案容器清單為空。
AppImage 未重打包，未提交或推送。敵人拾取的可變裝備顯示與已知同環尋敵差異仍待處理。
