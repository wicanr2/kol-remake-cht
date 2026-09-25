# SPEC-035：法術編輯器的繁中頁面

狀態：**READY**（2026-09-25，內容依 RE-128，已接線並有正常按鍵畫格；選單位置是重製版自訂）。

## 缺陷與證據

中文模式的 `InnCJK` 原本沒有處理 `Inn.SpellEdit`：進入 SCROLL 的「修改法術」後，畫面仍是
建築對話框，只剩兩行造價數字，13 個欄位都看不到。英文頁 `drawSpellEdit` 與原版對拍畫格
逐位元相同（`docs/verification/remake-parity/shots/town-spell-edit/editor.png`），內容與
欄位語意見 RE-128（`sub_178E0`、`sub_1874E`、`sub_18350`、KOLDATA 0x4D）。

## 契約

1. 版面是英文頁乘 `book.CJKScale`：同一張卷軸底圖、透明繪字、六條分隔線、13 個欄位與
   兩行造價都在原座標乘 3。游標那一欄用 `ColourSpellCursor`。
2. 欄名取 KOLDATA 0x4D 的譯文，用說明書第 7 章的術語：種族、亞種、屬性、距離、持續、強度。
   英文原文靠空白把 RANGE／DUR／DMG 對齊到三個欄位；中文字寬不同，譯文對這三個字各給
   `~Hxx`，座標正好是原文三個字的起始欄。`tools/verify_ui_translation.py` 只允許這種
   「等於原文某字起始欄」的額外欄座標。
3. 欄位值查顯示譯名（SPEC-033）：MAGIC.0 的 class、stat、subclass 與 `All`；距離取
   KOLDATA 0x50／0x51 譯文；骰式與造價是數字，照原樣。
4. 法術名保留原文並以實際字寬置中在原名稱欄，底下一行畫 SPEC-034 注解；輸入名稱時
   名稱欄顯示已打的字加 `_`，不畫注解。
5. 兩行造價的標籤與內容對調是原版行為（RE-128），中文照樣保留。
6. 欄位選單（英文頁目前也沒畫）在中文頁畫成黑底清單框，放在被編輯欄位的另一側、欄名列
   下方，讓欄名與正在修改的欄位保持可見；列數、捲動與游標沿用 `town.ListRows` 清單規則。
   這是重製版版面，不對拍。
7. 英文模式不變。

## 驗收

- 收據：[spell-pages-receipt.json](../verification/cjk-ui/spell-pages-receipt.json)，
  編輯頁、修改清單、選單、名稱輸入四張畫格經目視檢查。同一輸入的英文模式畫格與既有
  對拍畫格逐位元相同。
- 升 CONFORMED 前：選單選值後寫回、存檔／扣款往返的英中同狀態比較。
