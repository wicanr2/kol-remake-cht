# 《武士傳說》繁體中文重製

這是 Origin Systems *Knights of Legend* 的繁體中文乾淨重寫專案。重製版用 Go 與
Ebitengine，執行期只讀經可重現流程抽出的 PNG；正式音訊使用 Ogg Vorbis。標題雷聲已通過人耳驗收與[正常播放驗證](docs/verification/audio/README.md)。

目前提供書本與建角、組隊、城鎮與世界探索、戰鬥、任務、結局及 kol-save-v3 保存。
現行交付為 `v.1.0.1-20261012`，含三平台完整版、需匯入合法資料的 release 與本機推廣影片。
目前採[受控場景驗收](docs/spec/040-controlled-scene-acceptance.md)：允許鎖血與提高攻擊，
抵達指定場景後確認中文與畫面。登記場景已通過，結果見[驗收收據](docs/verification/controlled-scenes-20261011.md)。
本版已包含住宅告示、中文目錄、結局排版修正、首批 HD 與遊戲內畫面切換。

## 遊戲背景

故事發生在阿什塔拉瑞亞半島。人類、精靈與矮人在此建立聚落，卻曾遭巫師 Pildar
的大軍侵襲。騎士 Seggallion 與盟軍終結黑暗戰爭後，各地重建了十五年；如今英雄
失蹤，森林裡的怪物再度增多。玩家帶領冒險隊走訪城鎮、打聽消息與接取任務，
逐步揭開危機。背景依原版手冊的[阿什塔拉瑞亞王國](docs/manual/chapters/08-realm-of-ashtalarea.md)整理。

本作著重隊伍編成與回合規劃。玩家替每位隊員安排移動、攻擊與防禦，再觀看雙方行動；
裝備、傷勢和疲勞都會影響決策。回城後可治療、訓練、買賣裝備，並在旅店住宿保存。
玩法與操作見[繁體中文手冊](docs/manual/README.md)。

這套構想始於 1981 年四位玩家對紙筆與電腦角色扮演遊戲的討論。Todd Porter 後來製作
原型，與 Origin Systems 合作發行。開發經過見手冊的[《武士傳說》的誕生](docs/manual/chapters/00-front-matter.md)。

## 開始遊玩

現行版本為 `v.1.0.1-20261012`，包含已驗收雷聲、中文修正、首批 HD 與遊戲內切換。
封包與限制見 [交付收據](docs/verification/release-v1.0.1-20261012.md)。
正式交付根目錄為 [dist-all/v.1.0.1-20261012/](dist-all/v.1.0.1-20261012/)：

| 平台 | 檔案 | 開啟方式 |
|---|---|---|
| Linux | [完整版 AppImage](dist-all/v.1.0.1-20261012/full/kol-v.1.0.1-20261012-full-linux-x86_64.AppImage) | `chmod +x` 後執行；沒有 FUSE 可加 `--appimage-extract-and-run` |
| Windows AMD64 | [完整版 ZIP](dist-all/v.1.0.1-20261012/full/kol-v.1.0.1-20261012-full-windows-amd64.zip) | 解壓後開啟 `kol.exe`；已通過 Wine 啟動與切換 |
| macOS ARM64／Intel | [完整版 universal ZIP](dist-all/v.1.0.1-20261012/full/kol-v.1.0.1-20261012-full-macos-universal.zip) | 開啟 `武士傳說.app`；只完成建置與結構驗收 |

