# RE-242：執行段的戰場鏡頭

接續 [RE-241](20261005-attack-message-phases.md)。正式 Linux 為
`9ab6289798c3ee46351c25f335e536ee7b55396c01257879d9be90e42100afcd`，338 輸入已核對封存來源。
目前戰場顯示差異 2,773 至 3,287 像素，與傷害等待時點分開追查。

## 現行事實與驗收

- `State.Camera` 使用 `Cursor`。英文與繁中共用這個出口，也由同一槽取得騎乘視線旗標。
- 正常執行第一回合 `Cursor` 保留 0，第二回合保留 4；原版四次投石等待的 `word_37D2A`
  分別為 6、7、7、6。正常休息與移動等待也記錄各自出手者。
- 先查原版鏡頭取值、出手者切換與繪圖控制流；原版收據的 `camera` 是世界鏡頭，
  不能冒充戰場中心。
- 隔離原型只沿已證實的出手者欄位補切換，不改規則、資料格式或亂數。
- 同一固定 seed 1、249 正常鍵驗雙語；逐畫面核出手者與鏡頭，再比較八張投石戰場 PNG。
  近戰、法術、擊退、倒地圖示或掉落繪圖若另有差異，保留為各自閘門。

工作區 `workplace/execution-camera-20261005/`，入口 `ida-camera.py`。
目前狀態：鏡頭與此正常路徑的戰場呈現切片有限 CONFORMED，完整戰鬥仍未完成。

## 原始控制流與勘誤

IDA Pro 9.4，IDA linear；輸入 KOLTOWN-unpacked.EXE SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`，
正式 .i64 SHA-256
`5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`。
ida-camera.json 共 623 指令，補充查詢 ida-camera-callers.json 保留原名、位址、bytes 與 caller。

**已證實**：0x1FF6F 由排序記錄 +0x21 寫 word_37D2A；0x20036／0x2003A
將該槽傳給 sub_223C0。renderer 的 0x223EB／0x22401 讀該槽 +0x1F／+0x20，
寫 word_2FB4C／word_2FB44，作九乘九視窗中心。第二輪額外移動 0x20944
也換出手者。力竭與法術消失 renderer caller 分別為 0x20E4B／0x20F49。
正常 Step 在交出訊息結果時跟到 Actor，讓共用 Camera 與騎乘視線沿原版出手者取值。

**已證實**：同格候選條件在 0x22603…0x22637；0x2263D 呼叫 sub_22866，
其 0x22882／0x2289D 判原始頭／胸 byte 0xFF。倒地隊員 encoded 9、怪物 8，
站立隊員槽號加一、怪物 7。沒圖示時才由物品座標匹配填 encoded 39。
typed Down 已把原始倒地正規化為零，本輪不改規則；runtime sprite index 為 encoded 減一。
SpriteAt 舊版只返回站立圖示，也沒掃 Ground。原型補倒地、地上物品與中心槽／同格優先度。

RE-241 將 0x2746A 的 sub_1C1BF 誤寫為「重畫」。本輪直接匯出函式後確認，
它是卷軸文字的換行與輸出；戰場繪圖為 sub_223C0。此勘誤不改 0x27473 等鍵先於
0x274E9 傷害骰的證據與正式結算修正。RE-241 封存快照保持，在現行研究檔追加勘誤。

## 呈現時點與正式結果

第一版原型只跟出手者並補倒地／物品／同格優先度，規則回歸通過，正常雙語四次
共1,000狀態、324PNG。78出手者與Cursor相符，75英文戰場顯示區零差，
frame219／241／244仍分別有2970／3225／3046像素差異。整批在prototype-v1封存。

**已證實**：移動函式sub_26384在0x265CD先等第一句，0x265D9才重新畫戰場；
第一句保留移動前畫面，遊戲座標已更新。sub_277F4的0x279FB／0x279FF
將word_3237E被擊中者傳給renderer，倒地重畫中心與word_37D2A出手者分開。
第二版用只供呈現的ViewActor與移動前Actor快照，既有第二句推進時RefreshView；
正常傷害倒地改以被擊中者為中心。規劃重排清快照；不倒回Actor、不改RNG、不加保存欄位。
英文與繁中共用中心、騎乘視線與同格優先度。

原型第二版四次重播1,000狀態、324PNG，78英文完整戰場全部零像素差異。
先DRAFT、審查READY及ready-spec.md快照，再改七個既有檔案，336 Go／338輸入保持：
state.go、execution.go、round.go、attack_test.go、main.go、ui/battle.go與ui/battle_cjk.go。
舊正式兩個鏡頭／圖示規則反例失敗，V1三張呈現反例保留；新規則與兩段呈現測試通過。

正式Linux SHA-256：
`2ba7018a517378487bcabc342deebe0f744e2b4befcfb387d00a4d02e3143ed3`，
與第二版原型逐byte相同；全部cmd／internal通過。正式雙語再重播500狀態、162PNG，
與唯讀觀察版的完整trace與PNG相同，78英文戰場保持零差。
兩版原型及正式合計2,500正常狀態、810PNG；只有V2與正式的1,500狀態屬已修正呈現。

| 比較 | 結果 |
|---|---|
| 原版78出手者與Cursor | 相符，倒地畫面中心另取被擊中者 |
| 英文78張144×144戰場 | 逐像素全部相同，移動前／後與Olear倒地相符 |
| 繁中戰場 | 白框外全相同；既有細白框內310至328像素不同 |
| 完整英文頁面 | 全部仍有174至778像素不同，未宣稱整頁完成 |
| 78畫面seed與已解值 | seed差異零，保留12欄射後Command34／35 |
| 七AI入口／輸出、兩執行及後續規劃 | 保持；首次初始化前hook仍不可比 |
| 規則／輸入／保存 | 除Cursor外完整正常trace保持，同一249正常鍵，無新骰與保存格式 |
| 正常旅店保存、退出、新程序重載 | 雙語238狀態、十圖、完整v3只差saved_at |

最初額外驗證將繁中假設為英文完整三倍最近鄰放大而失敗。真正差異全部在既有白框
的內側邊界；CJK renderer原本採一實體像素細框，這輪沒有改線寬。
初始planning第172張繁中完整PNG與RE-241正式逐byte相同，證明既有樣式保持。
初次結果保留於candidate-nearest-neighbor-assumption.json；只修比較方法，四次已成功
遊戲重播不重跑。逐張保留白框內差異及全頁差異，沒有宣稱繁中全像素或全UI通過。

正常旅店短鏈沒有戰鬥，不稱新版本戰後保存或完整任務長鏈。射後別名、其他AI／法術／
擊退玩家路徑、完整英文UI與父規格保持開放。原版EXE／DB／PNG及衍生證據僅本機，沒有公開散布權變更。

入口prepare.py、build.sh、upgrade.py／build-v2.sh、run.sh candidate／formal、review.py、
promote.py、build-formal.sh、run-save.sh。before／formal-source、DRAFT／READY、V1來源與輸出分開。
附件及外部來源雜湊、PNG CRC／解碼、UID/GID1000與容器清理見receipt.json／audit.json。
