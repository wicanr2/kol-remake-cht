# 繁體中文介面（雙模式）

更新日期：2026-09-01

原版字模格是 8×6、字距 8，繁體中文字形放不進同一格。使用者於 2026-09-01 決定採
**雙模式**：

- **英文模式**（`-lang en`，預設）：320×200、原版抽出的 8×6 字模，用來和 DOS EGA
  原版做嚴格逐像素對拍。這是 `docs/verification/remake-parity/` 那三份收據的模式。
- **中文模式**（`-lang zh-Hant`）：畫布放大到 640×400，書本美術以最近鄰放大，文字改用
  16×18 的中文字模格。**這個模式不對拍**——原版沒有中文畫面，版面是 remake 自訂的。圖像層的右頁說明在中文模式另給一組行
（`book.CJKPictureInstructions`）：原版把一句話拆成五行是英文字距下的排版，
逐行硬譯會讀成「選擇角色／圖像」這種斷句。

兩個模式共用同一個角色建立狀態機，差別只在版面與字型。

## 字型從哪裡來

`tools/build_cjk_font.py` 把中文 UI 會用到的字元從系統 TrueType 字型烘成點陣字模表：
用 ffmpeg 的 drawtext（libfreetype）在固定格點上畫出每個字元，再以固定門檻二值化。
執行期只讀烘好的 PNG，不在執行期做字型排版，和原版抽出的英文字型走同一條路。

目前烘焙參數：`Noto Sans CJK TC`、字級 15、門檻 96、字模格 16×18，共 344 個字元
（全部可見 ASCII ＋ 中文 UI 用到的漢字）。漢字是全形固定一格；ASCII 是比例字，字寬取
墨跡右緣加一個間隙——用半形固定字距畫會把相鄰字模切掉。

**授權**：字模表衍生自 Noto Sans CJK（SIL Open Font License 1.1）。散布時必須隨附
授權全文。字型著作權屬原作者，這一點不因烘成點陣而改變。

## 翻譯來源與狀態

字串表在 `docs/i18n/newchar-zh-Hant.json`，狀態沿用術語表的三級：核准／暫譯／待查。
種族、屬性、`reroll` 等詞取自 `docs/manual/glossary.md` 與 SPEC-005 已固定的屬性契約；
職業名多為暫譯，精靈教團專名沿用說明書作法保留拉丁拼寫。查不到譯文的字串會原樣顯示
英文，不猜。

## 文字安全矩形檢查

`text-fit.json`：122 項，0 項失敗。

檢查用實際字模表的逐字字寬計算每個字串的寬度，逐條要求：

- 每個字元在字模表裡都有字模（缺字會被畫成空白，那是漏字，不是排版問題）。
- 字串畫在指定起點時右緣不超出所屬的文字安全矩形。
- 名字欄 16 個字元的最寬情形仍在左頁框內（因此中文模式把名字另起一行畫在整頁寬度上，
  和原版英文版面一樣）。
- 十列屬性、右頁最長的 12 項清單、六項角色選項選單、16 槽角色清單的最後一列底緣，
  都仍在文字安全矩形內。書頁底緣的裝飾從原版第 190 列開始，放大兩倍即 380，
  安全矩形到 y=380 為止。

版面座標由 `cmd/kol -dump-cjk-layout` 匯出，因此檢查用的座標不會和程式裡的版面表脫節。

## 產生方式

```
python3 tools/build_cjk_font.py --strings docs/i18n/newchar-zh-Hant.json \
    --out assets/ui-font --name zh-hant-16 --family "Noto Sans CJK TC" \
    --size 15 --cell 16 --cell-height 18
tools/run_kol_headless.sh -dump-cjk-layout workplace/cjk-layout.json
python3 tools/verify_text_fit.py --layout workplace/cjk-layout.json \
    --font assets/ui-font/manifest.json --strings docs/i18n/newchar-zh-Hant.json \
    --data assets/original-derived-data/character-creation.json \
    --out docs/verification/cjk-ui/text-fit.json
tools/run_kol_headless.sh -lang zh-Hant \
    -script "z e d RET RET RET RET DOWN RET" -shot-dir docs/verification/cjk-ui/shots
tools/run_kol_headless.sh -lang zh-Hant -delete-script "DOWN RET y" \
    -shot-dir docs/verification/cjk-ui/shots/delete
```

字模表的烘焙要用 `tools/run_cjk_font_build.sh`（有界容器＋主機字型目錄唯讀掛載）；
新增中文字串之後要重烘，否則新字沒有字模會被畫成空白。**譯文不只住在字串表**：
城鎮畫面的對拍樣本自帶預先斷行的中文台詞，因此烘焙要一併帶 `--extra-json`
指向那些樣本。

`shots/` 是名字→種族→性別→職業→擲屬性→圖像層的中文畫面，加上角色選項頁
（`menu.png`）、角色清單頁（`roster.png`）、刪除流程（`delete/`）、HARBOR MASTER
訊息頁（`harbor.png`）、旅店畫面（`inn.png`、`inn-recruited.png`）與戶外地圖
（`outdoor.png`），供版面審閱。
刪除確認視窗在中文模式是原版視窗放大兩倍再加高，讓兩行 16×16 中文字上下各留 8 px；
文字安全矩形就是視窗自己的內緣，一併進 `text-fit.json` 檢查。

## 這份收據不宣稱什麼

- 不宣稱中文版面與原版對拍。原版沒有中文畫面，這個模式的版面是 remake 自訂的。
- 不宣稱譯文已定稿。多數職業名是暫譯，仍要由說明書其他段落或原版畫面交叉確認。
- 不宣稱其他畫面已中文化。目前是 New Character 頁、角色選項頁、角色清單頁、
  旅店與戶外地圖；世界地圖與戰鬥的畫面本身都還沒實作。
- 角色名字與所在地是原版資料，維持原文不譯。
