# SPEC-003：繁體中文說明書翻譯與校對契約

- 狀態：`READY`
- 日期：2026-08-31
- 範圍：玩家手冊與參考卡的繁體中文 Markdown；不定義隱藏遊戲規則
- 一手來源：使用者自備的兩份固定 SHA-256 PDF
- 消費端：`docs/manual/`、`tools/manual_catalog.py`、`tools/verify_manual_translation.py`

## 輸入與定位

| 文件 | SHA-256 | PDF 實體頁 |
|---|---|---:|
| `knightslegend-manual.pdf` | `4c80f5d4…7086a` | 79 |
| `knightslegend-refcard.pdf` | `4815bc6c…abd3` | 5 |

逐頁不可變 key 為 `handbook/page/NNN` 或 `reference-card/page/NNN`。原書印刷頁碼另寫在
Markdown marker，不能取代 PDF 實體頁定位。來源清冊必須保存 PDF 雜湊、Poppler 版本、
逐頁文字 SHA-256、字元／單字／行數與頁面種類；英文全文只存 ignored 工作區。

## 擷取契約

1. 先使用 PDF 已有的 Adobe Paper Capture 文字層，以 Poppler `pdftotext -raw` 逐頁擷取。
2. 文字層只是 OCR 草稿，不可直接當翻譯或規則證據。每個 `reviewed` 頁必須對照掃描影像。
3. 只有「頁面有可見文字但文字層缺失」才啟動窄範圍 OCR；不得為空白頁製造文字。
4. 玩家手冊實體頁 76、77 已以 120 DPI PNG 目視確認為空白跨頁，只見背頁透印；清冊
   標成 `expected_blank`。其餘 82 頁都必須有非空文字層。
5. 任何 PDF 或逐頁文字雜湊漂移都必須停止並建立新來源身分，禁止覆蓋既有校對狀態。

## 翻譯狀態

每個實體頁恰屬一種狀態：

- `todo`：尚未完成初譯；
- `translated`：初譯完成，但未逐頁對照掃描；
- `reviewed`：數字／篇幅／英文殘留機械檢查通過，且已逐頁對照掃描影像、章節分界、
  圖說、表格與頁碼。

只有 `reviewed` 頁可列為正式 Markdown。狀態集合必須完整覆蓋 84 頁、不得重疊；空白頁
只有經目視確認後才可直接標為 `reviewed`。

## Markdown 與術語

1. 每頁以 `<!-- document/page/NNN｜印刷頁 … -->` 保留定位；跨章節時同一頁可在多個檔案
   出現，驗證器必須合併其區塊，不能只取第一段。
2. 保留章節層級、條列、表格、圖說、鍵名與交叉引用。插圖不直接嵌入；需要時另做權利審查。
3. 原版介面保留英文時，首次出現採 `English（繁中）`。專名依 `glossary.md` 分成核准、
   暫譯與待查；待查不得靠 OCR 字形自行定案。
4. 機械 gate 至少檢查：來源 key／hash、84 頁狀態覆蓋、reviewed 非空區塊、原文數字、
   中英篇幅異常及連續八個英文單字的疑似漏譯。
5. 機械通過不能取代語意校對；OCR 閱讀順序可能錯置。實體頁 3 的種族／職業目錄就是
   已知正對照：只讀 raw 文字會把矮人、凱爾登、精靈三組頁碼配錯。

## 規則證據邊界

說明書可作公開規則、操作與玩家流程的證據，但不能單獨證實 executable 的精確順序、
隱藏值、亂數消耗、bug 或版本差異。規則 spec 引用手冊時必須同列文件 SHA-256、實體頁與
印刷頁；若原版實驗衝突，保留衝突並回到 RE，不默默以譯文覆蓋。

## 權利邊界

翻譯 Markdown 與不含英文全文的雜湊清冊可提交 private Git。原始 PDF、逐頁英文全文、
掃描頁 PNG 與插圖不得提交或公開發布；中文譯本的公開散布權仍須在 release 前獨立審查。

## READY 審查

- [x] 兩份 PDF 的 SHA-256、頁數、尺寸與文字層工具已固定。
- [x] 84 頁逐頁 hash catalog 可重生，82 個內容頁均有文字層。
- [x] 兩個無文字頁已由掃描影像確認為空白跨頁。
- [x] 三態、頁面定位、Markdown、術語、機械 gate 與權利政策已明定。
- [x] 首批頁面揭露 raw OCR 錯序，並以掃描影像修正，形成正對照。

本規格授權逐頁翻譯、校對與 Markdown 產出；不授權玩法或 Go runtime 實作。全 84 頁完成
且逐頁 gate 通過後，才可升為 `CONFORMED`。