`full` 已含遊戲與首批 HD 必要資料，只留本機。[release](dist-all/v.1.0.1-20261012/release/) 是三平台引擎包，
需先以包內工具匯入合法持有、已抽取且符合雜湊的資料；說明見各包 `README.txt`。
三平台引擎包及校驗碼已發布至 [GitHub Release](https://github.com/wicanr2/kol-remake-cht/releases/tag/v.1.0.1-20261012)，儲存庫維持 private。
[76 秒推廣影片](dist-all/v.1.0.1-20261012/promo/kol-v.1.0.1-20261012-promo.mp4)
使用本版抽樣與遊戲內切換的新畫格，也只留本機。檔案與 SHA-256 見交付根目錄的 `SHA256SUMS.json`。
嚴格原版一致性與平台真機驗收另列後續研究，
目前狀態以 [CONTEXT.md](CONTEXT.md) 與 [工作清單](WORKLIST.md) 為準。

操作照原版參考卡：

| 鍵 | 作用 |
|---|---|
| 方向鍵 | 移動 |
| `A` `S` `Z` `X` | 斜走 |
| `,` `.` | 移動畫面下方圖示列的白框 |
| Enter | 執行 |
| Esc | 回上一頁 |
| F10 | 開啟畫面設定，切換 EGA／HD |
| Ctrl+Q | 離開 |

滑鼠可以直接點選單、清單與圖示列，右鍵等於 Esc，滾輪上下移動游標。點某一列
送出的是對應鍵盤操作；本專案接受滑鼠替代，只要求玩家操作結果等價。
不同畫面的滑鼠命中與正常玩家流程仍須逐項驗收，不以共用按鍵處理宣稱全部畫面已對拍。

介面預設繁體中文；`-lang en` 切成與原版逐畫面對拍的英文版面。住宿存檔寫在使用者
設定目錄（Linux `~/.config/kol-remake-cht/`、macOS
`~/Library/Application Support/kol-remake-cht/`、Windows `%AppData%\kol-remake-cht\`），
用 `-load <路徑>` 讀回來。

macOS 那一份沒有 Apple 開發者簽章，第一次開啟要在「系統設定 → 隱私權與安全性」
按「仍要打開」，或先執行 `xattr -dr com.apple.quarantine`。Windows 與 macOS
都還沒有真機驗收。

## 目前畫面

以下畫面由鎖定的 Docker 工具鏈重新擷取。繁中畫面為 960×600，使用 16×18 字模格。
圖片來源、輸入、畫格編號、SHA-256 與權利分類見[截圖清冊](docs/verification/cjk-closeout-20261011/manifest.json)。
城鎮及戰鬥來自正常標題、六人組隊與自然遭遇；建角畫面為展示入口。

| 中文目錄 | 建立角色 |
|---|---|
| ![中文目錄](docs/verification/cjk-closeout-20261011/zh-contents.png) | ![建立角色](docs/verification/cjk-closeout-20261011/zh-newchar.png) |

| 六人組隊 | 戰鬥規劃 |
|---|---|
| ![六人組隊](docs/verification/cjk-closeout-20261011/zh-party.png) | ![戰鬥規劃](docs/verification/cjk-closeout-20261011/zh-battle.png) |

既有英中並排畫廊與各自驗收範圍見[畫面索引](docs/screenshots/GALLERY.md)。
原版沒有繁中畫面，中文排版驗收與英文原版像素比較分開記錄。

## HD 化計畫

首批手繪旅店門面與六人肖像已接入實際遊戲，採 1920×1200 原比例。
點下緣「畫面 F10」或按 F10，選擇 EGA／HD 即可在同一場景比較；Esc 返回遊戲。
預設使用 EGA，切換在本次遊玩有效。三平台完整版已含首批素材。
角色按實際圖像匹配，自訂肖像沿用原圖；戰鬥資訊框顯示目前選中的人物。
中文、滑鼠點選、舊 v3 與住宿存讀檔已通過[首批驗收](docs/verification/hd-runtime-20261011.md)。
遊戲內切換的核准版面、中文、鍵鼠、縮放及正常住宿存讀檔已通過[正式驗收](docs/verification/display-switch-runtime-20261012.md)。

| 實際旅店門面與六人肖像 | 實際戰鬥資訊框 |
|---|---|
| ![HD 旅店與六人](docs/verification/hd-runtime-20261011/inn-door.png) | ![HD 戰鬥人物](docs/verification/hd-runtime-20261011/battle.png) |

首批素材位於 [assets/hd/first-slice/manifest.json](assets/hd/first-slice/manifest.json)，
核准樣張保存在[比較原型](docs/verification/hd-first-slice-20261011/review.html)。
接續分批處理書本、人物裝備、城鎮、戰鬥與天候；外框、圖示、戰場及傷勢遮罩仍使用 EGA。
批次與素材閘門見[HD 工作計畫](docs/spec/039-hd-art-ui.md)。

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

79 頁玩家手冊與 5 頁參考卡已完成繁體中文 Markdown 校對。現行程式涵蓋建角、
城鎮、世界、戰鬥、任務與存讀檔；可玩性、最新正常玩家路徑和原版一致性分開驗收。
目前的實際缺口、證據等級與下一閘門以 [CONTEXT.md](CONTEXT.md) 和
[WORKLIST.md](WORKLIST.md) 為準；舊自評留在工作清單的歷史段落，逐輪進度見
[WORKLOG.md](WORKLOG.md)。

## 授權、致謝與聲明

本專案採**復古重製 source-available 授權條款 1.0（RRSAL-1.0）**，全文見
[`LICENSE`](LICENSE)。**這不是開源授權**——非商業限制不符合 OSI 的開源定義，
正確的說法是 source-available（原始碼公開的專有授權）。

- **非商業用途免費**，不必事先取得同意：使用、重製、散布、修改並散布修改版、
  做其他語言的版本。條件是保留條款全文、標明改了什麼與原始出處、不收費、
  不散布原版素材。
- **實況、影片、評論、報導、論文與展覽明示允許**，而且**平台分潤與觀眾贊助不算
  商業使用**。請署名作品名、作者與儲存庫網址。
- **商業用途要先洽談**：販售、付費或訂閱、隨商品或服務散布、廣告與內購、營利事業
  營運中使用、商業展覽。歡迎來信 <wicanr2@gmail.com>。
- **授權範圍不包含原版的任何東西**：Origin Systems 及其權利繼受人持有原版的執行檔、
  資料檔、美術、音樂、音效、地圖、遊戲文字與商標。本儲存庫裡由原版重生的 PNG、
  原版執行畫面截圖、譯文中保留的原版英文字串，以及由原版量測得出的位址與數值表，
  其中屬於原版的部分仍屬原權利人；條款第 2 條逐項寫明了界線。
- 第三方元件依各自授權：中文點陣字模烘製自 **Noto Sans CJK TC**（SIL Open Font
  License 1.1），執行期使用 **Ebitengine** 及其相依套件。

原版執行檔、資料、美術、地圖、音樂、音效與轉換後素材均受獨立權利限制；
程式碼授權不代表可散布這些內容。
