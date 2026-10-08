# RE-225：高攻擊式修正表缺列

更新日期：2026-10-04。狀態：89條指令及184表bytes審查完成，SPEC-009增補READY，高攻擊式切片有限CONFORMED；整場戰後保存仍待驗。

## 固定輸入與入口

| 來源 | 身分 |
|---|---|
| KOLTOWN-unpacked.EXE | SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| KOLTOWN-unpacked-ida94-derived.i64 | SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8` |
| IDA Pro | 9.4；線性位址，DGROUP基底0x2DED0，MZ映像起點0x10000，重定位字加0x1000 |
| dosgolem | 提交 `968ca774705853276ae55c8d1163c369481ac71f`；Go1.26.7 |
| 正常原版收據 | RE-224，SHA-256 `3354d35308286885ca4d4f39d3ed4fd2481e2ba83f494487f8887883be8ffb90` |

工作區`workplace/attack-modifier-20261004/`，入口prepare.py、ida.py、run_oracle.sh、review.py。
before-source／before-inputs.json固定334 Go／336輸入；原版及舊失敗保持[RE-224](20261004-missile-defensive-posture.md)封存。
完整原版資料、指令、人物、候選與畫格僅本機，不發布。

## 正常玩家反例

同一正常防守／休息377等價鍵、進場seed1、無規劃重設。前七個執行入口相符。
第七回合怪物8近戰，attack-entry seed3520455839，Command0x21、Style8、目標1的Guard2。
原版已開始下一發投石時，重製仍擲近戰傷害；第361項骰原版n100、重製n9。
第八至十入口13項差異，終seed相同不足以證明通過。

## 原始定位與READY契約

| 原始定位 | 已定位內容 |
|---|---|
| `sub_26D76:0x26FDD..0x27002` | Style1以rand(3)-1選列，其餘Style+1 |
| `0x27005..0x27035` | 從DS:0x110A加14×列與2×目標Guard，讀有號word，加到攻擊分數 |
| DS:0x110A..0x11B1 | 12列×7欄，共168bytes；後接招架訊息表0x11B2 |
| `0x27038..0x27076` | 非負且Style>1，武器缺指定位元時分數除2 |

現行optionModifier只有九列，Style8／9／10的row9／10／11被越界檢查略過。
原版三列為`42,-38,-8,-35,-33,-20,-18`、`55,-25,5,-27,15,-2,-10`、
`59,-21,9,-18,24,7,-41`。固定EXE bytes與完整局部常式已通過審查，授權正式補回表值。

局部探針明示遭遇等級1、目標1、參數0及seed3520455839，只替換文字呈現，
得到攻擊64、防禦71、結束seed1729179665。等級來自重製見證的測試輸入，
原版正常收據沒有直接觀察該全域，不能稱為全部輸入狀態一致。
首次漏設此全域的局部結果59／71留於modifier-initial-missing-level.json，不外推為正常攻擊值。
不改現有列、骰序、武器位元、徒手、投射、資料格式、UI或人物保存v3。
驗收使用同一原版序列，不能換seed或縮短失敗路徑。


## 正式修正與正常驗收

READY後只改attack.go與attack_test.go；補三列、維持前九列與既有公式。正式334 Go／336輸入，
怪物v2與人物保存v3保持。舊正式反例為攻擊72，新正式為64；43組高攻擊式／武器位元／負分數
與第七回合避開傷害的測試、全部internal／cmd通過。

同一原版正常防守377鍵，雙語各378狀態／195觀察PNG／20正式PNG，正式與觀察相符。
十個執行入口每語言1900既定人物欄、100疲勞／體力計數器零差。
4564進場骰與534戰鬥骰的n／seed_before逐項相同，最後seed2907222844。
第十回合怪物7首次Kind由1降0，兩側正常戰敗；最後怪物Kind5／0相符。
原版與重製戰敗結果圖已目視核對，繁中「懸崖巨魔們擊敗了你們」可讀。
不外推正常畫面全部幾何或完整50-byte終態。

正常耗盡已觀察，耗盡後沒有下一規劃。原版after-consume的Command34與重製35既有差異保持，
不能稱完整射後狀態一致。原版0x211AD止於戰敗結果，尚未完成戰後回城、保存與重載同狀態鏈。
因此本高攻擊式表與RE-224姿態切片有限CONFORMED，近戰／投射父規格與整體#6／#44仍開放。
RE-224的舊失敗收據不改寫；本次新來源重播才消除其第七回合13項差異。

送禮八條合計300狀態／300PNG、完整v3保存重載、SLEDGE雙語各270狀態／14PNG保持。
正常休息路徑雙語各337狀態，154觀察／18正式PNG及trace保持。
三平台同來源候選重建，Linux乾淨重建逐byte相同；Windows與macOS只驗結構，未作實機驗收。
平台雜湊見platform-checks.json，候選僅本機中間產物，不替代dist-all已發布版本。

## 重現與下一閘門

工作區追加入口prepare_outputs.py／build.sh、run_normal.py／compare.py、
run_regression.sh／run_regression.py、run_rest_regression.sh／rest_regression.py、
cross_build.sh／mac_build.sh／check_platform.py、finalize.py、seal.py／verify.py。
before-source／formal-source保存前後336來源，receipt.json索引附件，audit.json保存清理狀態。
唯讀重驗用kol-package:20260907，專案唯讀掛/src，執行本工作區verify.py；
無網路、UID/GID1000、2GiB／2CPU／128程序、外層120秒。

下一步保留本正常戰鬥至戰敗結果的原版前綴，再延伸正常戰後回城、住宿保存與重載，
核對角色、裝備、傷勢及保存的實際輸出。不因十入口通過關閉整場戰鬥保存閘門，
不為未接到玩家體驗的完整別名或內部記憶體開放無界研究。
本輪未提交、推送、Issue寫入或發布。
