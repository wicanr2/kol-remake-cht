# 七城 NPC 與坐騎 HD 正式接入

狀態：本切片有限 **CONFORMED**。七城102張NPC與四種坐騎已接入正式開發版，四張既有NPC保留。
本批沿已核准的手繪、1920×1200、紅黑銅框、原有位置與F10，實作前依[SPEC-039](../spec/039-hd-art-ui.md#npc與坐騎正式接入審查)升READY。
現行發行包仍為`v.1.0.2-20261013`，全畫面HD目標仍在進行。

| 項目 | 結果與範圍 |
|---|---|
| 素材 | 27張原樣保存PNG、98個新NPC裁切及四種坐騎；NPC原生最小706×486，涵蓋528×384最大呈現 |
| 完整來源 | 106個不同來源；八組實際數量、來源PNG／RGB／尺寸、裁切RGBA及四張保留圖驗證 |
| 正式場景 | 七城102個進門畫面、四種馬廄與四種人物資料頁，共110場景440畫格 |
| 像素與中文 | EGA RGBA零差；原生圖片、圖外及相對候選完整HD最多一色階；完整Inn與中文紀錄相同，1194個實際顯示字元缺字0 |
| 正常操作回歸 | 標題至Kenneth物品頁31狀態，加既有旅店／資料頁10狀態；每側41狀態、中文與EGA保持，既有HD最多一色階 |
| 遊戲內切換 | 標準／原生NPC、馬廄與資料頁，兩種視窗、鍵鼠共48個F10事件；完整Inn保持，HD往返最多一色階 |
| 保存 | 正常住宿11狀態、新程序重載5狀態及前版產生的v3相容；完整保存只排除saved_at，重載不改來源檔 |
| 程式 | 395份Go／module、242份HD資產；全套及sceneverify通過 |

[同主題比較頁](hd-npc-mounts-runtime-20261015/review.html)保存七組正式畫面，包含橫置頭部、雙人物、藍膚人物、馬廄及坐騎資料頁。
README中的NPC與坐騎比較來自同一受控狀態的重製EGA／HD，沒有原版DOS截圖聲明。
來源、畫格及權利見[截圖清冊](hd-npc-mounts-runtime-20261015/manifest.json)，結果見[正式收據](hd-npc-mounts-runtime-20261015/receipt.json)。

## 載入與呈現

[正式素材清冊](../../assets/hd/first-slice/npc-mounts/manifest.json)使用`kol-hd-npc-mounts/1`。
來源固定為TWN1…7PICS所有偶數圖號2至末圖，以及STEEDS0…3；來源ZIP SHA-256為
`2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46`。
生成工具、28份完整提示、選用圖集及修正說明見[生成清冊](../../assets/hd/first-slice/npc-mounts/generation.json)。

載入先核對完整來源與PNG，再建立呈現快取。缺件、錯誤圖號、重複、錯來源、越界、透明裁切、尺寸不足或保留圖遭改動會拒絕整組。
首次繪圖另核對GPU來源RGB與尺寸。F10沿既有原子切換，失敗保留原呈現；runtime只讀PNG與清冊，不讀原版archive。

NPC沿標準對話及原生服務頁的既有矩形與描邊。馬匹完整等比置中於原馬廄／資料頁，保留黑底，不加NPC描邊。
商品、價格、移動率、對話、規則、原始資料及kol-save-v3保持。
原生服務頁另以528×384原圖獨立取樣，排除既有兩像素框線後最多一色階。

## 驗證與重生

- [正式來源](hd-npc-mounts-runtime-20261015/source-inputs.json)與[HD資產](hd-npc-mounts-runtime-20261015/asset-inputs.json)記錄實際輸入SHA-256。
- [完整場景](hd-npc-mounts-runtime-20261015/scene-check.json)、[原生裁切](hd-npc-mounts-runtime-20261015/native-check.json)與[擷取紀錄](hd-npc-mounts-runtime-20261015/scene-captures.json)分開保存。
- [F10](hd-npc-mounts-runtime-20261015/gui-check.json)、[正常操作](hd-npc-mounts-runtime-20261015/normal-check.json)及[保存](hd-npc-mounts-runtime-20261015/save-check.json)各列自身範圍。
- [全套日誌](hd-npc-mounts-runtime-20261015/tests.log)及[sceneverify](hd-npc-mounts-runtime-20261015/sceneverify.log)通過。20項素材負例、六項UI原子掛接負例與實際來源錯誤均拒絕。

重生沿[證據保存入口](retained-evidence.md#七城npc與坐騎hd正式接入)，在既有鎖版Docker image中還原`hd-npc-mounts-20261015`。
完整工具保留原工作區檔名：[建置](hd-npc-mounts-runtime-20261015/build.sh)重跑全套及實際畫廊；[比較器](hd-npc-mounts-runtime-20261015/check.py)核對正式與候選。
完整模型、商店／資料頁樣本、v3及GUI狀態只留本機；正式Git收據只保存結果、來源雜湊與已驗PNG。
首次隔離資料連結缺漏、兩個未改到目標的負例、GUI流程缺漏及啟動逾時均保留，修正後以同工具鏈重跑。

本結論限Linux Docker與上述場景。沒有新增原版oracle、自然完整通關、跨平台真機或新版Release驗收。
怪物、城鎮／世界地圖、戰場、傷勢、天候、片頭及其餘事件與未訪操作分支持續，見[13類覆蓋](hd-completion-20261014.md#目前覆蓋)。
