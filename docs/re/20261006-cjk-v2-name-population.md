# RE-261：怪物 v2 四套裝備的繁中名稱母體

日期：2026-10-06。狀態：現行資料母體與正式字型量測通過，正式程式保持。

既有最長怪物名與戰利品已有正常玩家畫格。RE-260 後核對剩餘清單時，
確認舊 316 筆戰利品母體尚未涵蓋正式怪物 v2 的四套裝備。
本輪只補完整名稱、字模與字寬核對，不重新開啟已完成的正常選樣。

## 現行母體

| 項目 | 核對結果 |
|---|---|
| 怪物名稱 | 32 種，最長「岩石食人魔」80 像素 |
| 裝備組 | 32 × 4 = 128 套，每套六槽 |
| 裝備槽 | 768 槽，其中 625 槽非空 |
| 五份來源的物品記錄 | 1435 筆，排除一筆獨立法術槽後為 1434 筆 |
| 戰利品候選 | 475 筆，最長值維持 80 像素 |
| 實際 Go 字寬 | 1132 筆、129 個不同顯示名稱，全部相符 |
| 缺譯、缺字 | 0、0 |

五份來源為 `items.json`、`town-buildings.json`、`character-roster.json`、
`quest-sites.json` 與 `monsters.json`，均位於 `assets/original-derived-data/`。
怪物資料必須為 `kol-monster-data-v2`，逐種核四套、每套六筆 16-byte 記錄，
以及非空 decoded item 與原始 record 相符。
保留 `monsters.json/monsters/<索引>/equipment_variants/<組>/items/<槽>` 定位；
`variant_rows.offset` 是原 archive payload 的怪物記錄內位移，並非 IDA 線性位址。
本輪沒有新增原版執行或逆向語意。

候選判準沿既有母體：生成物品來源、奇數 kind 的 TAKE 候選，以及 InHand。
法術與坐騎維持獨立分類。這是顯示上界，不保證每一件均能在同一場戰鬥取得。
所有名稱經既有 TSV 查表，原名與物品記錄不翻譯。

舊 316 筆母體中的 53 筆 `worn` 來源已搬到 v2 的裝備組欄位。
475 = 316 − 53 + 212；53 筆原名與完整 bytes 都仍在新母體。
212 筆是四組候選的來源定位數，不稱為新增 212 件不同物品。
實際增減及完整記錄見 `population-delta.json`。

## 正式字型與既有正常畫格

沿用 [既有 Go 量測器](../verification/cjk-ui/font-width-20261001/measure.go.txt)，
只改三個輸入、輸出路徑。以現行 `internal/font.Font.Load`／`Measure`，
讀正式 `zh-hant-16` PNG 與 manifest；1132 筆全部等於逐字字距量測。
32 個怪物名、625 個非空裝備槽及 475 個候選均包含在量測內，重複名稱保留來源。

另以現行字模回查兩張已保存的正常畫格：

| 既有畫格 | 字串與實際字模矩形 |
|---|---|
| `cjk-monster-loot-20261003/formal-inspect-zh-Hant/step-2745.png` | 岩石食人魔，364,51,80,18 |
| `cjk-longest-victory-20261004/formal-zh-Hant/step-565.png` | 棘刺釘頭鎚，24,57,80,18 |

兩張圖均在 `workplace/` 下。字模完整相符，缺一墨點與 79 像素縮窄反例拒絕。
這只證明既有正常畫格與現行字模相容；沒有宣稱舊流程以現行正式版重跑。
母體最長值維持，因此既有正常選樣的選擇理由保持有效。
各件裝備的正常取得、全部畫面、完整戰鬥與新保存鏈維持原驗收範圍。

正式 338 個輸入逐檔保持，Linux SHA-256 為
`1f31007c19ceac875a12df9b68d8629296764d01194983cf98d10cc8dcdbd274`。
本輪沒有遊戲重播、狀態注入、亂數操作或正式測試重跑。
既有十九套件結果沿 RE-258，完整玩家路徑證據沿 RE-259／260 的實際範圍。

## 工具失敗與重現

母體工具首稿誤把全零當作空槽。實際 `tools/extract_monsters.py:item` 的
既有契約是首 byte 為 0，即使其餘 bytes 有殘留也為空槽。
拒絕稿與反例保存於 `population-v1-rejected.py`／`null-contract-rejected.json`，
不修改原始記錄或判定原版資料損壞。

第二稿在排除法術前檢查譯名，誤把 Warder 的 `Vartwelyrmi` 當缺譯。
核對 SPEC-033 與路由後，沿既有法術獨立分類修正順序。
`population-v2-rejected.py`／`filter-contract-rejected.json` 保存失敗與來源。
兩稿都在 Go 量測或遊戲執行前停止，未改產品。

工作區為 `workplace/cjk-v2-name-population-20261006/`。
入口 `population.py`、`measure.sh`、`review.py`；`population.json` 保存來源 SHA-256、
名稱與原始定位，`runtime-widths.json` 是正式 Go 量測輸出。
`seal.py`／`verify.py` 保存並重算附件、來源、名稱母體、字寬、既有字模、
文件索引、檔案擁有權與容器清理。實際核對結果見 `receipt.json`／`audit.json`。
固定 Docker、UID/GID 1000、有界資源、network none、Xvfb trap 與 --rm。
資料與研究輸出只留本機，沒有新提交、推送、Issue 寫入、版號、封包或發布。

本輪只解除 v2 名稱母體與字寬的缺項。父規格、全部中文畫面、受阻規則 DRAFT、
HD 解析度與介面、平台真機及正式交付維持既有閘門。
