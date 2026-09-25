# 遊戲內繁體中文文字

本目錄保存建角、城鎮與共用文字庫的譯表。英文原文可由固定原始資料重生；
譯表中的暫譯名稱依[術語表](../manual/glossary.md)標示，不代表原版中文官方名稱。

| 譯表 | 範圍 | 覆蓋檢查 |
|---|---|---|
| [建角](newchar-zh-Hant.json) | 欄位、種族、職業與介面 | `tools/verify_runtime_translation.py` |
| [城鎮](town-zh-Hant.json) | 建築名稱、進門與後續對話 | `tools/verify_runtime_translation.py` |
| [共用文字庫](ui-zh-Hant.json) | KOLDATA、BOOKDATA、ENDTXT | `tools/verify_ui_translation.py` |
| [資料名稱](display-names-zh-Hant.tsv) | 怪物、裝備、商店商品與任務物品的原名及顯示譯名 | `tools/verify_display_names.py` |
| [非文字槽位](town-nontext-slots.json) | 經 RE-140 判定不以文字顯示的城鎮訊息槽位，附原文快照與證據等級 | `tools/verify_runtime_translation.py` |

在專案根目錄執行 `tools/run_python_tool.sh tools/verify_runtime_translation.py`
及 `tools/run_python_tool.sh tools/verify_ui_translation.py`。這兩項只檢查靜態譯表；
資料名稱另執行 `tools/run_python_tool.sh tools/verify_display_names.py`；此表僅由
中文畫面的顯示端查詢，不得改寫原始名稱或規則鍵值。
正式「全中文」還需要逐條正常玩家路徑、程式生成文字、字模與文字安全矩形驗收。

城鎮訊息槽位與建築 payload 共用位移，馬廄坐騎表等資料會被抽取器解成看似
亂碼的「訊息」。這些槽位列在非文字槽位清單，不列入翻譯母體；清單的原文快照
與來源不符、多列或少列時檢查會失敗。其餘城鎮文字靜態覆蓋為 0 缺譯。
缺譯時程式會退回英文，因此靜態通過仍不能代替正常路徑與動態字串驗收。
這輪的來源數量、可見名稱缺口與手冊功能對照見
[手冊功能與繁中收尾稽核](../verification/manual-fidelity-audit-20260923.md)。
