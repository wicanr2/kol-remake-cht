# SPEC-005：角色建立與角色儲存區

狀態：**READY**  
是否授權實作：**是**（既有已審查流程；2026-10-01 新證實的滿槽／重名／字元差異須先完成下述決策與規格審查）

2026-09-08 現況訂正：上句的「未驗證」限原版滿槽行為；重製版的 16 槽全滿、
刪除及重用已由 [完整保存整合](../verification/complete-save-integration-20260908.md)
正常 GUI 驗證。身分、v2 schema、v1 匯入與建角原子回填由
[SPEC-023](023-complete-save.md) 補充並優先適用；下方早期 schema 待定不再阻止已審查保存。
圖像／盾牌編輯器另立 SPEC-036，不算入已完成建角。

## 範圍

本規格定義 remake 的角色建立狀態機與角色儲存區邊界。它不包含旅店招募、隊伍排序、戰鬥
屬性計算、角色圖像編輯器或原版磁碟交換；角色刪除屬於儲存區邊界，含在本規格內。證據入口為
`docs/re/character-creation.md`。

## 建角狀態機

`角色選項 → 名字 → 種族 → 性別 → 職業 → 反覆擲屬性 → 接受屬性 → 選圖像 → 原子保存`

- 任何尚未保存的狀態都是 draft；取消不得留下半個角色、占用 slot 或修改既有角色。
- 名字輸入上限為 16 個原版字母／空格單位。原版 DOS UI 在輸入第 16 個字元時直接前進到
  種族層、不需 Return，且空名按 Return 不前進；remake 是否沿用「填滿即前進」屬於操作
  取捨，需與使用者共同決策，不得預設沿用或預設移除。繁中 remake 的實際 Unicode 長度與
  顯示寬度另待 UI／存檔 schema 決策，不能把 byte 上限直接冒充字素上限。
- 種族固定為四種；矮人與凱爾登只提供男性。職業選項必須由種族／性別 eligibility table
  產生，不得在 UI 硬編另一份清單。
- 每次 reroll 必須由可注入、可重播 RNG 產生完整候選屬性；接受前不更新持久資料。原版相容
  模式依 RE 已證實的職業記錄與有界擲骰公式運算，但正式 deterministic RNG
  演算法仍須在 READY 前固定，不得把目前工具鏈對 `_rand` 的函式命名當成完整序列證明。
- 接受圖像是唯一 commit 點。角色資料與其 PNG 圖像引用必須同一交易成功；任一寫入失敗
  即全部不保存。
- 角色儲存區上限 16 名；開始遊戲的隊伍上限六名。這兩個限制不得混成同一容量。
- 原版已證實的返回關係為：種族→名字、secondary→種族、職業→上一個適用選擇層、
  屬性→職業、圖像→commit 前的職業流程；正常 DOS UI 收據另證實取消回到名字層時，
  名字與摘要欄位都會清空。滿 16 槽時不得開始修改記錄。空名不前進已由正常 DOS UI
  收據證實。
- 圖像 selector 已由正常 DOS UI 收據證實恰好循環 80 張：由 0 向左一次等於由 0 向右
  80 次，再向右一次回到 0。接受圖像後畫面回到角色選項，且此時才寫入角色檔。
- 相容流程的圖像 selector 固定為 80 筆，方向鍵循環、Return 接受；typed PNG 引用與原版
  selector 的對照仍須在 READY 前固定。
- remake runtime 只讀 typed data、PNG 與 OGG；不得在 runtime 解析 `CHARDATA`、
  `CHARPICS` 或任何原版 executable。

## 已知屬性契約

