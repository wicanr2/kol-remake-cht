# 《武士傳說》繁體中文重製

這是 Origin Systems *Knights of Legend* 的繁體中文乾淨重寫專案。重製版用 Go 與
Ebitengine，執行期只讀經可重現流程抽出的 PNG；正式音訊使用 Ogg Vorbis。標題雷聲已通過人耳驗收與[正常播放驗證](docs/verification/audio/README.md)。

目前提供書本與建角、組隊、城鎮與世界探索、戰鬥、任務、結局及 kol-save-v3 保存。
現行交付為 `v.1.0.2-20261013`，含三平台完整版、需匯入合法資料的 release 與本機推廣影片。
目前採[受控場景驗收](docs/spec/040-controlled-scene-acceptance.md)：允許鎖血與提高攻擊，
抵達指定場景後確認中文與畫面。登記場景已通過，結果見[驗收收據](docs/verification/controlled-scenes-20261011.md)。
本版包含住宅告示、中文目錄、結局排版修正、HD介面與頭像、四首原創室內樂，以及遊戲內F9配樂與F10畫面切換。
建立隊伍與[商店購買](docs/verification/player-path/town-shop-receipt.json)已有操作收據；七段結局已通過受控展示，
從新開局正常完成全部任務至結局仍未驗證。HD 目前只完成部分介面與美術，完整範圍見下方計畫。

## 遊戲背景

故事發生在阿什塔拉瑞亞半島。人類、精靈與矮人原本各有聚落，巫師 Pildar 卻帶領
獸人、巨魔與食人魔大軍發動黑暗戰爭。矮人的要塞陷落，Shellernoon 遭到長年圍攻，
各族為守住家園付出慘重代價。

騎士 Seggallion 聯合人類、精靈與歸來的矮人軍隊，終於解除了圍城。Pildar 帶著殘部
退往湖中島嶼的黑塔。戰後十五年，各地修築道路、恢復貿易，逐漸重建生活。

