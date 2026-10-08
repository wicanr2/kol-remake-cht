# RE-217：怪物射後計數與裝填別名

日期：2026-10-04。怪物v2與人物保存v3已核准；本輪沿用RE-216正常初戰。
工作區`workplace/monster-ranged-consume-20261004/`，入口為`prepare.py`、
`run_native.sh`、`ida_alias.py`、`run_counter.sh`、`review.py`。
原版資料、原版圖片與完整記憶體收據只留本機。

| 輸入／工具 | 定位 |
|---|---|
| KOLTOWN-unpacked.EXE | SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| 正式.i64唯讀來源 | SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8` |
| IDA Pro 9.4 | IDA linear；DGROUP基底0x2DED0；173條指令與MZ重定位後bytes相符 |
| dosgolem | 固定提交968ca774705853276ae55c8d1163c369481ac71f；正常旅店招呼後固定seed1 |
| 正常原版收據 | native1/native-shop-private.json；SHA-256 `3360f4af472531098aa3504d6d06c30b7ebcdea27ff50c49763d02d009f06766` |
| 有界射後收據 | counter-original.json；SHA-256 `f8e6df999cffd33cecd01e6366cc96e1f1d0584278758a772dd6147aac4104ec` |

## 已證實

- sub_1E3A8:0x207A5／0x20838保留原運算元`[bx-4478h]`，
  實際DS讀取為`BB88+actor*300+3D`。正常怪物槽6的BX為0x123D，
  讀DS:CDC5值33；槽7的BX為0x153D，讀DS:D0C5值0。
- DS:CDC5是怪物記錄0的+0x33，即第一個防具記錄內的byte3。
  原版0x1EC29複製所選裝備時將0寫33；正常觀察只有此一寫入。
  它沒有讀手持或遠程物品的種類byte。
- DS:D0C5對應未使用的怪物記錄6之+0x33。本次初戰從旅店到射後皆為0，
  監看未見寫入。此結果不證實其他旅程或後續戰鬥的殘值。
- 正常兩發射空後，槽6／7的Flag1D清0、Command改0x22，Kind由14／8減為13／7。
  全部195狀態、骰序與202 PNG和RE-216一致；新觀察只有只讀hook及寫入監看。
- 0x20884..0x208CE不依陣營或裝填種類分支。Kind減一，原值0寫FF並清Command。
  有界真指令測試12槽×五個裝填輸入×256首byte，共15,360組相符，RND seed保持。
  入口堆疊與返回補丁、欄位注入都明示，這份測試不代表正常玩家路徑。

## 未知與正式修正範圍

其他怪物槽的別名落在人物及有效怪物區域外。正常觀察槽11有2,312次寫入，
不能將缺乏靜態xref解讀為沒有寫入。完整別名來源與跨場殘值仍未知。
SPEC-022只將不依種類的減值及依Kind重算負重升READY；非耗盡Command與Flag1D
保持DRAFT。不改資料或存檔格式，不宣稱完整射後、正常耗盡或整場parity。

初次觀測的WriteHit.At型別錯誤保留於first-compile；首次有界測試因缺少
套件相對原版路徑而SKIP，保留counter-initial-skip.log。修正隔離路徑後重跑全數通過，
這兩次是驗證腳本問題，正式程式未因它們改動。

## 正式有限驗收

三個正式檔案修正怪物計數、依Kind重算負重與獨立回歸。
333份Go／335程式輸入，Linux SHA-256 `c1046c25ba7c98276f5b57bbb699dfa72737b34e9b9fe49c24018c76f9abe237`。
舊正式兩項負例失敗，修正後15,360原版對照與全部internal／cmd通過。
雙語正常各203鍵／204狀態，只有觀察欄位Kind出現預期13／7變更，42PNG保持。
正式與觀察軌跡相符，四回合雙語540狀態／28PNG保持。
收據、來源副本、唯讀重驗與清理見同工作區receipt.json／verify.py／audit.json。

同一335來源的Linux重建相同，Windows AMD64與macOS雙架構／universal及最低11.0結構通過。
完整雜湊在platform-checks.json；Windows／macOS本輪未執行，這是建置驗收，不是平台可玩聲明。
