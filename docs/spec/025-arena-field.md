# SPEC-025：競技場固定地形

狀態：**READY**（2026-09-08）。不改競技場資格、下注、站位、對手、升階或保存格式。

## 證據與審查

輸入 `KOLTOWN-unpacked.EXE` SHA-256：
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
工具 IDA Pro 9.4；本段位址全為 IDA linear。固定 ZIP SHA-256：
`2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46`。

- **已證實**：`sub_2AAA8:0x2B0C5` 推入 `0x18`，`0x2B0C9` 呼叫 `sub_2B3D1`。
  後者 `0x2B3F8` 加 `0x3D`，得到 payload `0x55`；`0x2B400` 呼叫 archive 載入。
- **已證實**：固定 `TOWN0` 有 86 個 payload，85 號解包後 612 bytes，RLE 解成
  64×64、4096 格。它是第 25 張戰場地圖，既有工具的 24 張只涵蓋任務。
- **已證實**：`sub_2B3D1:0x2B41A` 執行 RLE，`0x2B436` 轉置。
  `TestArenaMapOriginal` 由 dosgolem 執行原版兩段，4096 格與 `battle.LoadField` 一致。
  archive 解包在測試中替換為固定來源資料；不是正常競技場到場驗收。
- RE-079／SPEC-008 的戰場共同 scatter 段仍適用：先載固定地形，再只在地面／粗地
  灑入樹叢；不得把整張圖換成隨機遭遇生成器。

先建立候選及逐格驗證，再審查為 READY。證據輸出：
`workplace/player-gaps-20260908/arena-map.json`、`oracle-rules/player-gaps.log`；
可重生入口 `tools/probe_arena_field.py`、`tools/ida_probe_player_gaps.py`、
`internal/oracle/player_gaps_test.go`。完整 bytes 與原始地址留本機，不提交原版 payload。

## 實作契約

1. 抽取器保留原 24 張地圖編號，再增加 `battle-map-24.png`；它不是第 25 個任務。
2. manifest 記錄 TOWN0、payload 85、原始及 PNG 雜湊、64×64 形狀與既有索引方向。
3. runtime 只讀 PNG，載入失敗或尺寸不符應報錯停止，不能退回隨機地形。
4. 每次開競技場複製固定底圖，再走既有 scatter；上次戰鬥不可污染底圖。
5. 站位及結算沿用現行 READY 規格；本規格不擴充存檔，也不保存戰鬥暫態。
6. 需檢查實際 atlas 範圍、PNG 格值、正常入口地形與原版受控地形；RND 的 scatter
   驗證依雙方固定 seed 鐵則，不要求未受控骰序相同。未完成正常入口前不升 CONFORMED。
7. 競技場沿用已核准的建築對話畫面；繁中模式必須分派到既有 `InnCJK`，
   英文模式使用 `Inn`。這是既有雙語畫面的接線，不新增排版或改動詢問／下注規則。
   肖像 archive 為空時，兩種介面均略過肖像繪製，繼續呈現對話及操作；
   不得以空名稱讀取素材。競技場與賭場的肖像是 COMBPICS 第 6、5 張（`sub_2B454`，
   RE-014 解出圖號與矩形，RE-156 接上）。

## 說明書對照

頁碼是說明書 PDF 實體頁（括號內是印刷頁），連到 `docs/manual/chapters/` 已校對的繁中章節。說明書是公開規則的來源；
執行順序、邊界與隱藏值以本規格引用的原版證據為準，說明書沒寫的部分不從說明書推定。

| 本規格段落 | 說明書 | 說明書寫了什麼 |
|---|---|---|
| 競技場 | [PDF 38（印刷頁 70–71）](../manual/chapters/07-advanced-gamers.md) | 競技場一對一、不能撤退 |