七項主要屬性為力量、敏捷、體型、健康、先見、魅力、智力。每筆 16-byte 職業記錄以
offset `0..6` 保存骰式 `D`、offset `8..14` 保存內部基值 `B`；每項內部結果為
`B + (D & 0x0f)` 次 `1..dice[D >> 4]` 擲骰總和，骰面表固定為
`[3,4,6,8,10,12,20,100]`。平衡、耐力、身體點數依序為力量與智力、力量與健康、
體型與健康的內部無號整數平均，皆向下取整；畫面再顯示
`floor(內部值 × 100 / 150)`。手冊職業表仍只是玩家量尺的平均值，不能直接當固定初始值、
最小值或最大值。39 筆記錄的 selector 分組、race／secondary enum 與逐名稱順序已由文字
資源 consumer 關閉；raw 記錄仍只作 RE 輸入，需待武器以外技能／物品的存在性、必要
reader 與存檔 schema 一起審查後，才可建立 READY typed data。

起始金冠、主要／次要武器及其兩個武器技能值已由 39 筆 executable metadata 與中文手冊
逐列交叉驗證。這只關閉起始武器垂直鏈；不得推廣成所有一般技能、物品或裝備欄位均已知。
最小 CHARDATA writer offset、CHARPICS／SHLDPICS 同槽寫入與三檔保存順序已固定，但原版
跨檔案 atomicity／rollback 未知。remake 的原子保存是設計要求，不是原版 parity 宣稱。
完整 writer coverage 已證實：slot 清零後，所有直接 CHARDATA 寫入都落在固定欄位或
REGION metadata route，沒有其他職業衍生的起始技能／物品。KOLTOWN 亦已讀回名字、種族、
職業、金冠與主要武器，足以關閉本規格的最小 reader／writer 欄位 gate；未列欄位仍不命名。

## 已由正常 DOS UI 收據關閉的邊界

| 行為 | 收據 |
|---|---|
| 名字欄容量 16 字元；輸入第 16 個字元直接前進種族層 | `character-ui/name-length-run{1,2}` |
| 空名按 Return 停在名字層 | `character-ui/blank-name-run{1,2}` |
| 各層 ESC 的返回目的地與名字層清空 | `character-ui/escape-layers-run{1,2}` |
| REROLL STATS 重擲九個顯示值 | `character-ui/reroll-run{1,2}` |
| 圖像 selector 恰好 80 張循環 | `character-ui/picture-cycle-run{1,2}` |
| 接受圖像後才寫入，且只寫單一槽位 | `character-ui/accept-and-save-run{1,2}`、`asset-parity/chardata-write-run{1,2}` |
| 刪除角色用角色清單頁選取，確認視窗只認 `y`／`n` | `character-ui/delete-character-run{1,2}` |
| 刪除只把槽位名字的第一個 byte 寫成 0，其餘欄位與兩個圖像檔不動 | `character-ui/delete-character-confirm-run{1,2}`、`chardata-delete/` |

刪除是角色儲存區的第二個 commit 點，和建立同樣是原子的：回答 `n` 不寫任何 byte。
畫面上名冊往上補齊是「只列出已建立槽位」的結果，不是資料搬移——remake 的槽位身分
因此不隨刪除改變。

2026-10-01 訂正：原版的滿槽、完全相同名字與可列印 ASCII 已補齊：
[滿槽收據](../verification/character-ui/roster-boundaries-20261001/full-receipt.json)、
[重名收據](../verification/character-ui/roster-boundaries-20261001/duplicate-receipt.json)、
[字元收據](../verification/character-ui/roster-boundaries-20261001/charset-receipt.json)。

- 滿槽仍可進建角草稿，接受肖像才顯示 `NO MORE ROOM ON THIS DISK!`，不寫角色檔。
- 完全相同的 Theodoric 在接受肖像直接取代既有第 0 槽，不另占空槽。
- ASCII 32…126 中字母與空白 53 項接受、其餘 42 項拒絕；MiXeD Name 實際保存保留大小寫與空白。延伸 ASCII／Unicode 未測。

