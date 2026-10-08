# RE-267：首名怪物的裝填來源

狀態：首名怪物來源切片正式有限CONFORMED。其他怪物別名、完整50-byte及整場驗收保持開放。

來源入口為`workplace/monster-first-reload-20261006/review.py`。本輪沒有新原版或IDA執行，
重核173道載入指令、128套現行v2資料，以及RE-227三場正常戰鬥的八筆裝備複製。

## 已證實的來源

| 項目 | 定位與範圍 |
|---|---|
| 射前／射後讀取 | IDA 0x207A5／0x20838；DS:BB88＋actor×300＋3D，完整byte與6、10、11比較 |
| 首名怪物 | actor6，位址DS:CDC5，等於第一筆128-byte怪物記錄＋33 |
| 寫入 | IDA 0x1EC29，逐byte複製已選四套裝備；＋33是頭部物品原始16-byte記錄＋3 |
| 空槽 | 使用v2的records保留原始填充，不能改讀可能為null的items，也不能以手持Class替代 |
| 原版讀取 | RE-264正常REST場，13個射後消費點皆33，最後寫入為0x1EC29 |
| 資料母體 | 32種怪物、128套，來源值有12種，均非6／10／11；此統計不取代完整byte公式 |
| 跨場複製 | RE-227三場的首名來源依序為33、0、65；八筆完整header＋裝備複製及現行v2來源相符 |

以上限定首名怪物。第二名讀第六筆裝備記錄，可能帶前場殘值；不能由本場未使用推定固定零。
第三名讀取已超出12筆怪物記錄區，第六名怪物的DS:DCC5落在MZ堆疊範圍D5B0至DD80。
堆疊與後續別名不按武器種類猜補，不擴大到無關helper研究。

## 原型驗收

- 單一execution.go研究覆蓋檔，只重生首名來源並接射前／射後分支，其他別名保持目前行為。
- 戰鬥套件通過；新增空槽填充、三個免裝填值、其他值、0到FF邊界及分段命中收尾檢查。
- 原型與唯讀觀察雙語四個正常程序共1,544狀態。每語言213個seed及1,917筆角色記錄，50個語意欄位相符範圍沿RE-265。
- 每語言Command差異由42降為16，首名差異歸零，剩餘都在第二名。其他比較欄位相符。
- 移除觀察欄位後，完整trace與原型相同；原型完整trace與前版正式相同，12組PNG相同。
- 雙語正常回城保存共1,136狀態，新程序重載26狀態。完整v3只排saved_at後與現行正式來源候選相同，16PNG保持。

原型SHA-256：`dba813be51cc7fd59b0c081b50f9993db32ae5bc2a0a984f54dcf7e2ef9a8170`。
唯讀觀察SHA-256：`23f7679c121f1fe51fc6e48ea93c12d14d9904544683fddda686290e6272f554`。
這些結果只支援已證實的首名切片，不能關閉其餘別名或完整戰鬥閘門。

## 證據與重現

EXE為`workplace/analysis/koltown/KOLTOWN-unpacked.EXE`，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
IDA Pro9.4資料庫SHA-256為`5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`。
位址空間為IDA linear基址0x10000、DGROUP0x2DED0；原始byte及非破壞性註記沿RE-217、RE-264。
核對MZ的9,216-byte header、2,281筆重定位及載入segment1000。
原版dosgolem固定`968ca774705853276ae55c8d1163c369481ac71f`，seed與控制器沿封存父收據。

在既有`eob-remake-go:1.26.7-ebiten2.9.9-audio`容器內依序執行prepare.py、build.sh、
run.sh、review.py、run-save.sh。來源唯讀，僅工作區可寫，UID/GID1000、無網路、有界資源與逾時。
Xvfb等待socket並由trap收尾。原始收據、覆蓋檔、候選及新PNG都只留本機。
正式接線前另保存READY審查，正式驗收入口見後續更新。

## 正式接線與驗收

DRAFT原型與保存鏈通過後，READY審查收據為ready-review.json，規格當時快照為spec-ready.md。
只更新internal/battle/execution.go及first_monster_reload_test.go。正式337Go／339輸入，
Linux與原型逐byte相同，完整SHA沿上文原型值。十九套件與五個無測試套件通過；
正式雙語共1934正常、保存及新程序重載狀態，28PNG與原型相同，完整v3只排saved_at。
全輪共4640狀態、68PNG。父規格及其他怪物仍未全驗；長任務與跨平台需在新來源回歸，
沒有封包、版號、發行或公開變更。正式入口為build-formal.sh及run-formal.py，
封存後以verify.py唯讀重驗，輸入／附件雜湊見seal.json及audit.json。
