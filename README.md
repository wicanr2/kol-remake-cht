# 《武士傳說》繁體中文重製

這是 Origin Systems *Knights of Legend* 的繁體中文乾淨重寫專案。正式重製版預定使用
Go 與 Ebitengine，畫面執行期只讀經可重現流程抽出的 PNG，音樂與音效使用 Ogg Vorbis。
目前處於「逆向知識庫與規格建立」階段，尚未有可玩版本；玩法必須依序通過
「RE 證據 → DRAFT spec → 證據審查 → READY spec」才會進入正式實作。

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

以下兩張外部截圖來自魔旅士的原文，未複製進本儲存庫。它們已和固定原版
`COMBPICS.0` 的解碼結果做配準交叉驗證：左圖對應 archive 圖號 0，右圖對應圖號 1。
這項證據確認場景身分、幾何與方向，不代表 JPEG 與解碼 PNG 逐像素相同；量化結果見
[COMBPICS 視覺 oracle](docs/re/combpics-visual-oracle.json)。

| 戰鬥失利後逃跑（圖號 0） | 競技場戰敗（圖號 1） |
|---|---|
| [![《武士傳說》逃跑戰鬥畫面](https://morrisjfwong.com/blog/wp-content/uploads/2010/03/kol-flee-001a-s.jpg)](https://morrisjfwong.com/blog/knights-of-legend-%E6%AD%A6%E5%A3%AB%E5%82%B3%E8%AA%AA-dos-game/) | [![《武士傳說》競技場戰敗畫面](https://morrisjfwong.com/blog/wp-content/uploads/2010/04/war-lose-001a-s.jpg)](https://morrisjfwong.com/blog/knights-of-legend-%E6%AD%A6%E5%A3%AB%E5%82%B3%E8%AA%AA-dos-game/) |

圖片著作權屬原權利人；此處只作研究來源標示與外部引用，不授予再利用或散布權。

## 現況與入口

- 私有 GitHub 儲存庫：[wicanr2/kol-remake-cht](https://github.com/wicanr2/kol-remake-cht)
- 專案規則：[AGENTS.md](AGENTS.md)
- 目前真相與交接：[CONTEXT.md](CONTEXT.md)
- 唯一工作清單：[WORKLIST.md](WORKLIST.md)
- 工作歷程：[WORKLOG.md](WORKLOG.md)
- 逆向知識庫：[docs/re/README.md](docs/re/README.md)
- 規格狀態：[docs/spec/README.md](docs/spec/README.md)
- 中文說明書：[docs/manual/README.md](docs/manual/README.md)

截至 2026-09-01：

- 79 頁玩家手冊與 5 頁參考卡已完成繁體中文 Markdown 校對，SPEC-003 為 CONFORMED。
- EGA 圖片清冊共 1,022 個 payload；996 筆已證實為可抽取來源、26 筆為非圖片，形成
  752 個已證實 PNG 單位。SPEC-002 仍為 DRAFT，正式 extractor 與顯示 theme 尚未決定。
- 六個 executable 的音訊 I/O 母體已普查；只有 TITLE 使用 PC Speaker，正常流程 WAV
  oracle 已取得。SPEC-004 仍為 DRAFT，OGG 呈現與人耳驗收尚未完成。
- 角色建立的資料表、RNG、主要屬性、起始金冠／武器、最小 writer／reader 與靜態 UI
  返回邊界已建立可重生證據。SPEC-005 仍為 DRAFT，正常 DOS UI oracle、繁中名字規則、
  typed schema 與 deterministic RNG 契約尚未關閉。
- Go／Ebitengine production runtime 尚未建立；未達 READY 的研究結果不會先接入正式路徑。

英文說明書與參考卡已逐頁翻譯並校對為繁體中文 Markdown。原版執行檔、資料、美術、地圖、
音樂、音效與轉換後素材均受獨立權利限制；程式碼授權不代表可散布這些內容。