如今 Seggallion 失蹤，北方襲擊不斷，森林裡的怪物也再度增多。玩家建立最多六人的
冒險隊，從旅店出發，在城鎮打聽消息、接取委託，再深入荒野與遺跡，為這片尚未恢復
安寧的土地尋找出路。背景依原版手冊的[阿什塔拉瑞亞王國](docs/manual/chapters/08-realm-of-ashtalarea.md#黑暗戰爭)
整理；隊伍與冒險流程見[開始冒險](docs/manual/chapters/01-getting-started.md)。

本作著重隊伍編成與回合規劃。玩家替每位隊員安排移動、攻擊與防禦，再觀看雙方行動；
裝備、傷勢和疲勞都會影響決策。回城後可治療、訓練、買賣裝備，並在旅店住宿保存。
玩法與操作見[繁體中文手冊](docs/manual/README.md)。

這套構想始於 1981 年四位玩家對紙筆與電腦角色扮演遊戲的討論。Todd Porter 後來製作
原型，與 Origin Systems 合作發行。開發經過見手冊的[《武士傳說》的誕生](docs/manual/chapters/00-front-matter.md)。

## 開始遊玩

現行版本為 `v.1.0.2-20261013`，包含HD操作介面與頭像、四首原創室內樂、F9配樂及F10畫面切換。
封包與限制見 [交付收據](docs/verification/release-v1.0.2-20261013.md)。
正式交付根目錄為 [dist-all/v.1.0.2-20261013/](dist-all/v.1.0.2-20261013/)：

| 平台 | 檔案 | 開啟方式 |
|---|---|---|
| Linux | [完整版 AppImage](dist-all/v.1.0.2-20261013/full/kol-v.1.0.2-20261013-full-linux-x86_64.AppImage) | `chmod +x` 後執行；沒有 FUSE 可加 `--appimage-extract-and-run` |
| Windows AMD64 | [完整版 ZIP](dist-all/v.1.0.2-20261013/full/kol-v.1.0.2-20261013-full-windows-amd64.zip) | 解壓後開啟 `kol.exe`；已通過 Wine 啟動與切換 |
| macOS ARM64／Intel | [完整版 universal ZIP](dist-all/v.1.0.2-20261013/full/kol-v.1.0.2-20261013-full-macos-universal.zip) | 開啟 `武士傳說.app`；只完成建置與結構驗收 |

`full` 已含遊戲、HD介面與配樂必要資料，只留本機。[release](dist-all/v.1.0.2-20261013/release/) 是三平台引擎包，
需先以包內工具匯入合法持有、已抽取且符合雜湊的資料；說明見各包 `README.txt`。
三平台引擎包及校驗碼已發布至 [GitHub Release](https://github.com/wicanr2/kol-remake-cht/releases/tag/v.1.0.2-20261013)，儲存庫維持 private。
[76 秒推廣影片](dist-all/v.1.0.2-20261013/promo/kol-v.1.0.2-20261013-promo.mp4)
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
| F9 | 獨立開關配樂，預設開啟 |
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

### HD 畫面

HD採1920×1200原比例，包含核准的操作介面、原生中文字、六人肖像與旅店老闆頭像。
下方保留已核准門面及戰鬥畫格，旅店與設定來自本版實際封包；戰場、傷勢圖及未展示子頁仍沿現有顯示。

手繪旅店門面與六人隊伍：

![HD 旅店門面、六人肖像與操作介面，1920×1200](docs/verification/hd-interface-runtime-20261012/door-hd.png)

![本版HD旅店頭像、原生中文字與配樂控制](docs/verification/release-v1.0.2-20261013/inn-hd-controls.png)

| 戰鬥資訊框的 HD 人物 | 遊戲內 HD 畫面設定 |
|---|---|
| ![戰鬥資訊框與七個HD指令](docs/verification/hd-interface-runtime-20261012/battle-hd.png) | ![本版遊戲內HD畫面設定](docs/verification/release-v1.0.2-20261013/game-menu-hd.png) |

按 F10 或點下緣「畫面 F10」，即可在同一場景切換 EGA／HD；Esc 返回遊戲。
三組[同主題、同場景比較](docs/verification/hd-interface-runtime-20261012.md#同主題比較)沿用相同人物與門面。
圖片來源、畫格、SHA-256與權利分類見[正式介面截圖清冊](docs/verification/hd-interface-runtime-20261012/manifest.json)。
新版旅店及設定畫面來自實際AppImage，來源見[新版截圖清冊](docs/verification/release-v1.0.2-20261013/manifest.json)。

### 開發版：書本與操作子頁 HD

開發版已接入手繪書頁、三個服務頭像、原生中文與編輯器控制。下列為同一程式、同一狀態的EGA／HD比較，尚未納入現行發行包。

| EGA | HD |
|---|---|
| ![EGA角色選項](docs/verification/hd-native-ui-runtime-20261014/menu-ega.png) | ![HD手繪書頁與角色選項](docs/verification/hd-native-ui-runtime-20261014/menu-hd.png) |
| ![EGA甲冑店清單](docs/verification/hd-native-ui-runtime-20261014/town-shop-ega.png) | ![HD甲冑店兄弟與清單](docs/verification/hd-native-ui-runtime-20261014/town-shop-hd.png) |
| ![EGA法術編輯頁](docs/verification/hd-spell-scroll-runtime-20261014/editor-ega.png) | ![HD法術卷軸與十三欄編輯](docs/verification/hd-spell-scroll-runtime-20261014/editor-hd.png) |

[十五組比較頁](docs/verification/hd-native-ui-runtime-20261014/review.html)另含建角、編輯器、訓練、馬廄、法術與任務清單。
遊戲內F10切換、鍵鼠、兩種視窗與存讀檔已驗證。自訂點陣圖保留內容與色值；NPC、坐騎、裝備及動作圖示的後續結果見下方各批驗收，
範圍及來源見[正式驗收](docs/verification/hd-native-ui-runtime-20261014.md)。
法術編輯頁亦已[接入並驗證](docs/verification/hd-spell-scroll-runtime-20261014.md)，名稱輸入、左右選單與遊戲內F10切換保持狀態。

### 開發版：八十張角色頭像

八十張角色模板已補齊手繪HD，原核准六圖保留。以下為正式重製程式在同一角色狀態下的比較，尚未納入現行發行包。

| 重製中文 EGA | 手繪 HD |
|---|---|
| ![EGA建立角色頭像](docs/verification/hd-character-templates-20261014/portrait-000-ega.png) | ![HD建立角色頭像](docs/verification/hd-character-templates-20261014/portrait-000-hd.png) |
| ![EGA有翼角色頭像](docs/verification/hd-character-templates-20261014/portrait-044-ega.png) | ![HD有翼角色頭像](docs/verification/hd-character-templates-20261014/portrait-044-hd.png) |

[五組對比](docs/verification/hd-character-templates-20261014/review.html)另含弓手與矮人。遊戲內F10可切換，16字元姓名完整縮放於原框；新增角色的頭像可隨v3保存重載。
人物裝備及NPC／坐騎的後續接入見下方各批驗收；角色模板的範圍、來源與限制見[本批驗收](docs/verification/hd-character-templates-20261014.md)。

### 開發版：68個原生HD動作圖示

68個動作圖示已沿同一銅色主題接入開發版。下方展示相同圖號的來源PNG與正式原生HD，來源展示由重製程式繪製。
第28圖是城鎮徽章遮罩，保持資料用途。完整裝備頁的後續接入見下節；這批尚未納入現行發行包。

| 來源PNG圖示展示 | 正式原生HD圖示 |
|---|---|
| ![68個來源PNG動作圖示，由重製程式繪製](docs/verification/hd-actions-runtime-20261015/source-symbols.png) | ![同主題68個原生HD動作圖示](docs/verification/hd-actions-runtime-20261015/native-symbols.png) |

[比較頁](docs/verification/hd-actions-runtime-20261015/review.html)另含正常旅店的繁中EGA／HD。遊戲內F10、兩種視窗與存讀檔已驗，圖示切換沿原控制器；來源與完整範圍見[正式驗收](docs/verification/hd-actions-runtime-20261015.md)。

### 開發版：完整物品與隊伍頁

完整物品17槽與六型隊伍已接入正式HD呈現，包含身體、裝備、武器、盾牌、手部遮擋與城鎮／野外背板。
下圖來自正常標題操作，六人由旅店招募；兩側沿相同角色、裝備及隊伍狀態。尚未納入現行發行包。

| 重製中文 EGA | 正式 HD |
|---|---|
| ![女性精靈物品17槽EGA](docs/verification/hd-equipment-runtime-20261015/items-ega.png) | ![正式HD物品頁、完整裝備與原生中文](docs/verification/hd-equipment-runtime-20261015/items-hd.png) |
| ![正常招募六型隊伍EGA](docs/verification/hd-equipment-runtime-20261015/party-ega.png) | ![正式HD六型隊伍與原遮擋順序](docs/verification/hd-equipment-runtime-20261015/party-hd.png) |

[比較頁](docs/verification/hd-equipment-runtime-20261015/review.html)另含卸裝前臂、固定腰帶與預設／自訂盾面。
已知預設盾面匹配手繪HD，修改盾面保留原40×38點陣；原角色資料與v3格式保持。
中文、EGA回歸、兩種視窗的鍵鼠F10及正常存讀檔已核對，範圍見[正式驗收](docs/verification/hd-equipment-runtime-20261015.md)。
地圖、戰場、天候與未驗操作分支仍續作。

### 開發版：七城 NPC 與四種坐騎

七城102張NPC與四種坐騎已補齊手繪HD，四張既有NPC保留。下方來自正式程式的相同受控狀態，包含繁中對話與坐騎資料頁；尚未納入現行發行包。

| 重製中文 EGA | 正式 HD |
|---|---|
| ![七城NPC對話EGA](docs/verification/hd-npc-mounts-runtime-20261015/npc-t7-p10-ega.png) | ![手繪NPC與繁中對話HD](docs/verification/hd-npc-mounts-runtime-20261015/npc-t7-p10-hd.png) |
| ![重型戰馬人物資料頁EGA](docs/verification/hd-npc-mounts-runtime-20261015/mount-sheet-3-ega.png) | ![完整馬身與中文名牌HD](docs/verification/hd-npc-mounts-runtime-20261015/mount-sheet-3-hd.png) |

[七組比較](docs/verification/hd-npc-mounts-runtime-20261015/review.html)另含馬廄、雙人物及原構圖的橫置頭部。
110個場景的中文、圖片、EGA回歸及HD一色階核對通過；遊戲內F10可在同一場景切換，兩種視窗、鍵鼠與正常存讀檔已驗。
來源與完整範圍見[正式驗收](docs/verification/hd-npc-mounts-runtime-20261015.md)。地圖、戰場、傷勢、天候、片頭與其餘事件仍續作。

### 開發版：32種怪物

32種怪物的資訊框與遭遇群像已接入手繪HD。下方由正式程式繪製，兩側使用同一受控狀態；尚未納入現行發行包。

| 重製中文 EGA | 正式 HD |
|---|---|
| ![怪物資訊框EGA](docs/verification/hd-monsters-runtime-20261015/monster-10-info-01-ega.png) | ![手繪怪物資訊框HD](docs/verification/hd-monsters-runtime-20261015/monster-10-info-01-hd.png) |
| ![霧地幼蟲遭遇群像EGA](docs/verification/hd-monsters-runtime-20261015/monster-31-encounter-12-ega.png) | ![12隻霧地幼蟲手繪HD與透明遮擋](docs/verification/hd-monsters-runtime-20261015/monster-31-encounter-12-hd.png) |

[八組比較](docs/verification/hd-monsters-runtime-20261015/review.html)另含牛頭怪、獨眼巨人、小魔怪與火焰人形。128場景的中文、原生圖片、透明遮擋與EGA回歸已核對；遊戲內F10可在同一場景切換，兩種視窗、鍵鼠及正常存讀檔已驗。
來源、完整提示與驗收範圍見[正式收據](docs/verification/hd-monsters-runtime-20261015.md)。戰場、傷勢、地圖、天候、片頭及其他HD仍續作。

### 開發版：七城門面與城門林地

七城102張門面及城門林地已接入手繪HD，原核准旅店保持。以下為正式重製程式在同一準備門位的EGA／HD，尚未納入現行發行包。

| 重製中文 EGA | 正式 HD |
|---|---|
| ![石造門面EGA](docs/verification/hd-town-facades-runtime-20261015/twn1pics-003-ega.png) | ![手繪石造門面HD](docs/verification/hd-town-facades-runtime-20261015/twn1pics-003-hd.png) |
| ![海邊樹屋EGA](docs/verification/hd-town-facades-runtime-20261015/twn2pics-021-ega.png) | ![海邊手繪樹屋HD](docs/verification/hd-town-facades-runtime-20261015/twn2pics-021-hd.png) |
| ![住宅中文告示EGA](docs/verification/hd-town-facades-runtime-20261015/twn5pics-019-ega.png) | ![手繪住宅與原生中文告示HD](docs/verification/hd-town-facades-runtime-20261015/twn5pics-019-hd.png) |

[八組比較](docs/verification/hd-town-facades-runtime-20261015/review.html)另含工坊、石塔、水車、城門與林地。
103場景中文與畫面已核對；F10可在遊戲內切換，四場景兩種視窗的鍵鼠、完整城鎮狀態及v3存讀檔通過。
圖中其餘六城徽章仍沿EGA，與地圖、戰場、天候等後續HD工作分開追蹤；完整範圍見[正式驗收](docs/verification/hd-town-facades-runtime-20261015.md)。

### 事件畫面：原版與 HD

七段結局已接入開發版，沿用手繪、紅黑銅框與原生 HD 中文。現行 `v.1.0.2-20261013` 發行包保持，尚未包含本批。
下方包含結局內容。原版由 dosgolem 受控執行擷取；HD 來自正式重製程式的受控場景入口。

<details>
<summary>展開七段結局的原版／HD 截圖</summary>

| 原版 DOS，320×200 | 開發版 HD，1920×1200 |
|---|---|
| ![原版第0段：巨石門](docs/verification/hd-completion-20261014/original-ending-0.png) | ![HD第0段：巨石門](docs/verification/hd-ending-runtime-20261014/ending-0-hd.png) |
| ![原版第1段：火炬洞窟](docs/verification/hd-completion-20261014/original-ending-1.png) | ![HD第1段：火炬洞窟](docs/verification/hd-ending-runtime-20261014/ending-1-hd.png) |
| ![原版第2段：牢門](docs/verification/hd-completion-20261014/original-ending-2.png) | ![HD第2段：牢門](docs/verification/hd-ending-runtime-20261014/ending-2-hd.png) |
| ![原版第3段：受囚的騎士](docs/verification/hd-completion-20261014/original-ending-3.png) | ![HD第3段：受囚的騎士](docs/verification/hd-ending-runtime-20261014/ending-3-hd.png) |
| ![原版第4段：獲救](docs/verification/hd-completion-20261014/original-ending-4.png) | ![HD第4段：獲救](docs/verification/hd-ending-runtime-20261014/ending-4-hd.png) |
| ![原版第5段：陰影](docs/verification/hd-completion-20261014/original-ending-5.png) | ![HD第5段：陰影](docs/verification/hd-ending-runtime-20261014/ending-5-hd.png) |
| ![原版第6段：離開牢房](docs/verification/hd-completion-20261014/original-ending-6.png) | ![HD第6段：離開牢房](docs/verification/hd-ending-runtime-20261014/ending-6-hd.png) |

</details>

[完整比較頁](docs/verification/hd-ending-runtime-20261014/review.html)另附各段繁中 EGA。
七段中文、兩種視窗及48個實際 F10 設定事件通過，切換保留段號與狀態，HD往返完整畫格最多一色階。
正常住宿保存與新程序重載通過。本批不代表正常通關或整頁原版一致；
來源、畫格、權利、雜湊與限制見[結局驗收](docs/verification/hd-ending-runtime-20261014.md)。

### EGA 畫面

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

開發版已接入核准的四首原創中世紀室內樂：[標題與旅店](docs/verification/hd-music-20261012/review.html)、
[探索與戰鬥](docs/verification/hd-adventure-music-20261013/review.html)。試聽頁附OGG、分軌、三輪循環與來源授權。
配樂預設開啟，按F9或點下緣「配樂」獨立切換；切換EGA／HD時持續播放，開關只在本次遊玩有效。
接入與驗證見[配樂收據](docs/verification/hd-music-runtime-20261013.md)。現行發行包包含四首配樂及已驗收雷聲。

首批手繪旅店門面與六人肖像已接入實際遊戲，採 1920×1200 原比例。
點下緣「畫面 F10」或按 F10，選擇 EGA／HD 即可在同一場景比較；Esc 返回遊戲。
預設使用 EGA，切換在本次遊玩有效。三平台完整版已含首批素材。
角色按實際圖像匹配，自訂肖像沿用原圖；戰鬥資訊框顯示目前選中的人物。
中文、滑鼠點選、舊 v3 與住宿存讀檔已通過[首批驗收](docs/verification/hd-runtime-20261011.md)。
遊戲內切換的核准版面、中文、鍵鼠、縮放及正常住宿存讀檔已通過[正式驗收](docs/verification/display-switch-runtime-20261012.md)。

首批素材位於 [assets/hd/first-slice/manifest.json](assets/hd/first-slice/manifest.json)，
核准樣張保存在[比較原型](docs/verification/hd-first-slice-20261011/review.html)。
核准的[操作介面與NPC原型](docs/verification/hd-ui-theme-20261012/review.html)已接入開發版，
[正式驗收](docs/verification/hd-interface-runtime-20261012.md)包含HD中文、面板、操作列、頭像、F10及正常住宿存讀檔。
全畫面 HD 續作涵蓋書本、建角、商店與其他服務子頁、人物裝備、城鎮／世界地圖、戰鬥、傷勢、天候及事件。
[13類畫面盤點](docs/verification/hd-completion-20261014.md#目前覆蓋)逐項列出缺口與驗收條件；放大原版素材不列為手繪 HD 完成。
七段結局已依核准樣張接入開發版；其他事件及介面按清單續作，均須驗中文、畫面、鍵鼠及遊戲內 F10。
書頁、服務面板、三個服務頭像及編輯器已完成[正式原生HD比較](docs/verification/hd-native-ui-runtime-20261014/review.html)；[法術卷軸](docs/verification/hd-spell-scroll-runtime-20261014/review.html)及[八十張角色模板](docs/verification/hd-character-templates-20261014/review.html)亦已接入。其餘圖形續作。
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
  License 1.1），HD字模使用Noto Serif CJK TC；配樂試聽使用FluidR3 GM音色庫（MIT）。執行期使用 **Ebitengine** 及其相依套件。

原版執行檔、資料、美術、地圖、音樂、音效與轉換後素材均受獨立權利限制；
程式碼授權不代表可散布這些內容。
