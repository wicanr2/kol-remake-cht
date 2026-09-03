# 《武士傳說》繁體中文重製

這是 Origin Systems *Knights of Legend* 的繁體中文乾淨重寫專案。重製版用 Go 與
Ebitengine，執行期只讀經可重現流程抽出的 PNG，音樂與音效使用 Ogg Vorbis。

目前已重繪出標題序列、書本（封面、目錄、建立角色、角色清單、刪除、Harbor Master）
與城鎮的旅店、戶外地圖、門口與建築進門，並與 DOS EGA 原版逐畫面對拍。**從標題序列到城鎮走動已經串成一條可玩路徑**：
標題 → 書本封面 → 目錄頁 → 角色選項 → 角色選擇清單 → 旅店 → DOOR → 戶外地圖 →
門口 → 進建築 → 出來，全程經過同一組狀態機，沒有除錯捷徑；26 步收據見
`docs/verification/player-path/`。建築內的後續對話、世界地圖與戰鬥都還沒實作。玩法必須依序通過
「RE 證據 → DRAFT spec → 證據審查 → READY spec」才會進入正式實作。

## 目前畫面

完整畫廊與產生流程見 [docs/screenshots/](docs/screenshots/README.md)。左欄是**英文
模式**（320×200、原版字模，逐畫面對拍用）；右欄是**繁體中文模式**（640×400、16×18
中文字模格，不對拍）。

| 標題序列 | 繁中：建立角色 |
|---|---|
| ![標題序列](docs/screenshots/en-title-knight.png) | ![繁中建立角色](docs/screenshots/zh-newchar-picture.png) |

| 建立角色（英文模式） | 繁中：旅店與組隊 |
|---|---|
| ![建立角色](docs/screenshots/en-newchar-picture.png) | ![繁中旅店](docs/screenshots/zh-town-inn-party2.png) |

這些畫面全部由 `tools/run_kol_shots.sh` 在鎖定容器內重生，`docs/screenshots/MANIFEST.json`
記下每張的產生指令、輸入雜湊與 PNG SHA-256，並要求有對拍收據的畫面與那張逐 byte 相同。

## 專案起源

這個計畫起於華人老遊戲玩家對《武士傳說》的長期記憶：它把細緻的角色建立、紙娃娃
裝備、疲勞與預先規劃攻防帶進電腦角色扮演遊戲，卻也因每回合選項繁多、反覆確認和大型
戰場而顯得非常緩慢。這種「值得保存的深度」與「阻礙今日玩家接近的摩擦」正是重製的
出發點；它們目前只是玩家體驗假說，之後仍須以說明書與原版實玩逐項驗證。

