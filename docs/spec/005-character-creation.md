# SPEC-005：角色建立與角色儲存區

狀態：**DRAFT**  
是否授權實作：**否**

## 範圍

本規格定義 remake 的角色建立狀態機與角色儲存區邊界。它不包含旅店招募、隊伍排序、戰鬥
屬性計算、角色圖像編輯器或原版磁碟交換。證據入口為
`docs/re/character-creation.md`。

## DRAFT 狀態機

`角色選項 → 名字 → 種族 → 性別 → 職業 → 反覆擲屬性 → 接受屬性 → 選圖像 → 原子保存`

- 任何尚未保存的狀態都是 draft；取消不得留下半個角色、占用 slot 或修改既有角色。
- 名字輸入上限為 16 個原版字母／空格單位；繁中 remake 的實際 Unicode 長度與顯示寬度
  另待 UI／存檔 schema 決策，不能把 byte 上限直接冒充字素上限。
- 種族固定為四種；矮人與凱爾登只提供男性。職業選項必須由種族／性別 eligibility table
  產生，不得在 UI 硬編另一份清單。
- 每次 reroll 必須由可注入、可重播 RNG 產生完整候選屬性；接受前不更新持久資料。原版相容
  模式依 RE 已證實的職業 byte 編碼與 bounded-roll 公式運算，但正式 deterministic RNG
  演算法仍須在 READY 前固定，不得把目前工具鏈對 `_rand` 的函式命名當成完整序列證明。
- 接受圖像是唯一 commit 點。角色資料與其 PNG 圖像引用必須同一交易成功；任一寫入失敗
  即全部不保存。
- 角色儲存區上限 16 名；開始遊戲的隊伍上限六名。這兩個限制不得混成同一容量。
- remake runtime 只讀 typed data、PNG 與 OGG；不得在 runtime 解析 `CHARDATA`、
  `CHARPICS` 或任何原版 executable。

## 已知屬性契約

七項主要屬性為力量、敏捷、體型、健康、先見、魅力、智力。每個職業以七個編碼 byte
產生主要屬性：`E + (E & 0x0f)` 次 `1..dice[E >> 4]` 擲骰總和，骰面表固定為
`[3,4,6,8,10,12,20,100]`。平衡、耐力、身體點數依序為力量與智力、力量與健康、
體型與健康的無號整數平均，皆向下取整。手冊職業表仍只是平均值，不能直接當固定初始值、
最小值或最大值；各職業的七個 `E` 尚待完成 enum／記錄對應後才能進 READY typed data。

## 存檔草案限制

正式 schema 必須有穩定角色 ID，不能用顯示名稱或 slot index 作永久身分。原版 768-byte
record 只作 oracle／匯入研究；remake 不承諾 binary compatibility。schema、migration、
Unicode 名字規則與角色圖像引用尚未完成共同決策，本 DRAFT 不固定欄位格式。

## READY gate

1. 以 IDA writer／reader 關閉 35 個職業 enum／16-byte 記錄對應、eligibility 與初始技能／物品；主要屬性公式已關閉。
2. 以正常 DOS UI oracle 抽測名字邊界、每層 ESC、reroll、slot full 與保存 commit。
3. 建立 768-byte record 的最小必要欄位索引；每欄保留原位址、偏移、證據等級與來源。
4. 由使用者決定繁中名字長度／重名 policy 與是否保留原版性別限制；決策前不現代化。
5. 固定 remake typed schema、migration 與 deterministic RNG 契約。
6. 證據審查通過後才可升 READY 並開始 production implementation。
