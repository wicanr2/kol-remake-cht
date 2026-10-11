# 首批 HD 遊戲接入驗收

狀態：**首批通過**。使用者於2026-10-11核准既有樣張並要求開始，
[SPEC-039首批契約](../spec/039-hd-art-ui.md#首批接入契約)由READY升有限CONFORMED。
本次驗收沿[SPEC-040](../spec/040-controlled-scene-acceptance.md)，不新增原版嚴格對拍聲明。

開發版以 `-graphics hd -lang zh-Hant` 啟用，畫面為1920×1200。
預設EGA與既有v1.0.0封包沿原版本；本次只接入核准的旅店門面與六人圖集。
門面與角色依原圖RGB像素SHA-256匹配，未知或自訂圖回退原圖；保存格式仍為kol-save-v3。
素材來源、完整生成提示及雜湊沿[核准原型](hd-first-slice-20261011/manifest.json)，
接入清冊為[manifest.json](../../assets/hd/first-slice/manifest.json)。

本批測試程式雜湊保存在收據；[build/kol-hd](../../build/kol-hd)目前為包含
[遊戲內切換](display-switch-runtime-20261012.md)的最新開發版。
可在專案根目錄以`build/kol-hd -graphics hd -lang zh-Hant`遊玩。
它是開發中間產物；正式交付入口仍為README所列的`dist-all/`。
重建沿[go_docker.sh](../../tools/go_docker.sh)：

```sh
KOL_GO_IMAGE=kol-package:1.26.7-ebiten2.9.9-20261011-r1 \
  tools/go_docker.sh build -o build/kol-hd ./cmd/kol
```

| 驗證 | 結果 |
|---|---|
| 書本至城鎮 | 27狀態，正常建角與選隊路徑 |
| 真實六人與門面 | 5狀態，逐格顯示各自匹配的HD肖像；旅店門面正確 |
| 資料頁 | 10狀態，沿既有鏡子操作腳本 |
| 自訂肖像 | 3狀態，v3第一人改為合法760-byte單點圖，該格原圖保持 |
| 既有v3 | 2狀態，直接讀取現行交付保存副本 |
| 住宿與新程序讀檔 | 11＋7狀態，沿正常旅店保存；修改前、EGA及HD的v3除saved_at相同 |
| 受控戰鬥至結局 | 33狀態，重用固定seed1、鎖血及20倍攻擊的驗收輸入；資訊框按實際選中人物替換 |
| 狀態與中文 | EGA與HD各98個狀態及文字紀錄完全相同；修改前65張EGA雜湊相同 |
| 畫面位置 | HD與EGA最近鄰乘二比較，差異只在核准的六人圖片、門面及戰鬥資訊圖區；姓名、外框與操作列保持 |
| 真實滑鼠 | Xvfb＋xdotool，1920×1200點進門並選第二人；縮至960×600後選第一人，實際token與命中位置相符 |
| 守門測試 | 原正式套件及sceneverify戰鬥／適配器通過；32組命中邊界、缺檔、雜湊損壞、尺寸與重複人物指紋反例通過 |

| 實際六人與門面 | 實際戰鬥人物 |
|---|---|
| ![六人門面](hd-runtime-20261011/inn-door.png) | ![戰鬥資訊框](hd-runtime-20261011/battle.png) |

目視確認六人、門面、戰鬥資訊框、資料頁、自訂圖及實際互動畫面；中文與圖框無新增覆蓋。
人物姓名沿保存內容，頂列沿原契約縮寫。共保存9張代表PNG，來源、畫格與權利分類見
[PNG清冊](hd-runtime-20261011/manifest.json)。

工具為[build_hd_first_slice.py](../../tools/build_hd_first_slice.py)與
[verify_hd_first_slice.py](../../tools/verify_hd_first_slice.py)。後者在Docker內依序執行
`prepare`、`captures`、`gui`、`seal`，專案掛`/src`、本輪工作輸出掛`/out`。
`kol-baseline`由Git `8a6b52da`的cmd、internal及module檔重建；`kol`為本輪正式來源，
`kol-sceneverify`只用於受控戰鬥。種子在執行前固定為1，不重擲或挑結果。
環境為`kol-package:1.26.7-ebiten2.9.9-20261011-r1`、Go1.26.7、Ebitengine2.9.9、Pillow及Xvfb。
命令、輸入、程式及工具雜湊見[收據](hd-runtime-20261011/receipt.json)。

完整紀錄包含提示層修正前畫格與輸入／自動化錯誤，按[保存流程](retained-evidence.md)
存至[本機清冊](local/hd-runtime-20261011/retention.json)，核對後移除本輪workplace。
全部圖片只限目前private GitHub。其他美術、人物裝備組合、傷勢遮罩及平台封包仍由
[HD工作項](../../WORKLIST.md)追蹤，本收據只確認首批範圍。
