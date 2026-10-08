# RE-190：正常招募、施法與世界返回

日期：2026-10-02。正式程式保持RE-187的`dea837bc`，返回成本仍為DRAFT。
重驗入口為[verify_spell_return.py](../verification/spell-return-20261002/verify_spell_return.py)。

## 固定來源

| 項目 | 來源與身分 |
|---|---|
| 原版程式 | KOLTOWN-unpacked.EXE，SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| IDA資料庫 | KOLTOWN-unpacked-ida94-derived.i64，SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`，本輪未修改 |
| 靜態定位 | IDA Pro 9.4，IDA linear；沿用RE-188必要指令匯出，不新增推測性命名 |
| 執行器 | dosgolem revision `968ca774705853276ae55c8d1163c369481ac71f` |
| 執行期位址 | DGROUP相對位移；BB84是原始word_39A54的位移，不與IDA linear混用 |
| 隔離工具鏈 | `eob-remake-go:1.26.7-ebiten2.9.9-audio`，沿用鎖版Go與dosgolem來源 |
| 正常法術 | 原始Warder記錄`+0x200`的`bb0654b9599935207400000005008a53`，Vartwelyrmi，沒有注入或改寫 |

## 正常玩家來源

從KOL.EXE正常啟動，經標題、書本選Olear，再在Halfway House的COMPANION
名冊選Warder。先完成硬碟確認，再於名冊下移七次。離城與世界步行沿既有來源，
第一次伏擊以正常FIGHT進入規劃。Warder先丟掉原有ElfBow，手空後選第一筆法術、
Positive及合法目標游標鍵。Olear沿既有步行、攻擊與休息策略。

seed1在旅店招呼後及首次規劃前固定，設定方式為DS:1600=1、1602=0。
三版保持此設定及戰鬥策略，沒有換種子選結果。沒有direct-entry、角色／法術／
座標／指令注入、等待替換或亂數結果攔截。正常來源保存於
`workplace/spell-return-20261002/source3/`。

| 已證實項目 | 實際結果 |
|---|---|
| 正常招募 | 隊員Olear、Warder；初態完整原始角色與法術記錄保留 |
| 原有物品 | Warder手上ElfBow為`72b0c013ee000000225828ff190be200`；正常DROP後施法時手空 |
| 施法 | 第二、三回合各一次，訊息記錄`0x169`，表示目標格沒有人 |
| 成本覆寫 | IDA linear `0x20D1B`觀察兩次BB84=720；法術記錄保持原始16bytes |
| 戰鬥終態 | 42回合、921個規劃／執行按鍵；隊伍兩人倒地，敵人未倒地，正常敗北 |
| 原生等待／亂數 | 出口後八次原生等待；戰中988次亂數，總計1021次，包含32搜刮骰及一次天候骰 |
| 最後寫入 | 清槽觀察`0x21821`最後留下BB84=16，先前720沒有保留到世界扣時刻 |
| 世界返回 | 49,23、次格4,4；`[bx+12B8h]`讀byte30，移動率20，成本100；3360降3260 |
| 天候種子 | `0x29A83`前為397419821，`0x29800`觀察為4068287292，扣時刻在天候後 |

475個成本觀察及561筆變值寫入保留原始定位。變值寫入記錄的IP是指令執行後位置，
例如`0x20D1B`對應原始寫入`0x20D17`；不能把它當作另一條寫入指令。
成本觀察的actor欄仍是DS:9E5A，不外推成所有常式的arg_0目標。

## 與正式規則比較

以兩次原版狀態的原始法術bytes建立Spell，直接呼叫既有Go `CastTime`，兩例均為720。
測試只加在Docker暫存副本；314個cmd／internal／tools輸入及正式程式雜湊保持。
護符第一格74的類別成本80，後面00的類別成本100較高，因此取00，沒有低位額外成本。
效果5成本200，強度8A骰式成本420，合計720。不能只看第一個護符便推定764。

這只驗施法成本計算與正常原版資料流。兩次施法的結果都是NothingThere，沒有證實
對目標生效、治療、完整重製GUI或正常施法後住宿保存。RE-189的840亂數位址示例仍
屬直接建立輸入的指令實驗，本輪沒有將840或720帶到世界讀取。

## 失敗分類與收斂

source1把新增Down送給TypeKeys，變成文字；source2已改具名方向鍵，但仍停在硬碟
Yes／No確認畫面。兩版都招募Theodoric後於姓名檢查停止，尚未進戰鬥。
原始來源、run.log與畫格保留；source3只訂正輸入層及原始C字串檢查。
一次source2準備腳本的字串錨斷言在產生檔案前失敗，未啟動原版。
這些是驗證驅動器問題，沒有記為產品缺陷。

正常施法後敗北返回的來源缺口已補。返回成本仍須在正式規則中依原版寫入順序
追蹤，再按已驗的勝利、敗北與逃走來源審查有限契約。表外索引的實際讀值、正常
施法後勝利／逃走及完整世界進場亂數仍未知，不能以靜態表或按結算固定成本補猜。
父SPEC-011／027及第6／44項保持開放，沒有新資料／保存格式或發行聲明。

## 收據與重驗

工作區入口是`preserve_parent.py`、`prepare_spell3.py`、`source_spell3.sh`與`check_cast.sh`。
1336份共享輸入先封存於parent-inputs；RE-189收據及下層歷史保持原身分。
原版副本、完整payload、正常畫格與失敗樣本僅本機，不在本輪提交或發布。

在既有Docker工具鏈把專案唯讀掛到`/src`，執行
`python3 /src/docs/verification/spell-return-20261002/verify_spell_return.py`重驗。
首次封存以`--seal`及唯一輸出`/out/receipt.json`；`--audit`另檢查擁有權、清理與Git狀態。
正式入口為[receipt.json](../verification/spell-return-20261002/receipt.json)及
[audit.json](../verification/spell-return-20261002/audit.json)。
