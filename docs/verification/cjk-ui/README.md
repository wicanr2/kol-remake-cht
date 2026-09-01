# 繁體中文介面（雙模式）

更新日期：2026-09-01

原版字模格是 8×6、字距 8，繁體中文字形放不進同一格。使用者於 2026-09-01 決定採
**雙模式**：

- **英文模式**（`-lang en`，預設）：320×200、原版抽出的 8×6 字模，用來和 DOS EGA
  原版做嚴格逐像素對拍。這是 `docs/verification/remake-parity/` 那三份收據的模式。
- **中文模式**（`-lang zh-Hant`）：畫布放大到 640×400，書本美術以最近鄰放大，文字改用
  16×18 的中文字模格。**這個模式不對拍**——原版沒有中文畫面，版面是 remake 自訂的。

兩個模式共用同一個角色建立狀態機，差別只在版面與字型。

## 字型從哪裡來

`tools/build_cjk_font.py` 把中文 UI 會用到的字元從系統 TrueType 字型烘成點陣字模表：
用 ffmpeg 的 drawtext（libfreetype）在固定格點上畫出每個字元，再以固定門檻二值化。
執行期只讀烘好的 PNG，不在執行期做字型排版，和原版抽出的英文字型走同一條路。

目前烘焙參數：`Noto Sans CJK TC`、字級 15、門檻 96、字模格 16×18，共 199 個字元
（全部可見 ASCII ＋ 中文 UI 用到的漢字）。

**授權**：字模表衍生自 Noto Sans CJK（SIL Open Font License 1.1）。散布時必須隨附
授權全文。字型著作權屬原作者，這一點不因烘成點陣而改變。

## 翻譯來源與狀態

字串表在 `docs/i18n/newchar-zh-Hant.json`，狀態沿用術語表的三級：核准／暫譯／待查。
種族、屬性、`reroll` 等詞取自 `docs/manual/glossary.md` 與 SPEC-005 已固定的屬性契約；
職業名多為暫譯，精靈教團專名沿用說明書作法保留拉丁拼寫。查不到譯文的字串會原樣顯示
英文，不猜。

## 文字安全矩形檢查

`text-fit.json`：98 項，0 項失敗。

檢查用實際字模表的逐字字寬計算每個字串的寬度，逐條要求：

- 每個字元在字模表裡都有字模（缺字會被畫成空白，那是漏字，不是排版問題）。
- 字串畫在指定起點時右緣不超出所屬的文字安全矩形。
- 名字欄 16 個字元的最寬情形仍在左頁框內。
- 十列屬性的最後一列底緣仍在左頁框內。

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
```

`shots/` 是名字→種族→性別→職業→擲屬性→圖像層的中文畫面，供版面審閱。

## 這份收據不宣稱什麼

- 不宣稱中文版面與原版對拍。原版沒有中文畫面，這個模式的版面是 remake 自訂的。
- 不宣稱譯文已定稿。多數職業名是暫譯，仍要由說明書其他段落或原版畫面交叉確認。
- 不宣稱其他畫面已中文化。目前只有 New Character 頁；城鎮、世界、戰鬥都還沒有。
