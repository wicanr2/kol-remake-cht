# 正常戰鬥入口的隊員數值比較

日期：2026-09-08。兩名隊員共 40 個初始欄位通過；完整戰鬥同狀態仍未完成。

## 原版來源

dosgolem 重新走正常選角 Olear、招募 Theodoric、離店、出城、露營及次日 SLEDGE
伏擊，確認訊息後到達第一名隊員規劃入口。`TestNormalBookEntry` 119.47 秒通過。
旅店 seed=1，正常規劃入口再設定共同測試 seed=1；未注入原版隊員、站位或遭遇。

輸入為 `KOLTOWN-unpacked.EXE`，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
dosgolem 修訂為 `fd746405cf0ef9e535121e215b935b0bf59cd1de`，未提交來源逐檔清單
保存在本次原版 `manifest.json`，不以修訂號單獨代表執行來源。

原版輸出：`workplace/battle-state-records-20260908/`。
`battle-alignment.json` 保留 50-byte 參戰記錄；`battle-state-records.json` 增加角色原始記錄、
怪物記錄片段及耐力／疲勞。地址為執行期 DGROUP 相對位移，實際段值 `0x1EFD`。
依 RE-080／129，角色 DS:BB88、步幅 0x300；參戰者 DS:6712、步幅 50；
耐力 DS:1C50、疲勞 DS:322E 各為每槽一個 word。未解 byte 原樣保留，不附推測語意。

## 比較與結果

`tools/prepare_battle_state_fixture.py` 使用既有抽取器的 `character_sheet`、`equipment`
及 `worn_slot` 解析本次原版角色資料；再由正式 `townSheet`、`townGood`、`townWorn`、
`battleMember` 及 `battle.NewQuest` 計算參戰者。預期值取自另一份原版參戰記錄及
獨立耐力／疲勞陣列，不使用 remake 公式產生。

`TestOriginalNormalPartyState` 通過。每人比較 20 欄：種類、陣營、等級、力量、快捷、
洞察、體型、智力、累積傷害、昏迷門檻、六個部位、兩軸位置、耐力與疲勞。
Olear 耐力 183、Theodoric 耐力 230，兩者疲勞皆為 0。
這是受控 remake 建構測試，不是正常 remake GUI 戰鬥驗收。

輸出：`workplace/battle-party-state-20260908/`，包含 typed 輸入、比較紀錄、
實際受測來源及雜湊收據。沒有修改正式規則或重建 AppImage。

## 未完成與勘誤

完整裝備能力、怪物狀態、地形、規劃、指令／攻擊結算及正常戰後保存鏈仍待驗。
怪物記錄目前只擷取 DS:CD92 起 0x80 byte，不能當成全部穿戴欄已完整保存。

本次正常原版 DS:D39A 為 3。依 RE-097 對應 IDA 線性 `word_3B26A`，
基底 `0x2DED0`，相對位移為 `0xD39A`。這證實本次 SLEDGE 遭遇等級為 3，
否定 SPEC-008 早期「野外戰鬥恆為 1」概括，但不憑單次觀察推導全部等級分布。
## 同日追加：SLEDGE 初始數值

沿用本次 dosgolem 正常快照，先以怪物原始記錄核對 typed SLEDGE 的六項屬性及
手上武器完整 16-byte 記錄，再將 typed 怪物與原版實際等級 3 交給正式 `NewQuest`。
兩名隊員及 SLEDGE 合計 59 個初始欄位比較通過；新結果位於
`workplace/battle-monster-state-20260908/comparison.log`，輸入在同目錄 `party-fixture.json`。

怪物 Kind 的原版值與 typed 資料表索引不同，映射尚待證實，因此該欄明列未比較；
不能把本次結果解讀為全部怪物狀態一致。完整裝備能力、地形、規劃及攻擊結算仍待驗。
本次沒有重新執行 dosgolem，而是擴充既有正常收據的 remake 側比較；原版收據不改寫。

## 同日追加：補回怪物遠程能力欄位

SPEC-029 證實參戰記錄 +0 為怪物遠程能力後，移除上述 Kind 跳過條件。
抽取器同時核對原版怪物記錄 +0x1E／+0x1F 與資料包的機率／骰碼。
SLEDGE 兩欄皆為零；以固定 seed=1 呼叫正式初始化，再將規劃亂數重設為共同
seed=1，兩名隊員與 SLEDGE 合計 **60 個初始欄位通過**。
初始化與規劃使用明確分離的種子邊界，不宣稱整段自然遭遇亂數流一致。

輸出：`workplace/battle-monster-complete-initial-20260908/`，包含重新產生的
`party-fixture.json`、完整 `tests.log` 與建置產物。沿用既有 dosgolem 原版收據，
本次未重新執行原版。完整裝備、地形、規劃、攻擊結算及正常 GUI 戰後保存仍待驗。
