# 繁體中文介面（雙模式）

更新日期：2026-09-23

原版字模格是 8×6、字距 8，繁體中文字形放不進同一格。使用者於 2026-09-01 決定採
**雙模式**：

- **英文模式**（`-lang en`）：320×200、原版抽出的 8×6 字模，用來和 DOS EGA
  原版做嚴格逐像素對拍。這是 `docs/verification/remake-parity/` 那三份收據的模式。
- **中文模式**（`-lang zh-Hant`，啟動預設）：畫布放大到 960×600，書本美術以最近鄰放大，文字改用
  16×18 的中文字模格。**這個模式不對拍**——原版沒有中文畫面，版面是 remake 自訂的。圖像層的右頁說明在中文模式另給一組行
（`book.CJKPictureInstructions`）：原版把一句話拆成五行是英文字距下的排版，
逐行硬譯會讀成「選擇角色／圖像」這種斷句。

兩個模式共用同一個角色建立狀態機，差別只在版面與字型。

## 字型從哪裡來

`tools/build_cjk_font.py` 把中文 UI 會用到的字元從系統 TrueType 字型烘成點陣字模表：
用 ffmpeg 的 drawtext（libfreetype）在固定格點上畫出每個字元，再以固定門檻二值化。
執行期只讀烘好的 PNG，不在執行期做字型排版，和原版抽出的英文字型走同一條路。

目前烘焙參數：`Noto Sans CJK TC`、字級 15、門檻 96、字模格 16×18，共 2,542 個字元
（全部可見 ASCII ＋ 中文 UI 用到的漢字）。漢字是全形固定一格；ASCII 是比例字，字寬取
墨跡右緣加一個間隙——用半形固定字距畫會把相鄰字模切掉。

**授權**：字模表衍生自 Noto Sans CJK（SIL Open Font License 1.1）。散布時必須隨附
授權全文。字型著作權屬原作者，這一點不因烘成點陣而改變。

## 翻譯來源與狀態

任務物品交付的正常按鍵、訊息 6／7、英中狀態與畫格收據見
[任務交付收據](quest-turnin-receipt.json)；這是從觀察用存檔起跑的重製版驗證，
原版同狀態仍待驗。
第 7 城問中 `nobj`、顯示兩頁任務敘述及接受問句的局部正常按鍵畫格見
[接任務收據](quest-offer-receipt.json)；起點是已指定城鎮位置的觀察用存檔。

書本畫面的標籤走 **`docs/i18n/ui-zh-Hant.json` 的 BOOKDATA 那一份**：那是三份 UI
文字庫 486 筆逐筆翻過也驗過的譯文，`internal/i18n` 的 `LoadBank` 把它疊到
`docs/i18n/newchar-zh-Hant.json` 上——剝掉 `~Vxx`／`~Hxx` 定位碼、按 `@` 切段、
去掉尾端的冒號與補齊寬度的空白，英文段就是書本標籤那個字串。中英段數對不上的
記錄整筆跳過，不猜哪一段配哪一段。

`newchar-zh-Hant.json` 留下的是 remake 自己加的字串（`New character`、`Select one`、
`CANCEL`／`ACCEPT` 這些原版沒有的）。查不到譯文的字串會原樣顯示英文，不猜。

兩份譯文本來各自為政，接起來之後有 21 個詞改用文字庫那一份，其中
`NAME` 從「名字」變「姓名」、`MALE`／`FEMALE` 從「男性／女性」變「男／女」、
精靈教團專名從拉丁拼寫變成中譯。`HIGHWAYMAN` 原本在文字庫裡誤譯成「築路工」
（把 highway 當成道路工程），已改成「攔路盜」。

狀態沿用術語表的三級：核准／暫譯／待查。職業名多為暫譯。

## 文字安全矩形檢查

`text-fit.json`：9,301 項，0 項失敗（含 387 筆資料譯名在五種欄位的 1,935 項）。

檢查用實際字模表的逐字字寬計算每個字串的寬度，逐條要求：

