# RE-254：受阻打暈後的返回暫存與正常保存

日期：2026-10-05。狀態：DRAFT，有限正常鏈已驗，正式規則未變更。

沿[RE-253](20261005-blocked-knockback.md)已證實的正常 seed2 受阻持物樣本。
正式338輸入及Linuxc8764592保持。隔離候選去除掉落會保留ReturnScratch=6，
舊版則覆寫0；前輪原版沒有採集此值，不能由候選測試推定原版。

本輪重跑同一正常seed2、六人招募、旅行與首場REST，不改原版或模擬器。
只讀新增IDA linear word_39A54的DGROUP:BB84，H仍讀word_30D02的DGROUP:2E32。
原版EXE SHA-256沿RE-253固定輸入
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`，
IDA linear基底0x10000、DGROUP0x2DED0。工具仍是固定dosgolem
revision968ca774705853276ae55c8d1163c369481ac71f。

預先界限：首場20回合／2500鍵、戰後8確認、回城350方向、途中戰鬥120鍵、
住宿16確認。先核對前433狀態與已封存原版，再核第317格暫存與後續成本入口。
接著沿正常方向返回40／23、進城、住宿保存。沿已驗治療圖示選NO；
未知問句停止並保留。正常保存與新程序重載均需實際寫出／載入CHARDATA。
未通過之前不提升READY，也不把這條有限鏈當完整戰鬥、其他分支或全任務parity。

研究入口`workplace/blocked-knockback-return-20261005/prepare.py`、`probe.go`、
`build-original.sh`與`run-original.py`。沿固定Go映像與UID/GID1000:1000，
來源唯讀、明確研究輸出可寫、network none、有界資源與--rm。
前輪已查滑鼠輪詢記錄持續成長，本輪包含戰後旅途，以12GiB硬上限、
GOGC20及GOMEMLIMIT9GiB執行並保留gctrace，不修改dosgolem記錄器。
原始與衍生資料僅本機，不涉及HD版面、資料契約、版號、平台或發布。

## 有限結果

| 核對 | 結果與範圍 |
|---|---|
| 原版首場前綴 | 前433狀態與RE-253相同，新增唯讀欄位另列；289張前綴PNG逐檔相同 |
| 受阻打暈 | 第317格ReturnScratch為6，H為0，Olear手持16-byte精靈弓完整保留，位置保持，指令清0 |
| 首場返回 | 十四回合戰敗，正常確認後返回世界；時間1680至1580，有限成本100 |
| 正常回程 | 第二場遭遇後逃離，正常進城及住宿，共760原版狀態 |
| 原版保存 | 六次實際12288-byte CHARDATA寫入，只有角色槽0至5改變；新程序41狀態正常招募重載，六人持久欄保持 |
| 候選首場 | 六個雙語程序共2616狀態，觀察副本trace及完整PNG與控制程序相同；233點seed與ReturnScratch逐點相符 |
| 候選正常鏈 | 雙語候選及觀察副本共2548狀態，另26狀態實際v3新程序重載；六人每人17個原始carried物品槽、已登記持久欄位與原版相符 |
| 回程比較 | 每語言168個世界位置／次格／時刻／時段比較點、268個戰鬥seed與ReturnScratch比較點相符 |

原版CHARDATA SHA-256為
`bac5491f77ae802959a0fb7107e68292ae5365de7c72803a1cb3a6968154d54d`。
候選仍為RE-253的`876a22c87b6d4aace2e78410660806763b2d4e484d5f2070caf52c21129cbf18`。
正式338來源與Linux完整SHA
`c87645925936137788443b6eda8b753842c0a725cbf4d7dd9139a631ffa302e5`
保持。

原版新程序沒有重播種、狀態注入、direct-entry、等待替換或寫回CHARDATA。
候選v3雙語完整保存只排saved_at相同，觀察副本保存亦相同。
原版每人17個原始carried槽共102槽，不把typed裝備的重複呈現另算新槽。
`review-original-save.py`、`review-prefix.py`、`review-remake.py`與各自JSON保存比較範圍。

## 等效操作與失敗分類

首版前綴檢查誤把book-final.png當固定前綴。該圖由startup的defer在探測結束時擷取，
前輪終點在戰敗頁，本輪在住宿末端，兩者不能作同狀態圖片比較。
保留首稿及拒絕JSON，修正範圍後重核既有輸出，沒有重跑原版或挑換種子。

候選V1的702鍵腳本在第536個SPACE被野外輸入拒絕。
隔離正常輸入診斷器精確定位到原版第592格。
兩側首場返回及第二場遭遇、逃離均已對齊；原版回程額外SPACE只把共用
ReturnScratch由1改0，角色、地上槽、亂數、位置、時間等其他已採集欄保持。
V1程序、plan、準備腳本與分類保存於rejected-plan-v1/。
V2省略這類額外SPACE，但逐筆保留scratch差分，不宣稱世界暫存或操作parity。
首個診斷包裝誤用映像沒有提供的xdpyinfo，未啟動遊戲；改用既有X socket就緒契約。

V2共636鍵／637狀態。另有一處原版Right確認逃離結果改用現行RET，
原版第581格與候選第530步明示為等效確認，沒有改正式按鍵語意。
仍是正常啟動、固定seed2、不重播種、無遭遇骰覆寫、無座標或狀態注入。

## 限制與下一閘門

原版舊H訊息仍寫Short Sword，當時手持為Elf Bow且地上槽全零。
候選受阻分支不猜補名稱；H語意與全滿地上槽仍未知。
原版戰敗訊息是CYCLOPS，候選為CYCLOPSEs，另列複數呈現差異。

RE-256追加勘誤：上句原圖名稱誤記。正常第432格顯示CYCLOPES，候選第435步為CYCLOPSES。
0x21524退輸出欄位後覆寫E，0x178接s have；詳見[RE-256](20261005-monster-plural-ending.md)。
原始PNG與RE-254歷史收據保持，不改其受阻規則DRAFT或保存驗收範圍。
原版住宿末端隊伍清空，候選保留六人；既有方向確認與滑鼠決定仍待答覆。
沒有完整50-byte記錄、42別名、整頁畫面、全旅途亂數、長任務、真機或封包完成聲明。

受阻持物候選的返回暫存及這條正常戰後保存限制已由本輪有限證據解除。
正式規則仍DRAFT，待確認是否採原版受阻持物保持，再做READY審查與正式回歸。
可檢閱本機[同狀態原型](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/blocked-knockback-return-20261005/review.html)。
歷史RE-253封存收據維持，不回寫其當輪限制。

封存入口：[有限收據](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/blocked-knockback-return-20261005/receipt.json)、
[完整重驗](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/blocked-knockback-return-20261005/verify.py)。
PNG逐檔保存來源、程序、畫格、雜湊、尺寸與僅本機權利分類。
現行及舊SPEC-008／RE-253的不可變鍵勘誤由backlink-review.json與重驗器核對。