魔旅士在 2010 年 4 月 16 日發表的
[〈Knights of Legend 武士傳說〉](https://morrisjfwong.com/blog/knights-of-legend-%E6%AD%A6%E5%A3%AB%E5%82%B3%E8%AA%AA-dos-game/)
是本專案的重要起源：他記得本作變化多、耐玩、畫面在當年相當出色，也具體描述戰鬥
難度高、普通遭遇可能拖上很久，以及隊伍編成、地形、斥候、逃跑與武器保管等玩家生存
經驗。他還提到自己期待中文使用手冊；這直接支持本專案把完整繁體中文說明書列為主要
交付物，而不只是附帶翻譯。

另一項可回查來源是
[《CRPG 通鑑》中文版〈傳說騎士〉](https://crpgbook.cn/2022/01/01/065-knights-of-legend/)，
該文整理了遊戲的開發背景、細緻世界觀與戰鬥系統，也同樣指出繁複操作和漫長戰鬥的
代價。這些文章說明「值得保存的系統深度」和「值得重新檢視的操作節奏」為何同時存在；
它們屬於玩家觀點與專案動機，不會被直接當成原版規則證據。

## 已確認的原版畫面

以下兩張外部截圖來自魔旅士的原文。它們已和固定原版
`COMBPICS.0` 的解碼結果做配準交叉驗證：左圖對應 archive 圖號 0，右圖對應圖號 1。
這項證據確認場景身分、幾何與方向，不代表 JPEG 與解碼 PNG 逐像素相同；量化結果見
[COMBPICS 視覺 oracle](docs/re/combpics-visual-oracle.json)。

| 戰鬥失利後逃跑（圖號 0） | 競技場戰敗（圖號 1） |
|---|---|
| [![《武士傳說》逃跑戰鬥畫面](https://morrisjfwong.com/blog/wp-content/uploads/2010/03/kol-flee-001a-s.jpg)](https://morrisjfwong.com/blog/knights-of-legend-%E6%AD%A6%E5%A3%AB%E5%82%B3%E8%AA%AA-dos-game/) | [![《武士傳說》競技場戰敗畫面](https://morrisjfwong.com/blog/wp-content/uploads/2010/04/war-lose-001a-s.jpg)](https://morrisjfwong.com/blog/knights-of-legend-%E6%AD%A6%E5%A3%AB%E5%82%B3%E8%AA%AA-dos-game/) |

圖片著作權屬原權利人；此處只作研究來源標示與外部引用，不授予再利用或散布權。本專案
另允許已驗證、可重生且有 manifest 的原版衍生 PNG 保存於目前的 private GitHub；這項
專案例外不包含原版 raw 檔，也不授權 GitHub Release 或公開散布。

## 現況與入口

- 私有 GitHub 儲存庫：[wicanr2/kol-remake-cht](https://github.com/wicanr2/kol-remake-cht)
- 專案規則：[AGENTS.md](AGENTS.md)
- 目前真相與交接：[CONTEXT.md](CONTEXT.md)
- 唯一工作清單：[WORKLIST.md](WORKLIST.md)
- 工作歷程：[WORKLOG.md](WORKLOG.md)
- 逆向知識庫：[docs/re/README.md](docs/re/README.md)
- 規格狀態：[docs/spec/README.md](docs/spec/README.md)
- 中文說明書：[docs/manual/README.md](docs/manual/README.md)
- 目前畫面：[docs/screenshots/README.md](docs/screenshots/README.md)

截至 2026-09-03：

- 79 頁玩家手冊與 5 頁參考卡已完成繁體中文 Markdown 校對，SPEC-003 為 CONFORMED。
- EGA 圖片清冊共 1,022 個 payload；996 筆已證實為可抽取來源、26 筆為非圖片，形成
  752 個已證實 PNG 單位。SPEC-002 已 READY，首版採 EGA-only，CGA／Hercules 列為後續。
- 角色建立與刪除的資料表、RNG、屬性、起始金冠／武器與寫入欄位都已由正常 DOS 玩家
  路徑收據關閉，SPEC-005 已 READY。刪除只寫一個 byte，remake 的輸出與原版逐 byte 相同。
- remake 對拍收據二十份共 259 對畫面，允許差異區之外全部零像素差；允許區只列原版自畫的
  滑鼠游標與已查證成因的殘影。收據見 [remake 對拍](docs/verification/remake-parity/README.md)。
- 繁體中文介面採雙模式：英文模式維持原版解析度供嚴格對拍，中文模式放大兩倍以容納
  16×18 中文字模格。書本五個畫面、旅店、戶外與野外地圖已中文化，
  七座城 101 棟建築的 1,938 則對話全部翻完，文字安全矩形檢查 6,725 項全過。
- 六個 executable 的音訊 I/O 母體已普查；只有 TITLE 使用 PC Speaker，正常流程 WAV
  oracle 已取得。SPEC-004 仍為 DRAFT，OGG 的人耳驗收尚未完成，因此還沒接上音訊路徑。
- 城鎮可實際遊玩：走動、門口、進建築、招呼語與 LISTEN／MOUTH 對話、招募同伴、
  客棧住宿與修道院治療、MIRROR 的角色資料頁；走出城門接到野外地圖，
  含 ROAD／SIGNPOST 自動循路與進城。買賣、訓練、戰鬥與任務都還沒實作。
- 存檔是 remake 自有的可讀 JSON（SPEC-006），住宿時寫出；原版 `CHARDATA` 一律唯讀，
  要搬原版角色進來走離線匯入工具。載入存檔繼續玩這條路徑還沒接上。
- 鏡頭跟隨的 dead zone 是由 22 張原版畫格反解的**強推論**，不是從程式碼讀出來的。

英文說明書與參考卡已逐頁翻譯並校對為繁體中文 Markdown。原版執行檔、資料、美術、地圖、
音樂、音效與轉換後素材均受獨立權利限制；程式碼授權不代表可散布這些內容。
