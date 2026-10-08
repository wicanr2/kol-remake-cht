# RE-246：新版任務操作腳本續核

日期：2026-10-05。狀態：明示遭遇控制的新版任務、旅店收尾、真正保存與新程序重載通過。

## 來源與條件

- 正式程式沿 RE-245，336 Go／338 輸入保持。Linux SHA-256：
  `27212080a6a8bfd047c9250cbefb20da8a4a9ca7237a0934bb557053b6a16b8a`。
- 固定 image：`eob-remake-go:1.26.7-ebiten2.9.9-audio`，Go 1.26.7；image ID：
  `sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa`。
- seed 1 在執行前固定，途中不重播種。沿用 200 個明示遭遇控制骰 100。
  沒有讀取預置存檔、替換角色、座標、任務地點、怪物、戰局或強制結果。
- 可丟棄輸入產生器讀目前狀態，再呼叫原有正常按鍵分派。它不改戰鬥狀態。
  產生階段不畫每個鍵，另以未修改的正式程式逐鍵繪圖重播，才列為驗收。
- 怪物 v2、人物 v3、音訊、美術、UI、版號及公開方式保持。

## 腳本失配與修復

已保存事件顯示，舊腳本第 881 步的出手索引首次不同。第 888 步舊版已進第五回合，
新版仍在第四回合結算；第 890 步新版才進第五回合。第 1325 步開始落到不同階段。
因此舊按鍵在後面選單失去對應，最終提早答應逃跑。
此結論來自 RE-245 的既有唯讀事件與 RE-243 正常 trace，不將腳本失配判成新產品缺陷。

輸入產生器 V1 指定 Moraine 運送並依疲勞休息，5421 鍵、80 回合戰敗，未取得物品。
V2 沿舊版最近隊員策略，第 4781 步、69 回合由 Fighting Bull 取物，5411 鍵、96 回合戰敗。
V3 保留 V2 取物前綴，返程改走避開敵人相鄰格的路線，疲勞時選正常 REST。
同一 seed 與前綴完成任務；沒有改骰序或重擲挑結果。

V3 的 5969 鍵已實際保存，但末圖仍停在 Kenneth 的住宿價錢問句。
V4 保留全部 5969 鍵，只追加 RET、SPACE，完成缺錢提示並回到旅店一般選單。
前兩個戰敗與 5969 鍵基準均保留，未算成完整住宿佇列通過。

## 正式重播結果

| 項目 | 結果 |
|---|---|
| 正常起點 | 標題、書本、六人招募、接任務、正常世界行走 |
| 取物 | 第 4781 步，第 69 回合，Fighting Bull 正常 DROP／PICK UP |
| 逃離 | 第 109 回合走到戰場邊緣，第 5582 步確認，第 5583 步回世界 |
| 回報 | 六人各一次；Treasure Map 保留完整記錄 `ca48a19d64503430000a000000000000` |
| 住宿 | Olear 與 Fighting Bull 各付 56，四人缺錢正常略過，佇列收完 |
| 真正保存 | 人物 v3，六人在檔，位置為城鎮 7、入口 3、57／8 |
| 正常狀態 | 英文與繁中各 5972，共 11944；產生器與正式逐鍵狀態相同 |
| 新程序重載 | 各 13，共 26；重載不改存檔 SHA-256 |
| 雙語比較 | 只排除四種顯示文字，全部其他軌跡一致；完整保存只差 saved_at |

保存後金冠依隊伍次序為 0、0、1721、0、386、0；傷勢為 3、1、26、1、1、10。
取得物品、邊緣逃走、旅店完成及新程序人物頁另作目視核對。
英文來源畫布為 320×200，繁中為 960×600，沿既有 canvasFor／book.CJKScale=3 契約。
1280×800 是 Xvfb 桌面，沒有將它誤當來源 PNG 尺寸。
本輪未改產品碼，沿用 RE-245 十九套件結果，不無變更重跑全量 Go。

## 現行平台候選

同一 338 來源離線重建 Windows AMD64、macOS AMD64／ARM64／universal。
Linux 沿正式執行檔的逐 byte 身分，沒有無變更重建。
PE32+／Mach-O 架構、universal 內含兩份 thin、最低 macOS 11.0、系統相依及音訊依賴通過。
全部執行檔完整 SHA-256 見 `platform-build-checks.json`。

Wine 9 預設 DirectX 從標題正常招募六人，自然遇敵，雙語各 386 狀態走完十三回合戰敗。
另由各自新程序載入本輪真正寫出的 v3，雙語各 13 狀態；完整 trace 與 22 張 PNG 與 Linux 相同。
兩組條件都在執行前固定 seed 1，未套遭遇控制骰；重載不改原保存檔。
Wine 不等於 Windows 真機，macOS 本輪只驗建置與結構；沒有正式封包或 Release。

## 入口與限制

工作區：`workplace/quest-script-refresh-20261005/`。

- `analyze.py`、`existing-events-review.json`：舊操作最早偏離。
- `prepare-driver.py`、`build-driver.sh`、`driver-main.go`：隔離輸入產生器。
- `generate-v1.py` 至 `generate-v3.py`、`generate.py`、`strategy.py`、`generation-v1/` 至 `generation-v4/`：
  固定前綴、策略失敗及正常按鍵計畫。
- `adapted-plan-5969.json`、`formal-en/`、`formal-zh-Hant/`：保存已寫出但住宿佇列尚未收完的基準。
- `adapted-plan.json`、`run-complete.py`、`run-complete.sh`、`complete-en/`、`complete-zh-Hant/`：
  正式逐鍵重播、真正存檔與新程序重載。
- `review-complete.py`、`review.json`、`png-manifest.json`：完整比較與本機圖片清冊。
- `prepare-platform.py`、`build-windows.sh`、`mac-build.sh`、`check-platform.py`：現行平台來源與候選。
- `run-wine.sh`、`run-wine.py`、`wine-review.json`：Wine 正常十三回合及真實 v3 重載。
- `seal.py`、`verify.py`、`receipt.json`、`audit.json`：來源、附件雜湊與衛生驗證。

這是明示控制遭遇的重製版回歸。沒有新增原版整條任務 parity、自然完整旅程或結局驗收。
原版十三回合戰敗後確認、世界返回與保存仍待驗。平台真機與封包尚未補驗。
HD 手繪媒材已確認，比例及正式美術仍待已提出的原型決策。
所有候選、圖片與原版衍生資料保持本機，沒有 Git 提交／推送、Issue 寫入、定版或發布。
