# 任務亂數接續後的跨平台回歸

開始：2026-10-07。驗收：2026-10-08。狀態：有限跨平台回歸通過。
正式來源沿[任務亂數驗收](quest-rng-20261007.md)，342份輸入／340份Go保持。
固定seed1，沿既有正常按鍵核騎乘、REST、真正住宿保存與新程序重載，
再以最新任務5前綴及回城保存鍵序核Windows執行。原版收據只重用。

本輪不建立版號、正式封包或發行。Wine不是Windows真機，macOS只做交叉建置及結構核對。
完整戰鬥、HD與全部原完成條件保持。

前輪Go、Wine及macOS映像標籤均無法inspect或啟動；列舉結果不當作可用證據。
已核對原建置來源及版本，再以本專案可重現revision恢復：

| 工具 | 建置來源 | 映像 |
|---|---|---|
| Go／Linux／Windows | [Dockerfile.verify](../../docker/Dockerfile.verify) | `kol-platform-verify:1.26.7-ebiten2.9.9-20261007-r1` |
| Wine9／Xvfb | [Dockerfile.verify-wine](../../docker/Dockerfile.verify-wine) | `kol-wine-verify:ubuntu-noble-20261007-r1` |
| macOS雙架構 | [Dockerfile.osxcross-go](../../docker/Dockerfile.osxcross-go) | `kol-osxcross-go:20261007-r1`，以`KOL_GO_BASE`選上述Go revision |

Go基底固定官方digest，module由現行go.mod／go.sum鎖定，映像保存套件版本清冊。
所有工具只留本機。建置只提供Dockerfile及module檔，不傳原版輸入或素材。
工作目錄為`workplace/platform-quest-rng-20261007/`，來源唯讀、輸出獨立可寫，
執行UID/GID1000:1000、預設無網路，有界CPU／RAM／PIDs及外層逾時。
下載與映像恢復明確使用網路；正式程式、資產及保存格式保持。

| 驗證 | 結果 | 限制 |
|---|---|---|
| 來源與Linux重建 | 342份輸入／340份Go保持；Linux與前輪逐byte相同，十九套件再次通過 | 正式規則與資產未改 |
| Windows建置 | AMD64、PE32+、Go1.26.7、Ebitengine2.9.9與Oto3.4.0符合 | Wine不是真機 |
| macOS建置 | AMD64／ARM64／universal、最低11.0、ARM64簽章及系統依賴通過 | 沒有執行macOS程式 |
| 原正常路徑 | Linux／Wine各雙語2310狀態、404PNG；完整trace、PNG及v3除saved_at符合RE-295 | 原版末端與全部平台閘門保持 |
| 最新任務Windows回歸 | 雙語3888狀態、90PNG；前綴、1045正常鍵回城保存、26新程序重載符合現行Linux；完整v3只差saved_at，保存未改 | 沒有新增原版執行或末端oracle |

本輪共8508個實際執行狀態、898PNG。Windows任務保留原seed1、原正常鍵及自然ORC遭遇，
沒有重擲、注入或強制勝利。原版語意證據沿上一輪，不把本輪重播算成新增原版覆蓋。
工作區、工具準備與控制器皆保存來源及雜湊；必要證據壓縮核對後才清理工作目錄。

正式Linux SHA-256為
`2a885b9a37e884d1830010f7ccb1051b030fe4b14c678fbb71a25955a745fd66`，
Windows SHA-256為`dbc60c4ef9b836ee888018dd764cdd7d6557d79b39db7f5c1090dff56344ca2c`。
全部平台雜湊、映像ID、套件版本與限制見本機[收據](local/platform-quest-rng-20261007/receipt.json)，
[計畫](local/platform-quest-rng-20261007/controller-plan.json)及
[唯讀核對](local/platform-quest-rng-20261007/verify.py)。來源保持，沒有提交、推送、Issue寫入、封包或發行。

日常Go建置、測試與腳本GPU截圖可用本輪Go revision：

```sh
KOL_GO_IMAGE=kol-platform-verify:1.26.7-ebiten2.9.9-20261007-r1 tools/run_kol_shots.sh <規格檔>
KOL_PY_IMAGE=kol-platform-verify:1.26.7-ebiten2.9.9-20261007-r1 tools/run_python_tool.sh <工具與參數>
```

這個Go映像保存目前module cache，支援Linux與Windows CGO=0建置、Xvfb及標準Python核對。
macOS使用上述專用osxcross revision。映像恢復的建置命令、脈絡檔及logs留在壓縮證據，
前版已不存在的工具標籤只作歷史定位。

獨立核對時沿Go revision，專案唯讀掛`/src`、本機證據父目錄唯讀掛`/out`，
UID/GID1000:1000、無網路、1GiB、1CPU、96程序與120秒逾時，執行：

```sh
python3 -B /out/platform-quest-rng-20261007/verify.py
```
