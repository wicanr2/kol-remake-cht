# RE-251：戰鬥指令區的主列重建

日期：2026-10-05。狀態：有限 CONFORMED，完整介面仍開放。

沿「原版 GUI 還原」入口查核完整指令區。文件職責採用
`~/.codex/knowledge-base/local/project-document-responsibilities.md`。
目前正常對拍第175原版畫格對應重製第178步，原版有 OPT 與指令圖示，現行缺漏。

原始輸入 `KOLTOWN-unpacked.EXE`，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
IDA Pro 9.4 線性位址，載入基底0x10000、DGROUP基底0x2DED0。
資料庫 SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`。

既有原版來源：

- `workplace/battle-cursor-20261001/cursor-summary-redraw-private.json` 的 sub_249AE。
- `workplace/battle-panels-20260908/panels-private.json` 的 sub_23638 及 sub_22EFC。
- 正常原版 `workplace/shield-knockback-confirm-20261005/original/normal-six-private.json`。

主列重建於0x249CB、0x249E2、0x249F9，讀目前人物+0x25、+0x23、+0x24並
重載命令、招式、防禦暫存值。原始指令與名稱保留；來源 byte 逐項核對後才提升結論。
目前重製 Planning 清空新選擇用的值，顯示再讀它而漏掉已提交的命令。
隔離候選只改唯讀 CommandPanel 的主列來源，並核對肖像／裝備標籤字型。
不改選擇、提交、亂數或存檔流程。

入口 `workplace/command-panel-lifecycle-20261005/prepare.py`、`build.sh`。
正式來源維持338輸入與Linuxd7faa6c9，等待證據審查。

## 證據審查

原版 sub_249AE 與 sub_23638 全部指令經 MZ 重定位後逐項核對，詳見 evidence-review.json。
主列重載與字型契約為已證實。三次正常1158狀態與39次AI、13次執行事件保持；
213個英文指令區零像素差，候選40個畫格只改左下文字與圖示矩形。
完整畫面105頁只差右下游標；其餘差異及原版操作仍未結案。
模式切換724狀態保留命令圖示與裝備內容，肖像回復整圖相同。
這批沒有新增裝備模式原版oracle，只沿既有已證實來源核對正常重製生命週期。

## 正式驗收與限制

正式三檔變更：CommandPanel 主列來源、BattleField 隊員標籤字型及既有來源回歸測試。
新版Linux SHA-256 `c47c9a29ece5b5394bb89d12116209810c978d3ba404a73a9e8b752acec21638`，與候選位元組相同。
19套件通過，雙語1136正常保存狀態與26新程序重載保持；完整v3只差saved_at。
原版固定seed1與dosgolem968ca774，重製seed1在執行前設定，未重擲。
385等效正常鍵與原版映射沿RE-249／250既有收據，沒有記憶體或座標注入。
原版方向確認、現行ENTER／SPACE及末端隊伍差異維持，未稱操作parity。
三次候選1158狀態、模式724狀態、正式保存與重載1162狀態，總計3044。
新增40個候選畫格只差指令區，其中繁中1張；英文指令區213張全部零差。
完整213張畫面無一整圖相同，105張只差右下游標，其他108張仍有訊息等差異。

本輪兩次準備／建置失敗為腳本語法與候選漏匯入fmt，修正後同一容器工具鏈重跑。
既有兩份IDA匯出欄位不同，首次讀取使用錯欄位，改讀實際形狀後才核對指令。
這些失敗未寫為產品缺陷，也未放寬像素或狀態比較。
字型與視覺檢查包含英文178、繁中230、雙語裝備179；標籤、圖示、箭數20可讀。
沒有新增Windows／macOS真機或封包；HD與其他未決選擇保持待確認。

完整收據與校驗入口為本機 `workplace/command-panel-lifecycle-20261005/receipt.json`、
`verify.py`。素材與原版畫格沿現有本機權利分類，未公開或發布。

後續卷軸背景與成功掉落名稱見[RE-252](20261005-battle-message-background.md)。本文主列與字型證據保持，完整UI仍差滑鼠，不外推原版完整戰鬥。
