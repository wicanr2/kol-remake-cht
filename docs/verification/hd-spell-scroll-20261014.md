# 法術編輯頁HD卷軸原型

狀態：DRAFT隔離原型。沿[SPEC-039](../spec/039-hd-art-ui.md#法術卷軸後續原型)及既有手繪、1920×1200與欄位位置，重繪`KOLPICS[1]`空白卷軸。
來源基底90783017，374份Go／module保持；原工作區`workplace/hd-spell-scroll-20261014/`的計畫、生成提示、覆蓋檔與固定輸入已封存核對後清理。
四組編輯、名稱輸入及左右選單每側107狀態與文字相同；[比較頁](hd-spell-scroll-20261014/review.html)保存16張畫格與可讀性修正。
真實視窗18個F10事件、輸入緩衝、滑鼠None選項與兩視窗通過，三組完整畫格往返零差；結果見[GUI收據](hd-spell-scroll-20261014/gui-check.json)。
啟動時產生的v3轉換檔與無輸入基底對照只差saved_at，工作副本沒有提交；不宣稱Ctrl+Q會自動保存。
法術規則、名稱輸入契約、教派限制及kol-save-v3保持；全畫面HD仍開放。

原圖來源為鎖定ZIP的`Knights-of-legend/KOLPICS.0`，SHA-256為`502367c613109eb326e57c54a727c8e3e414c1f8c3d73c9ad16497c453b8b402`。
該母體有14個payload、8個已驗PNG，第1圖為320×200；原PNG SHA-256為`44d92fa8eb399ba73df0b152ade737d72f691153c582cc6e68911f61177a7cd2`。
新卷軸1586×992，保留原構圖；首版造價第二行落到下卷軸陰影，已用內建imagegen延長明亮紙面，欄位位置保持。
來源及衍生PNG只限目前private儲存庫與本機完整版。

圖片、畫格、權利及雜湊見[清冊](hd-spell-scroll-20261014/manifest.json)，完整生成提示、覆蓋檔與失敗輸出由[封存收據](hd-spell-scroll-20261014/receipt.json)索引。
GUI末尾原誤判不得存在存檔；原碼在接受輸入前已轉換舊v1。已用無輸入基底證實轉換結果相同，重核完成的18個事件與畫格，沒有再重播GUI。
實際執行的腳本與後續正確檢查版本分開保存。正式接入由[後續收據](hd-spell-scroll-runtime-20261014.md)追蹤。
原型395份證據與三份程式雜湊另存[本機清冊](local/hd-spell-scroll-20261014/retention.json)，核對及解開沿[保存入口](retained-evidence.md)。
