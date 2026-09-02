# 《武士傳說》重製工作清單

更新日期：2026-09-02

## 目前狀態

| 分類 | 狀態 | 目前證據／限制 |
|---|---|---|
| remake | 書本與旅店可對拍 | SPEC-002／005 READY，752 張 PNG；十份收據共 180 對逐畫面通過，允許區外零像素差 |
| 原版 oracle | build 已固定 | ZIP 與全部項目逐檔雜湊已固定；發行來源／1998 timestamp 成因未知 |
| 玩家觀點 | 起源資料完成 | 魔旅士繁中個人心得＋《CRPG 通鑑》中文長篇評介已進 README |
| 繁中介面 | 書本五頁＋旅店 | 雙模式；中文字模表 344 字；文字安全矩形 122 項全過 |
| 可選美化 | 未決策 | 不得在忠實度與操作方向確認前進入 production |
| 圖像 theme | 已決策 | 2026-09-01：首版 EGA-only；CGA／Hercules 列後續選項 |
| TITLE 音訊 | 候選資產待人耳驗收 | 規格近似合成已產出；時長與原版擷取差 0.6 ms；SPEC-004 仍 DRAFT |
| GitHub | 已建立 | private：`wicanr2/kol-remake-cht`；`main` 已推送 |

## M0：專案建立

- [x] 撰寫 `AGENTS.md`，固定 Docker-only、證據、Go／Ebitengine、PNG-only runtime、
  OGG、中文說明書、授權與 `dist-all/` 契約。
- [x] 建立本工作清單與 README 起源骨架。
- [x] 初始化 Git，確認原版封裝與未驗證衍生素材由 `.gitignore` 排除；2026-09-01 起只允許
  已驗證且有 manifest 的 PNG 衍生物進 private repo 指定路徑。
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
  已 CONFORMED，且只證實手冊文件工作。
- [x] 逐頁文字擷取、翻譯、人工校對；玩家手冊 79 頁與 Apple II 參考卡 5 頁全部 reviewed
  （玩家手冊 76–77 為目視確認空白跨頁）。
- [ ] 讓規則 spec 連回說明書頁碼；說明書未涵蓋的邊界行為不得自行推定。
- [x] 驗證 Markdown 連結、表格、圖片權利分類與完整章節清冊；84 頁／12 輸出全套 gate 通過。

完成條件：全書與參考卡都有可回查的繁體中文 Markdown；零個未標示的 OCR／專名疑點。

## M3：可重現素材抽取

- [ ] 先辨識既有專用 Docker image；缺件才建立有鎖版來源的 `Dockerfile`。
- [ ] 寫唯讀輸入的資產 inventory 與格式規格：COMBPICS 全 archive、TWN0..7PICS 全 243 筆、MONPICS、PERPICS、
  MISCPICS、TITLE SET0／1／2／FACE／BACK、ENDPICS、BOOKPICS、KOLMPICS 全 83 筆 EGA
  垂直鏈、WEATPICS 全 18 筆六組分層合成、AWD1PICS 全 24 張勳章圖示、SHLDPICS
  全四組盾牌、HANDPICS 全 12 組手部 sprite、WEAPPICS 全 28 組角色武器、
  HEADPICS／TORPICS／LEGSPICS 各 54 組角色分層 sprite、CHARPICS.0 全 80 張角色模板、AWD
  固定快照邊界、KOLPICS 全 14 筆混合 archive 與三個無副檔名固定記錄表的 40×38
  4-bit EGA 像素契約已證實；TWN 七筆圖號 1 placeholder 已明列排除；其餘 archive、PCGA／HERC
  像素格式尚未完成。
- [x] 勘誤 MONPICS／PERPICS 早期 PNG 探針的 AND-mask alpha 極性；依
  `sub_1BD17` 重生為 bit 1 透明、bit 0 不透明，並用 8×8 棋盤背景重驗。
- [x] 建立離線抽取器 `tools/extract_ega_assets.py`：30 個 archive、752 張 PNG 輸出到
  `assets/original-derived-png/ega/`，原子化替換、逐張重讀驗證、對照 visual probe 雜湊，
  並產生含來源與權利分類的 manifest。
- [x] 建立完整 EGA extraction set：30 個邏輯 archive／1,022 payload 全數有 disposition；
  996 筆已證實、0 筆待證實、26 筆非圖片，並有清冊一致性 verifier。
