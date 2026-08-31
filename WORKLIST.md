# 《武士傳說》重製工作清單

更新日期：2026-08-31

## 目前狀態

| 分類 | 狀態 | 目前證據／限制 |
|---|---|---|
| remake | 尚未開始 | RE 知識庫已建立；SPEC-001 READY 只授權研究 oracle，SPEC-002 DRAFT 不授權 Go implementation |
| 原版 oracle | build 已固定 | ZIP 與全部項目逐檔雜湊已固定；發行來源／1998 timestamp 成因未知 |
| 玩家觀點 | 起源資料完成 | 魔旅士繁中個人心得＋《CRPG 通鑑》中文長篇評介已進 README |
| 可選美化 | 未決策 | 不得在忠實度與操作方向確認前進入 production |
| GitHub | 已建立 | private：`wicanr2/kol-remake-cht`；`main` 已推送 |

## M0：專案建立

- [x] 撰寫 `AGENTS.md`，固定 Docker-only、證據、Go／Ebitengine、PNG-only runtime、
  OGG、中文說明書、授權與 `dist-all/` 契約。
- [x] 建立本工作清單與 README 起源骨架。
- [x] 初始化 Git，確認原版封裝與衍生素材由 `.gitignore` 排除。
- [x] 建立 `wicanr2/kol-remake-cht` private GitHub 儲存庫、設定 `origin` 並複核 visibility。
- [x] 提交並推送專案骨架；在 `WORKLOG.md` 記錄 commit、push 與 Docker 清理狀態。

完成條件：private 遠端可查、工作樹無誤納原版素材、`AGENTS.md` 與本清單已推送。

## M1：鎖定原版與研究來源

- [x] 在唯讀 Docker 容器中列出 ZIP 內容，記錄全部輸入大小、SHA-256、平台與已知來源；
  發行 revision／修改歷史明示為未知。
- [x] 確認 `Knights-of-legend.zip` 中的 DOS MZ executable、資料檔與圖像候選；音訊 driver
  ／資料尚未辨識，留給工具鏈指紋階段。原始檔
  不寫回、不提交。
- [x] 建立 `CONTEXT.md`、`WORKLOG.md`、`docs/re/RESEARCH-LOG.md`、`docs/spec/` 與
  `docs/verification/`。
- [x] 確認魔旅士 2010-04-16〈Knights of Legend 武士傳說〉原文，README 已記錄直接
  連結、作者、日期與摘要。
- [x] 華人玩家來源涵蓋魔旅士繁體中文個人心得與《CRPG 通鑑》中文長篇評介；逐篇記錄
  作者、網址、日期、摘要及其是玩家觀點而非規則 oracle 的限制。
- [x] 更新 README「專案起源」，只納入可直接回查的來源，不杜撰魔旅士連結。

完成條件：原版身分固定；README 每項起源敘述可回查；找不到的來源有查詢紀錄。

## M2：說明書繁體中文化

- [x] 記錄說明書與參考卡 SHA-256、頁數、頁面尺寸及 metadata 擷取工具版本；OCR 工具
  版本待正式文字擷取時記錄。
- [x] 建立 `docs/manual/README.md`、章節檔、逐頁原文定位表與 `glossary.md`；SPEC-003
  已 READY，只授權手冊文件工作。
- [ ] 逐頁文字擷取、翻譯、人工校對；目前玩家手冊實體頁 1–70 已 reviewed，76–77 為
  目視確認空白頁；其餘玩家手冊 7 頁與參考卡 5 頁仍為 todo。
- [ ] 讓規則 spec 連回說明書頁碼；說明書未涵蓋的邊界行為不得自行推定。
- [ ] 驗證 Markdown 連結、表格、圖片權利分類與完整章節清冊。

完成條件：全書與參考卡都有可回查的繁體中文 Markdown；零個未標示的 OCR／專名疑點。

## M3：可重現素材抽取

- [ ] 先辨識既有專用 Docker image；缺件才建立有鎖版來源的 `Dockerfile`。
- [ ] 寫唯讀輸入的資產 inventory 與格式規格：COMBPICS、TWN index 24、MONPICS EGA
  垂直鏈、AWD 固定快照邊界與三個無副檔名固定記錄表已證實；其餘 archive、PCGA／HERC
  像素格式尚未完成。
- [ ] 建立離線抽取器，將所有 runtime 圖像輸出為 PNG；生成 manifest、來源索引與 SHA-256。
- [ ] 生成 atlas／接觸表，逐批人工檢查尺寸、色盤、透明色、裁切與動畫序列。
- [ ] 建立 OGG 轉換管線與 catalog；驗證 codec、時長、非靜音、峰值、loop 與人耳聆聽。
- [ ] 建立測試，保證 remake runtime 只會讀 PNG 與 OGG，禁止讀原版 archive／EXE。
- [ ] 分離本機衍生素材與可公開輸出；Git／release deny-list 預設失敗即關閉。

完成條件：從固定雜湊原版可重生 PNG／OGG；runtime 格式守門全綠；權利分類齊全。

## M4：Go／Ebitengine 垂直切片

- [ ] 鎖定 Go、Ebitengine 與 Docker build image，建立最小可啟動程式。
- [ ] 定義 package 邊界、固定 timestep、可注入 RNG／clock／input 與存檔 schema 草案。
- [ ] 選一條最小玩家鏈：真實 PNG → typed data → 世界／城鎮一個互動 → UI → 存讀檔。
- [ ] 對應原版同狀態畫面與操作，記錄差異及證據等級。
- [ ] 建立 Linux／Windows／macOS 可重現編譯與最小啟動 smoke；真機驗收另列。

完成條件：不靠 debug 入口可啟動、操作、存檔與重載；沒有 runtime 原版格式依賴。

## M5：戰鬥、任務與正常玩家路徑

- [ ] 以 IDA Pro 9.4、說明書與原版實驗建立戰鬥規劃／先見／疲勞／傷害的 DRAFT spec。
- [ ] 由使用者決定是否保留原版繁複節奏、提供可選加速，或另作現代化模式；決策前只做
  可丟棄 prototype，不改正式規則。
- [ ] 每份玩法 spec 經證據審查升 READY 後才實作，完成同狀態驗證再標 CONFORMED。
- [ ] 貫通隊伍建立、城鎮、世界旅行、隨機遭遇、任務戰場、物品、訓練與勝負流程。
- [ ] 驗證存檔可跨事件與戰鬥階段恢復，避免重複付款、任務或戰鬥結算。

完成條件：正常玩家路徑不用 direct-entry 可走到預定結局；核心規則宣稱都有收據。

## M6：中文 UI、驗收與交付

- [ ] 統一遊戲與說明書術語；每個 CJK 欄位定義文字安全矩形與 overflow policy。
- [ ] 完成最長繁中內容的視覺回歸、正常路徑截圖與輸入測試。
- [ ] 預先登記近似完成／嚴格 parity 的抽樣母體、硬失敗與平台矩陣。
- [ ] 由乾淨輸入重建 `dist-all/<版本>/{full,release,promo}`。
- [ ] 驗證每包啟動、manifest、SHA-256、授權、素材 deny-list、存檔位置與平台／架構。
- [ ] 真機驗收 Windows／macOS；Linux 容器 smoke 不冒充真機結果。

完成條件：交付包、證據、權利分類與已知限制一致；本專案 Docker 容器已清理。

## 下一個唯一工作

以 SPEC-003 READY 契約繼續玩家手冊實體頁 71–75，逐頁翻譯並以掃描影像校對；
theme 範圍仍等待使用者決定，SPEC-002 READY 前不接 Go／Ebitengine production runtime。
