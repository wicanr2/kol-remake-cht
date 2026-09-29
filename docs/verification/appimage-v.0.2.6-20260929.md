# 本機 Linux AppImage 驗收與原版同狀態（旅店、戰鬥、存讀檔）

日期：2026-09-29。版本：`v.0.2.6-20260929`。本包含原作衍生素材，只留本機，沒有公開
散布權。來源是 commit `7504f81` 的乾淨工作樹（#102 規劃段按 YES、#104 攻擊式遮罩殘值、
RE-159 敵人規劃、SPEC-036 點陣圖編輯器、存檔 `kol-save-v3`）。

## 交付與封包

- 本機檔案：`dist-all/v.0.2.6-20260929/full/kol-v.0.2.6-20260929-full-linux-x86_64.AppImage`。
- AppImage SHA-256：`73178858b99029e33c339a2b6e45cab0d52d7cc77ecd10af39fd63509e38788f`。
- 內含執行檔 SHA-256：`ae0185488fc2f0f1f266b48580c781258494bcb69c902971f5c87dfe881f4cc6`（解出的
  `usr/bin/kol` 與 `build-inputs.json` 相符）。`build-inputs.json` 記錄 1,148 筆輸入雜湊。
- 建置：`kol-package:20260907`、唯讀 `/src`、全新空 `/out`、UID/GID `1000:1000`、
  `--rm --network none`，`bash /src/tools/build_dist.sh --local-linux-in-container v.0.2.6-20260929`。

## 與原版同狀態

封包裡的執行檔經 `KOL_BIN` 交給 `tools/run_kol_shots.sh`（不從原始碼重建，工作目錄與素材
是工作樹；封包內素材與工作樹逐檔相同由 `verify_local_appimage.py` 核對）。

| 段落 | 比較方式 | 結果 |
|---|---|---|
| 旅店 | `plan-town-inn` 對原版畫格：進旅店、圖示列白框六格、招募一名之後 | 8／8 |
| 旅店招募 | `plan-town-recruit` 對原版畫格：硬碟問句、清單、第二列、入隊 | 9／9 |
| 住宿與存檔 | `plan-town-rest` 對原版畫格：價錢、存到硬碟、隊伍空了問離開、全隊與逐人價錢 | 11／11 |
| 戰鬥 | 8 份戰鬥機制收據（伏擊、逃走兩種、射擊、撿武器、騎走、施法、戰利品）以封包執行檔重跑 | 8／8 |
| 戰鬥與存讀檔長鏈 | 25 條玩家路徑以封包執行檔重跑，軌跡與收據一致（含 full-chain 打贏住宿存檔、full-chain-reload 讀檔續玩） | 25／25 |

戰鬥規則與原版的同狀態證據在同一個 commit 的原版對拍測試（`tools/run_original_parity_tests.sh`：
六回合執行段、敵人規劃、規劃逐格、初態、地形、怪物裝備、點陣圖編輯器），封包執行檔由
那一份原始碼建出。

旅店「招募一名之後」那一對原本用 `-inn` 樣本產生；這一次改用正常按鍵路徑
（`-scene play -play-script "L R RET SPACE SPACE > > > > > RET RET RET"`），同樣通過。

## 正常保存操作

- `tools/verify_complete_save_gui.sh` 以真的 AppImage（`--appimage-extract-and-run`）在 Xvfb 下操作：
  建角兩人、重啟選角與忠實住宿保存、住宿離隊後重載、刪除不搬槽與空槽重用、滿槽、v1 匯入、
  受控世界時間重載，全部通過；存檔 schema 為 `kol-save-v3`。
- `tools/verify_local_appimage.py` 封包內容與操作輸出核對通過。
- 未指定 `-save` 的兩個預設位置：`XDG_CONFIG_HOME` 通過；`HOME` 後備第一次在重載後
  送出 `6` 之後就沒有再收到按鍵而失敗，重跑一次通過。第一次失敗的原因未查明。

收據都在 `workplace/appimage-v.0.2.6-20260929/`（本機）。

## 不宣稱

- 存讀檔與原版的比較只到畫面（住宿問句、存到硬碟、離開問句）；remake 的存檔格式與原版
  CHARDATA 不同，重載之後的續玩只與 remake 自己的收據比較。
- 戰鬥的原版同狀態只有一場正常遭遇（SLEDGE）的六個回合；撤退、側翼、射擊、撿武器的
  敵人行為沒有原版收據（#11）。
- 這一包沒有重新做音訊、其他平台真機與公開發行的驗收。
