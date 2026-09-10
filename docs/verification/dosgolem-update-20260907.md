# dosgolem 上游更新評估（2026-09-07）

結論：建議整合上游，但須先完成 KOL 診斷介面相容性修正，不能只依 Git 無衝突就直接切換驗證工具。

## 目前狀態

| 項目 | 本輪確認 |
| --- | --- |
| 遠端預設分支 | `origin/master`，不是 `main` |
| 已抓取上游 | `fd746405cf0ef9e535121e215b935b0bf59cd1de` |
| 實際使用分支 | `kol-force-call` |
| 實際使用提交 | `7f30ad8427086738b037c50a417baf3d0d56a5d2`，本輪未切換 |
| 分歧 | 現用分支獨有 0 筆，上游新增 352 筆，可快轉 |
| 工作樹 | dosgolem 保持乾淨；KOL 原有修正保留 |

已執行 `git fetch origin`。上游已在 `0bd051f` 合入 `kol-force-call`；不需要重新挑選該分支的舊提交。此次先抓取及評估，未執行會移動目前分支的 `pull`／合併，也未推送。

## 為何值得整合

上游包含 EGA 平面顯示、BIOS 鍵盤與 IRQ1、滑鼠回呼、DOS EXEC／記憶體管理及檔案服務等改進，涉及 KOL 原版正常啟動、畫面與操作驗證。這是整合理由，不代表已證實解決 KOL 的全部玩家路徑問題。

## 實際相容性阻塞

- `oracle.FileOps()` 改變欄位，KOL 的 `boot_test.go` 仍引用 `Kind`、`Offset`、`Want`、`Got`、`Into`、`Failed` 等舊介面，無法編譯。
- `MemOps()` 的資料欄位也已改變；舊診斷輸出引用 `Kind`、`Block` 等欄位而無法編譯。
- `Regs()` 的回傳型別與原本 `%s` 日誌格式不再相容，Go 檢查會拒絕。
- 上游內部 `dos.FileOp` 仍有 `Whence`、`Failed`，但公開 `oracle.FileOp` 沒有轉出。不能以刪除失敗資訊的方式宣稱診斷能力完整相容；正式整合應保留這些證據欄位。

在唯讀來源的臨時副本改用結構化日誌，僅為執行規則測試。候選差異保存於 `workplace/dosgolem-update-20260907/kol-api-adaptation.patch`，不是可直接採用的正式修補。

## 驗證與限制

採用既有 `kol-package:20260907`，Go 1.26.7，斷網 Docker、非 root、資源限制與外層逾時；以 Git archive 的上游快照配合 KOL 工作樹來源測試。

上游 `internal/cpu/...`、`internal/dos/...`、`internal/dosfile/...`、`internal/machine/...`、`oracle/...` 所含測試通過。此快照未掛載外部 CPU 測試語料，不能宣稱完整 CPU 語料驗收。

原樣 KOL 測試編譯失敗，分類為上述介面不相容。臨時適配後六項規則測試全部通過（97.881 秒）：疲勞比例、技能懲罰、武器技能列、傷害承受邊界、3,000 次亂數與 400 次有界骰。這些是受控常式驗證，不是正常玩家路徑證明。日誌位於 `workplace/dosgolem-update-20260907/kol-1.log`，結束碼為 0。

本輪未驗收新上游的正常 DOS 玩家路徑，也未重新驗證 AppImage。後續順序為：保留診斷欄位並修正呼叫端 → 規則測試 → 快轉目前分支 → dosgolem 正常路徑重跑 → 恢復 AppImage 對拍。