remake 目前提早拒絕滿槽、拒絕重名、正式鍵盤接受數字並固定大寫，本規格整體
仍 READY，不能升 CONFORMED。重名涉及既有角色重設，已提出「拒絕／確認後覆寫／
照原版覆寫」，尚待回答；名字政策及滿槽時點亦須規格審查後實作驗證，保存身分
仍依 SPEC-023。刪除確認按 ESC 與空名冊進刪除畫面仍未知。

重生入口為 `tools/oracle/check_book_roster.sh`／`book_roster_test.go`；僅在 Docker
內掛 `/src` 本專案唯讀、`/repo` dosgolem 唯讀、`/out-root` 本機輸出可寫，以
`KOL_ROSTER_ATTEMPT` 指定尚不存在的輸出名。固定 dosgolem `968ca774…ac71f`、
`kol-package:20260907`、目前 UID/GID 與有界逾時，需本機合法原版
`workplace/oracle/game/`。同一容器以 `python3 tools/verify_book_roster.py` 核實收據。
證據與限制見 [RE-169](../re/character-creation.md#2026-10-01-名冊邊界正常-ui-訂正re-169)。

## 保存契約與早期草案訂正

正式 schema 必須有穩定角色 ID，不能用顯示名稱或 slot index 作永久身分。原版 768-byte
record 只作 oracle／匯入研究；remake 不承諾 binary compatibility。
schema、migration 與既有圖像引用以 SPEC-023 READY 為準，已接正式 v2 與 v1 匯入。
此訂正不擴張 Unicode 名字政策，也不核准尚未有保存契約的圖像／盾牌編輯器。

## READY gate

1. [x] 起始金冠、武器與武器技能已關閉；writer coverage 排除其他職業衍生起始資料。
2. 靜態控制流已關閉首 byte 零分支、每層取消、reroll、80 張圖像循環與 slot full 不進
   writer；仍須以正常 DOS UI oracle 抽測空名、15／16 字元、重名、畫面返回、slot full
   與保存 commit。
3. [x] 最小 writer 欄位索引與 KOLTOWN 必要 reader 已交叉查核；未知 offset 不猜補。
4. 由使用者決定繁中名字長度／重名 policy 與是否保留原版性別限制；決策前不現代化。
5. 固定 remake typed schema、migration 與 deterministic RNG 契約。
6. 證據審查通過後才可升 READY 並開始 production implementation。

## 說明書對照

頁碼是說明書 PDF 實體頁（括號內是印刷頁），連到 `docs/manual/chapters/` 已校對的繁中章節。說明書是公開規則的來源；
執行順序、邊界與隱藏值以本規格引用的原版證據為準，說明書沒寫的部分不從說明書推定。

| 本規格段落 | 說明書 | 說明書寫了什麼 |
|---|---|---|
| 建角狀態機 | [PDF 4（印刷頁 2–3）](../manual/chapters/01-getting-started.md) | 開始遊戲、建立角色：輸入名字，選種族、性別、職業與圖像 |
| 名冊與刪除 | [PDF 5（印刷頁 4–5）](../manual/chapters/01-getting-started.md) | 角色清單最多 16 名、位置文字為旅店名前十字母加城鎮三字母縮寫；刪除要再確認，選錯可按 No 返回 |
| 已知屬性契約 | [PDF 27（印刷頁 48–53）](../manual/chapters/07-advanced-gamers.md)、[PDF 30（印刷頁 54–57）](../manual/chapters/07-advanced-gamers.md) | 種族與性別限制職業、職業屬性範圍表、各屬性與次要屬性的意義 |
| 圖像／盾牌編輯 | [PDF 5（印刷頁 4–5）](../manual/chapters/01-getting-started.md) | 編輯圖像或盾牌的用途；操作細節見參考卡第 2–3 頁（Apple II 版，不直接套用 DOS） |

說明書沒寫：建角逐層取消、滿槽拒絕、名字欄字元集合與存檔欄位，這些以本規格引用的 KOLBOOK 反組譯與 UI oracle 收據為準。