- [ ] 生成 atlas／接觸表，逐批人工檢查尺寸、色盤、透明色、裁切與動畫序列。
- [x] 以雙新建 IDA DB 關閉六個 executable 的直接 x86 音訊 I/O 母體；只有 TITLE
  PC Speaker 音爆，並排除 `TANDYPIN.EXE` 是音訊播放器。
- [x] 用 DOSBox-X `DX-CAPTURE /A` 在正常 TITLE 流程取得非靜音 WAV，固定四段活動窗、
  codec、時長、峰值與 SHA-256；不追逐週期 parity。
- [x] TITLE cue 的 OGG 呈現方向已決策：規格近似合成（依已證實 PIT 除數序列與公開硬體
  規格自行合成，不散布原版錄音）。
- [ ] 依決策實作 PIT 方波合成器並完成人耳驗收；合成輸出與 DX-CAPTURE WAV 的段落結構
  比對後才可升 SPEC-004。
- [x] 建立 OGG catalog 的 SPEC-004 DRAFT 與機器可讀音訊 inventory；尚不授權轉換實作。
- [ ] SPEC-004 證據審查並升 READY 後，建立 OGG 轉換管線；驗證 codec、時長、非靜音、
  峰值、loop 與人耳聆聽。
- [ ] 建立測試，保證 remake runtime 只會讀 PNG 與 OGG，禁止讀原版 archive／EXE。
- [ ] 分離本機、private-repo PNG 與可公開輸出；GitHub Release／公開 release deny-list
  預設失敗即關閉，並檢查 repo visibility 未被改成 public。

完成條件：從固定雜湊原版可重生 PNG／OGG；runtime 格式守門全綠；權利分類齊全。

## M4：Go／Ebitengine 垂直切片

- [x] 建立角色建立 RE 證據表與 SPEC-005 DRAFT；固定手冊流程、16 個角色槽、768-byte
  CHARDATA stride 與 KOLBOOK 開檔路由，但不授權 production 實作。
- [x] 關閉主要屬性 RNG 的 seed、有界擲骰、獨立骰式／內部基值、骰面表、100/150 顯示
  量尺與三項衍生公式；兩個全新 IDA DB 輸出一致。
- [x] 關閉 35 個手冊職業列與 39 個 executable 記錄的 enum／名稱／eligibility 對應；
  解釋男女精靈共名及四對矮人記錄逐 byte 相同。
- [x] 關閉 39 筆起始金冠、主要／次要武器、武器技能與最小 writer 欄位索引。
- [x] 查核必要 KOLTOWN reader，並以完整 writer coverage 排除其他職業衍生起始資料。
- [x] 以雙新建 IDA DB 固定「輸入 buffer 首 byte 為零」分支、各層取消、reroll、80 張
  肖像循環與滿槽不進 writer；零 byte 不再擅自等同玩家按 Enter 的空名。
- [x] 建立畫面驅動的正常 DOS UI 擷取器（`tools/dosbox_character_ui_oracle.py`、
  `tools/x11_key_inject.py`、`tools/compare_character_ui_runs.py`），並封存名字長度、
  空名、各層 ESC 與 reroll 四個情境各兩輪的畫格、輸入序列與三檔前後 SHA-256 收據。
- [x] 完成正常 DOS UI 的接受屬性、圖像選擇與實際保存 oracle；SPEC-005 經證據審查升
  READY。16 槽全滿仍只有靜態證據。
- [ ] 補 16 槽全滿、重名與名字欄字元集合的正常 UI oracle。
- [x] 鎖定 Go 1.26.7、Ebitengine v2.9.9 與建置映像 `eob-remake-go:1.26.7-ebiten2.9.9-audio`；
  `cmd/kol` 可建置、可啟動，並有 `tools/go_docker.sh`／`tools/run_kol_headless.sh` 兩支包裝。
- [ ] 定義 package 邊界、固定 timestep、可注入 RNG／clock／input 與存檔 schema 草案。
- [ ] 選一條最小玩家鏈：真實 PNG → typed data → 世界／城鎮一個互動 → UI → 存讀檔。
- [x] 建立畫面對拍工具與首批收據：`tools/verify_remake_parity.py` 的三層判準，
  82 對書本畫面全部通過（`docs/verification/remake-parity/`）。
