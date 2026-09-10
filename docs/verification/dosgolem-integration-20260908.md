# 新版 dosgolem 相容性修正（2026-09-08）

## 目前狀態

| 項目 | 結果 |
| --- | --- |
| 現用分支 | `/home/anr2/cht/dosgolem-kolrand` 的 `kol-force-call` |
| 上游整合 | 已快轉至 `fd746405cf0ef9e535121e215b935b0bf59cd1de` |
| 本機 dosgolem 修正 | 公開 `FileOp` 補回 `Whence`、`Failed`，逐筆原值轉出；附回歸測試與 READY 契約補正 |
| KOL 相容性 | 檔案、記憶體、暫存器診斷改用新版欄位，編譯通過 |
| 規則對照 | 六項通過，126.714 秒 |
| 正常原版畫面 | **未通過：六張 RGB 擷取全黑**；不可宣稱正常顯示或操作一致 |
| AppImage | 未重新打包，前輪修復仍在工作樹 |

dosgolem 與 KOL 的修改均尚未提交或推送。前輪 KOL 修復保留，存檔 schema 未改。

## 相容性修正

`internal/oracle/boot_test.go` 的三處檔案日誌統一走 `logFileOp`：保留步數、操作、AH、檔名、handle、seek 基準、要求量、位置、實際長度及失敗旗標。新版 `Arg`／`Pos` 按原義顯示，不硬套舊 `Offset`／`Want`／`Got` 的語意。

兩處記憶體日誌保留原始 `Fn`、`BX`、`ES`、`AX`、`OK`；不假造新版未提供的舊欄位。`Regs()` 使用結構格式 `%+v`。

dosgolem 的 `oracle/oracle.go` 補回內部原本已記錄的兩欄，沒有修改 DOS 執行規則。`oracle/file_trace_test.go` 同時驗證成功與失敗定位的公開軌跡。依據是既有 `docs/spec/004-dos-bios-services.md` 的失敗呼叫可追查契約，補正附於同檔。

## 驗證結果與界線

1. dosgolem：`internal/cpu/...`、`internal/dos/...`、`internal/dosfile/...`、`internal/machine/...`、`oracle/...` 通過。首輪包含本機 `testdata/8088` 語料，CPU 套件耗時 96.827 秒；沒有宣稱超出該測試既有容許界線的逐週期一致。
2. KOL：`TestFatigueRatioMatchesOriginal`、`TestSkillPenaltyMatchesOriginal`、`TestPartySkillRowFollowsWeaponClass`、`TestTakeDamageMatchesOriginal`、`TestRandStreamMatchesOriginal`、`TestRollMatchesOriginal` 全部通過。亂數含 3,000 抽，有界骰含 400 抽；屬受控常式驗證。
3. 正常入口：使用未修改的 `KOL.EXE` 與原始遊戲目錄，只送 BIOS 按鍵，不寫記憶體、不替換執行檔。六段各 2,000 萬指令，共 1.2 億指令，能讀取 TITLE、SET0PICS、SET1PICS、BACKPICS；無缺檔與未實作服務。五個送入按鍵尚未被讀取，不能推論書本操作通過。
4. 逐張檢視 `normal-00.png` 至 `normal-05.png`，全部全黑。最初僅檢查執行錯誤的探針曾顯示 PASS，已補上全黑畫面必須失敗的閘門；歷史 PASS 不作畫面驗收。新版 `ScreenRGB()` 經 `PlanarRGB()` 使用 DAC 色盤，該路徑須另補顯示驗證；不以自行套色冒充原版收據。

本輪相容性修正已完成；正常 RGB 顯示與書本玩家路徑仍是後續驗證阻塞。不得將上游更新直接等同 AppImage 與原版操作一致。

加入畫面閘門後，以同一容器設定與正常入口乾淨重跑，13.465 秒後正確回報失敗。六段的非黑 DAC 色數均為 0；例如第 4 段有 16,612 個非零色號像素、第 5 段有 6,053 個，證實索引像素與 RGB 顯示不可混為同一驗收。這是 dosgolem 顯示驗證缺口，不是 KOL remake 規則失敗。

## 重跑與收據

採用既有 `kol-package:20260907`（Go 1.26.7），原始資料與來源唯讀，非 root、斷網、限額 Docker。容器內腳本為 `tools/oracle/check_updated_dosgolem.sh`，正常入口探針為同目錄的 `normal_startup_test.go`。

掛載 `/src`＝KOL、`/repo`＝dosgolem（皆唯讀），`/out`＝明確可寫輸出後，在容器執行：

```sh
bash /src/tools/oracle/check_updated_dosgolem.sh
```

`KOL_CHECK=normal` 僅重跑正常入口。腳本使用最小來源副本，不攜帶外部 CPU 語料；重驗完整 CPU 語料時需另將 `testdata/8088` 唯讀掛入測試副本。Go 工作區在容器建立，不更動正式 `go.mod`。

本機收據在 `workplace/dosgolem-integration-20260908/`：核心／規則／正常入口日誌、六張 PNG、manifest、dosgolem 本機差異及回歸測試副本。原版圖片僅本機，未加入公開產物。

收尾確認本輪 `kol-dosgolem-integration`、`kol-dosgolem-normal` 容器皆已移除；本輪寫入檔案與產物為 1000:1000。沒有清理其他專案的容器。