- 每個字元在字模表裡都有字模（缺字會被畫成空白，那是漏字，不是排版問題）。
- 字串畫在指定起點時右緣不超出所屬的文字安全矩形。
- **KOLDATA 譯文畫在城鎮對話框裡的每一行**都不超出框的內緣。只驗寬度不驗行數：
  一筆記錄不一定是「一次畫出來的一段」——部位名那一筆（0x8D）是十六個詞的名稱表，
  程式一次只取一行。索引 ≥ 0x100 的戰鬥訊息、含 `~Vxx` 自己跳位置的記錄
  （例如資料頁的標籤表 0x73）與野外標籤那一段（0xD5–0xE1，另有自己的框）
  排除在外。
- 名字欄 16 個字元的最寬情形仍在左頁框內（因此中文模式把名字另起一行畫在整頁寬度上，
  和原版英文版面一樣）。
- 十列屬性、右頁最長的 12 項清單、六項角色選項選單、16 槽角色清單的最後一列底緣，
  都仍在文字安全矩形內。書頁底緣的裝飾從原版第 190 列開始，放大三倍即 570，
  安全矩形到 y=380 為止。

版面座標由 `cmd/kol -dump-cjk-layout` 匯出，因此檢查用的座標不會和程式裡的版面表脫節。

## 產生方式

```
tools/run_cjk_font_build.sh --strings docs/i18n/newchar-zh-Hant.json \
    --extra-json docs/i18n/newchar-zh-Hant.json \
    --extra-json docs/i18n/ui-zh-Hant.json \
    --extra-json docs/i18n/town-zh-Hant.json \
    --extra-tsv docs/i18n/display-names-zh-Hant.tsv \
    --extra-json docs/i18n/ui-symbols-zh-Hant.json \
    --extra-json docs/i18n/spell-words-zh-Hant.json \
    --extra-json docs/re/town1-outdoor-sample.json \
    --extra-json docs/re/town-outdoor-sample.json \
    --extra-json docs/re/town-inn-sample.json \
    --extra-json docs/re/town-inn-sample-recruited.json \
    --out assets/ui-font --name zh-hant-16 --family "Noto Sans CJK TC" \
    --size 15 --cell 16 --cell-height 18
tools/run_kol_headless.sh -dump-cjk-layout workplace/cjk-layout.json
python3 tools/verify_text_fit.py --layout workplace/cjk-layout.json \
    --font assets/ui-font/manifest.json --strings docs/i18n/newchar-zh-Hant.json \
    --bank-strings docs/i18n/ui-zh-Hant.json --ui-i18n docs/i18n/ui-zh-Hant.json \
    --data assets/original-derived-data/character-creation.json \
    --town-i18n docs/i18n/town-zh-Hant.json \
    --town-sample docs/re/town-outdoor-sample.json \
    --town-sample docs/re/town1-outdoor-sample.json \
    --town-sample docs/re/town-inn-sample.json \
    --town-sample docs/re/town-inn-sample-recruited.json \
    --display-names docs/i18n/display-names-zh-Hant.tsv \
    --out docs/verification/cjk-ui/text-fit.json
tools/run_kol_headless.sh -lang zh-Hant \
    -script "z e d RET RET RET RET DOWN RET" -shot-dir docs/verification/cjk-ui/shots
tools/run_kol_headless.sh -lang zh-Hant -delete-script "DOWN RET y" \
    -shot-dir docs/verification/cjk-ui/shots/delete
for s in menu roster harbor inn outdoor; do
  tools/run_kol_headless.sh -lang zh-Hant -scene $s \
      -shot docs/verification/cjk-ui/shots/$s.png
done
tools/run_kol_headless.sh -lang zh-Hant -scene inn \
    -inn docs/re/town-inn-sample-recruited.json \
    -shot docs/verification/cjk-ui/shots/inn-recruited.png
```

字模表的烘焙要用 `tools/run_cjk_font_build.sh`（有界容器＋主機字型目錄唯讀掛載）；
新增中文字串之後要重烘，否則新字沒有字模會被畫成空白。