- [x] 抽出原版內嵌點陣字型並機械量得書本文字版面；remake 以原版繪字規則重繪空白
  New Character 頁，與原版只差名字欄的 31 個原版殘留像素。
- [x] 量測角色選項、屬性數值與右頁選單的文字版面，並實作名字→種族→性別→職業→
  擲屬性的狀態機；9 對逐畫面對拍通過，殘餘差異只剩四類已查證成因（RE-045）。
- [x] 實作圖像層版面（右頁說明、`CANCEL   ACCEPT`、青色雙格外框、名牌、上下箭頭）；
  完整流程加 80 張圖像循環共 92 對逐畫面對拍通過（RE-046）。同時指認大字型的第一個
  畫面用途：codepoint 94／95 就是上下箭頭。
- [x] 角色選項頁、角色清單頁與兩段翻頁動畫（RE-049）；新增 `list-characters` 正常 DOS UI
  情境兩輪，`CHARDATA` 的槽位版面由畫面讀回的名字唯一命中確認。對拍 16 ＋ 20 對全通過。
- [x] 角色寫入的逐 byte 對拍（RE-050）：原版改動的 58 個 byte，remake 的 writer 兩個職業
  各 58/58 逐一相同，寫錯 0、多寫 0。建立→寫入→清單已閉環（清單由 14 列變 15 列）。
- [x] 起始武器樣板的來源與索引規則已反追（RE-050）：樣板是 `BOOKDATA` 第 3 項的 14 筆
  16-byte 記錄，索引 ＝ selector − 1；selector 藏在職業 metadata 第一個 byte 的兩個
  nibble。39 個職業都能寫，21 筆沒有副武器的不會誤寫副武器欄位。
- [x] 刪除角色（RE-051）：新增 `delete-character`、`delete-character-confirm` 兩個正常
  DOS UI 情境各兩輪，13/13 與 15/15 個 checkpoint 逐檔相同。刪除畫面就是角色清單頁
  加反白條，確認視窗只認 `y`／`n`；刪除只把槽位名字的第一個 byte 寫成 0，remake 的
  輸出與原版逐 byte 相同。對拍 8 ＋ 5 對全通過。
- [x] `HARBOR MASTER`（RE-052）：只安裝一個 region 時是一張三行訊息頁，1 對對拍通過。
- [x] 書本四個功能的入口已查清（RE-052）：`DELETE CHARACTER`、`EDIT PICTURE`、
  `EDIT SHIELD`、`HARBOR MASTER` 共用同一張角色選擇清單，進入後第一張畫面是同一個 PNG。
- [ ] `EDIT PICTURE`／`EDIT SHIELD` 的點陣圖編輯器：原版是滑鼠驅動、ESC 無作用，
  現有鍵盤 oracle 走不完；要往下做得先讓 oracle 能送滑鼠事件。
- [x] 走出書本進入城鎮（RE-053）：目錄頁 `6` → 共用角色選擇清單 → 選領隊 → 半途之家
  旅店。圖示列不吃方向鍵，改用參考卡的 `<`／`>` 移動白框、Return 執行，末格繞回；
  `tools/x11_key_inject.py` 補上 Shift 位 keysym 的送鍵。
- [x] 組隊流程已量測（RE-053）：COMPANION → 是否搜尋硬碟 → 該旅店的可招募清單 →
  上下選、Return 加入，頭像出現在頂端。招募不寫磁碟，三個角色檔全程未變。
- [x] 旅店畫面的組成已拆解（RE-054）：底圖是 `KOLMPICS[0]`（與 `BOOKPICS[3]` 同一張），
  上面疊 11 個矩形，分群之外殘餘 0 像素；六個圖示逐像素對上 KOLMPICS 圖號。
- [x] 店主肖像已定位（RE-054）：`TWN7PICS[12]` 88×64 畫在 (8,64)，88×64 像素全同。
  找法是拿畫面上一列 60 像素當指紋去搜素材，不是拿素材去搜畫面。
- [x] 旅店畫面的 remake 實作與對拍（RE-055）：`internal/town` ＋ `ui.Inn`，
  八對逐畫面對拍全過（六個圖示白框位置、單人隊伍、招募後兩人隊伍），
  允許區只有原版滑鼠游標。中文模式一併補上。
- [x] 戶外地圖已拆解（RE-056）：底圖、徽章（`TWN7PICS[0]`）、日期框、頭像格與單一
  `KOLMPICS[5]` 圖示都零差異對上；地圖視窗是 12×6 格、每格 16×16 的**捲動磚塊圖**，
  移動時以 16 px 為單位捲動。
