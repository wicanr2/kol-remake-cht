# RE-211：法術名稱的正常按鍵輸入

日期：2026-10-03。沿用 [RE-210](20261003-sustained-spell-chain.md) 的失效與保存收據。
工作區為 `workplace/spell-name-input-20261003/`。

## 已核對事實與READY範圍

RE-210正常開名稱輸入後，BS／N兩個鍵的typing均維持Kelnalon。
`Session.Key`沒有呼叫既有 `TypeSpellName`／`BackspaceSpellName`；
既有單元測試直接呼叫這兩個helper，因此未發現正式路由缺口。
這是重製版接線缺陷，證據等級已證實。

既有SPEC-035已READY的名稱欄位契約包含顯示輸入文字與caret、輸入時隱藏注解，
確認寫入工作副本，取消保留原副本。RE-128記錄原版 `sub_1B5C8` 收12個字、
`sub_1C0F6`打包，原始名稱槽為16-byte法術的+0…7。
來源 `KOLTOWN-unpacked.EXE` SHA-256：
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
位址採IDA Pro 9.4 linear，DGROUP相對偏移另列，基底0x2DED0。

本次只恢復已存在的名稱文字函式接線：ASCII字母、空白、退格在名稱輸入時修改typing。
RET／圖示選取與取消沿用既有行為，12字上限與成本不改。
正式鍵盤的方向鍵在名稱文字模式不送移動token，字母U／D／L／R仍可輸入。
原版姓名大小寫、字元接受集與輸入畫面的嚴格同狀態仍未知，不藉接線修正擴大宣稱。
先保留舊碼正常按鍵反例，再依既有READY契約修改正式路由。

## 驗收閘門

正常按鍵須看見退格與新字母，確認／取消分開抽樣；12字上限、方向字母與欄位游標保持。
雙語正式與唯讀觀察正常文字修改／拒絕保存，以及名稱修改後正常住宿／正式重載。
RE-210自然失效圖、正常Strength修改保存鏈與持久欄位抽樣保持。
全部適用Go測試與清單守門通過後，只閉合本接線切片。
原版完整名稱輸入／住宿保存、整場戰鬥及SPEC-035其他未驗規格不自動升格。

## 本輪結果與限制

RE-211正常法術名稱輸入已接線，BS／N可見Kelnalon→Kelnalo→KelnaloN，取消保持原名與資源。舊碼兩個文字反例失敗，新碼全部internal／cmd測試通過；只改三份Go接線，新增正常按鍵回歸。雙語正常修改名稱／Strength後住宿各660鍵／661狀態，保存只變預期法術名稱，其他資料與RE-210相同；正式重載各16狀態、兩人34槽與14持久欄位相符。正式／觀察合計4936PNG及可比trace相同，雙語各360自然失效圖保持前版；取消除189／190／191輸入圖外各194圖相同。本接線有限CONFORMED，原版名稱字元集／大小寫／正常保存嚴格同狀態仍未知，整體SPEC-035保持READY。正式324份Go，候選a20728cb；怪物四套格式、繁中法術注解A／B沿用原有待決。整場／替代路徑、任務／競技場及完整保存parity仍開放。入口docs/re/20261003-spell-name-input.md。

初版反例fixture漏設既有教派，停在入會問句。補回正常前置狀態後乾淨重跑，舊碼分別因BS不被接收及N未修改typing失敗，沒有拿環境失敗當產品反例。修改前封存335份共用輸入，9313份前批附件保持。

正式鍵盤方向鍵過濾使用既有keyboardTokens文字模式，腳本驗證的是正常文字token與UI，沒有宣稱平台實體鍵盤真機驗收。法術字元集只接既有ASCII字母／空白／退格，不擴充繁中名字、輸入法或原版未知字元規則。

收據：[receipt.json](../verification/spell-name-input-20261003/receipt.json)。
唯讀入口：[verify.py](../verification/spell-name-input-20261003/verify.py)。
Docker掛專案`/src:ro`與本工作區`/out:ro`，執行`kol-package:20260907 python3 /src/docs/verification/spell-name-input-20261003/verify.py`，核對本輪附件、324Go及前批9313附件。
清理、440個入口異動與擁有權見工作區`audit.json`；無提交、推送或發布。

## RE-212追加勘誤

前文正常名稱Kelnalon→Kelnalo→KelnaloN是當時重製結果，未核對原版。原版有效輸入從空開始，RET／BS／N／RET結果是N；第12字自動確認，打包名稱顯示各單字字首大寫。前文英文「可見」說法過度：trace有改名、正式／觀察PNG相同，但英文drawSpellEdit當時未畫輸入文字。繁中名稱輸入可見的證據仍有效。RE-212補英文輸入框及20原版框零差，正常N與上限名稱存錄／成本雙語相符；來源與目前收據見[RE-212](20261003-spell-name-original.md)。原始收據與payload不改寫。