**上面那串 `--extra-json` 是字模表的完整輸入，少一個就會掉字**：`--strings` 只讀
建角字串表的五個鍵，而共用文字庫、城鎮台詞與四份畫面樣本另有用字。重烘之後先比對新舊
manifest 的碼位集合，確認**沒有任何字消失**，再重跑截圖與 `verify_text_fit.py`。

`shots/` 是名字→種族→性別→職業→擲屬性→圖像層的中文畫面，加上角色選項頁
（`menu.png`）、角色清單頁（`roster.png`）、刪除流程（`delete/`）、HARBOR MASTER
訊息頁（`harbor.png`）、旅店畫面（`inn.png`、`inn-recruited.png`）與戶外地圖
（`outdoor.png`），供版面審閱。
刪除確認視窗在中文模式是原版視窗放大三倍再加高，讓兩行 16×16 中文字上下各留 8 px；
文字安全矩形就是視窗自己的內緣，一併進 `text-fit.json` 檢查。

## 這份收據不宣稱什麼

- 不宣稱中文版面與原版對拍。原版沒有中文畫面，這個模式的版面是 remake 自訂的。
- 不宣稱譯文已定稿。多數職業名是暫譯，仍要由說明書其他段落或原版畫面交叉確認。
- 本節是 2026-09-01 的建角與城鎮畫面收據；後續世界與戰鬥畫面進度以目前
  `CONTEXT.md` 與各項新收據為準。
- 角色名字與所在地是原版資料，維持原文不譯。

## 2026-09-23：戰鬥法術清單抽驗

SPEC-032 的繁中法術與效果清單已從正常玩家按鍵路徑重播，兩張 960×600
畫格完成目視檢查：[法術清單](shots/battle-lists/spell/step-109.png)、
[正負效果清單](shots/battle-lists/spell/step-110.png)。來源雜湊、工具映像、
畫格雜湊、狀態與權利分類見[收據](battle-lists-receipt.json)；
[重播規格](battle-lists.spec)保存固定按鍵。

其後以固定 seed 39、六人存檔及正常按鍵重播，於第 22 回合勝利後取得
[找到的物品](shots/battle-lists/spoils/items-found.png)與
[角色清單](shots/battle-lists/spoils/characters.png)兩張 960×600 畫格；
中文標題及物品譯名已目視檢查。[勝利畫格收據](battle-spoils-winning-receipt.json)
記錄固定輸入、來源雜湊、畫格狀態與權利分類；
[按鍵序列](battle-spoils-winning.tokens)可重播。
這條路徑使用觀察用的怪物與任務地點資料，只證明重製版兩頁可達且可讀，
不代表原版同狀態戰鬥或自然資料的勝利機率。
續按 `N N RET` 可離開清單回到野外；[繁中](shots/battle-lists/spoils/exit-world-zh.png)
與[英文](shots/battle-lists/spoils/exit-world-en.png)模式的結束狀態軌跡逐位元相同。
同一路徑也揭出動態日期原先退回英文，已改由時鐘產生中文日期並目視檢查；
這仍不代表其他畫面的動態文字已全數翻譯。

## 2026-09-24：六人購馬與角色資料頁

從明示修改位置與金幣的六人觀察用存檔出發，同一組正式按鍵依序走過馬廄
[商品清單](shots/stable-six/shop-list-zh.png)、
[坐騎明細](shots/stable-six/steed-detail-zh.png)、
[購買後角色資料頁](shots/stable-six/sheet-zh.png)、
[物品換頁](shots/stable-six/items-zh.png)、
[隊伍立繪換頁](shots/stable-six/party-zh.png)及
[返回資料頁](shots/stable-six/returned-sheet-zh.png)。六張繁中畫格已目視檢查；
英文版[資料頁](shots/stable-six/sheet-en.png)與
[返回畫格](shots/stable-six/returned-sheet-en.png)供層次對照。

[收據](stable-six-page-receipt.json)記錄原始與觀察用存檔雜湊、修正版程式雜湊、
24 個按鍵、PNG 與軌跡雜湊。英中 25 筆狀態軌跡逐位元相同，
但起點位置與金冠經觀察用工具改動，沒有原版馬廄同狀態畫格；
這份收據只結案重製版的中文版面與換頁操作。