- [x] 戶外地圖的磚圖已定位（RE-057）：在 `TOWNn.0` 城鎮資料檔裡，payload 0／1／2 各 96
  塊，格式是 16×16、四個 EGA 平面、plane-major，每塊 128 bytes。畫面上 72 塊磚有 67 塊
  逐 byte 命中，offset 全是 128 的倍數。`TOWNn.0` 之前不在 inventory 裡，是因為清冊工具
  只認名字含 `PIC` 的檔案。
- [x] 磚圖的載入路徑已讀出（RE-058）：`sub_15920` 把 `town%d` 的 payload 0／1／2 放到同一
  segment 的 0x0000／0x3000／0x6000，再交給 `sub_19B29` 轉色盤索引——證實磚塊是 288 筆的
  平坦陣列，第 n 塊位移 `n×128`。
- [ ] 地圖格子陣列：六種資料側搜尋全部零命中（清單見 RE-058），因此不再對資料硬搜；
  下一步是從 `sub_16183`／`sub_1766F` 這一帶的城鎮畫面流程往下追繪製迴圈。
- [ ] 把 `TOWNn.0` 納入正式抽取器與清冊（目前只有探針，還沒有 production extractor）。
- [ ] 圖示列其餘五個圖示、隊伍上限六人、發言者切換與 `REST/SAVE/HEAL` 的寫入。
- [ ] remake 自己的存檔格式：目前的 writer 是為對拍而寫的原版格式輸出；正式 schema 屬
  共同決策，尚未決定。
- [ ] 量測其餘畫面（城鎮、世界、戰鬥）的文字版面。
- [x] 繁中雙模式：英文模式 320×200 走原版字模與嚴格對拍；中文模式 640×400、16×18
  中文字模格。字模由 `tools/build_cjk_font.py` 從 Noto Sans CJK TC 烘成點陣表，
  執行期只讀 PNG。New Character 頁全頁中文化，文字安全矩形 98 項全過。
- [x] 玩家可見的展示截圖：`docs/screenshots/` 19 張由 `shots.spec` 重生，有 manifest
  與就地驗證（尺寸、顏色數、與對拍收據逐 byte 相同）；README 改引用這一批。
- [ ] 其他畫面的中文化（城鎮、世界、戰鬥）與職業名暫譯的交叉確認。
- [x] 標題序列：從 `TITLE.EXE` 的 `_main` 靜態讀出腳本（RE-047），remake 依腳本重播一輪，
  原版兩輪觀察的每一張畫格都逐像素對上，沒有未解釋的畫面。
- [x] 標題序列的實際播放：時間迴圈與按鍵中止已實作（`-scene title`）。
- [x] 翻書動畫：從 `KOLBOOK.EXE` 的 `sub_114F5` 讀出表驅動的逐欄捲動（RE-048），
  remake 依同一張表重播，18 對逐畫面對拍全通過。
- [ ] 標題序列的雷聲播放：OGG 已產出但尚未接上 Ebitengine 音訊路徑（SPEC-004 仍 DRAFT）。
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

地圖格子陣列。磚圖與它的載入路徑都解出來了（RE-057、RE-058）：288 塊 16×16 四平面
磚塊，第 n 塊位移 `n×128`。剩下的是「哪一格畫哪一塊磚」。

RE-058 列出六種已經排除的資料側問法（一格一 byte、一格兩 byte、相等關係、偏移不變
差分、nibble-RLE、掃 ×128 的算式），全部零命中——**不要再對資料硬搜**。
下一步改從 `sub_16183`／`sub_1766F` 這一帶的城鎮畫面流程往下追繪製迴圈。

格子陣列解出來之後：把 `TOWNn.0` 納入正式抽取器 → 實作戶外畫面 → 逐畫面對拍。
在那之前不動 production path。

等使用者決策才能動的項目（不擋上面的工作）：

1. remake 正式存檔 schema（SPEC-005 共同決策）。
2. 戰鬥節奏與忠實度方向（原版繁複節奏是否保留、是否提供可選加速）。
3. `assets/audio/title-thunder.ogg` 的人耳驗收，通過才把 SPEC-004 升 READY。
4. 原版衍生 typed data 的權利分類（六份本機 JSON 是否比照 PNG 進 private repo）。
