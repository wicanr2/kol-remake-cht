# 《武士傳說》繁體中文說明書

本目錄把使用者自備的 *Knights of Legend Player's Handbook* 與 Reference Card 逐頁翻譯
為繁體中文 Markdown。英文 PDF 與逐頁文字層不進 Git；每一頁以來源 PDF SHA-256、實體
PDF 頁碼及文字層 SHA-256 回查。

## 目前狀態

- 玩家手冊：79 個實體 PDF 頁；第 1–70 頁已逐頁對照並翻譯，第 71–75、78–79 頁待處理。
- 參考卡：5 個實體 PDF 頁，待翻譯。
- 正式狀態只有 `todo`、`translated`、`reviewed`：`translated` 只是初譯；只有逐頁比對
  原文、核對數字／專名／段落後才能標成 `reviewed`。
- 文字層來自 Poppler `pdftotext -raw`。它是 Adobe Paper Capture OCR 的衍生文字，已知有
  `Knights ofLegend`、`TIIE`、`charactzr` 等辨識錯誤；譯文依掃描頁面校正，不依 OCR 猜補。
- 玩家手冊實體頁 76、77 已由 120 DPI PNG 目視確認為空白跨頁（只有背頁透印），因此
  清冊標成 `expected_blank`；其餘 82 個實體頁都必須有非空文字層。

## 文件入口

- [逐頁定位與狀態](page-index.md)
- [術語表](glossary.md)
- [封面、製作緣起與目錄](chapters/00-front-matter.md)
- [開始遊戲與建立角色](chapters/01-getting-started.md)
- [在城鎮裡四處看看](chapters/02-poking-around-town.md)
- [在盔甲店](chapters/03-at-the-armorers.md)
- [進入荒野](chapters/04-into-the-wilderness.md)
- [短兵相接](chapters/05-coming-to-blows.md)
- [回到城鎮](chapters/06-back-in-town.md)
- [進階玩家](chapters/07-advanced-gamers.md)
- [阿什塔拉瑞亞王國](chapters/08-realm-of-ashtalarea.md)
- [傳說種族](chapters/09-races-of-legend.md)
- 來源清冊：`source-catalog.json`
- 翻譯狀態：`translation-status.json`

## 翻譯契約

1. 每個區塊以 `<!-- handbook/page/NNN -->` 或 `<!-- reference-card/page/NNN -->` 保存實體
   PDF 頁定位；原書印刷頁碼另行保留，兩者不可混用。
2. 保留章節、條列、表格、圖說與操作鍵。原版介面仍顯示英文時，首次出現採
   `英文（繁中）`，其後可用核准繁中術語。
3. 無法辨識或專名未定，以 `〔待查：…〕` 明示；不得默默補寫。
4. 說明書能證實公開規則與玩家流程；隱藏數值、精確執行順序與版本差異仍須回到原版
   executable／資料與可重現實驗。
5. 插圖目前不嵌入 Markdown。若未來引用掃描插圖，必須另做權利審查與來源定位。

## 重生來源清冊

在專案根目錄以既有 `civ1-manual-pdf:local` image 執行：

```sh
docker run --rm --network none --memory 2g --cpus 2 --pids-limit 256 \
  --user "$(id -u):$(id -g)" -e HOME=/tmp \
  -v "$PWD:/repo" -w /repo civ1-manual-pdf:local \
  python3 tools/manual_catalog.py /repo \
    docs/manual/source-catalog.json workplace/manual/source.local.tsv
```

`workplace/manual/source.local.tsv` 含英文全文，受 `.gitignore` 排除。版控內清冊只保存雜湊、
字數、行數與文字層是否非空。
