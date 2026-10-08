# RE-185：正常逃走來源與返回次格修正

日期：2026-10-02。正式候選為 `9c64e343`，只有返回次格切片通過有限驗收。

## 原版退出結果

原版隨機戰鬥返回後，次格都設為4,4。耗時依退出時的BB84計算，各出口不能固定
使用同一數值。這輪使用正常啟動、旅店招募、出城、遭遇及原生等待，不注入角色、
座標或指令，也沒有替換原版常式。

| 證據等級 | 正常路徑 | 返回BB84 | 表byte | 耗時 | 返回世界狀態 |
|---|---|---|---|---|---|
| 已證實 | RE-184正常61回合自然敗北 | 16 | 30 | 100 | 49,23；次格4,4；時刻3260 |
| 已證實 | 本輪SPRINT在第五回合累積疲勞，全隊倒下 | 16 | 30 | 100 | 49,23；次格4,4；時刻3260 |
| 已證實 | 第一回合兩人選REST及YES，再選回合確認列的逃走格並答Y | 0 | 10 | 60 | 49,23；次格4,4；時刻3300 |
| 未知 | 正常勝利的最終BB84及完整世界／戰鬥亂數接續 | 不推定 | 不推定 | 不推定 | 仍由父閘門追蹤 |

兩條新來源在第一個正常規劃時點的完整初態相同，seed為1。與RE-184比較，角色、
計數及seed相同；本輪跳過舊版的先行規劃展示，指令欄及累積步數不同。收據保留原始
差異，不把不同觀察時點宣稱為全欄位相同。逃走10鍵在記錄265後答Y；兩名隊員
頭部仍15／16，出口兩側倒地計數都為255，BB84至函式返回保持0。

SPRINT來源有193鍵、第五回合退出，兩人頭部都為255且沒有Y輸入。它是自然敗北，
不能因測試程序通過就登記為逃走。前兩版重播漏選MOVE及逐人YES，失敗診斷保留。

## 固定身分與位址

| 項目 | 固定值 |
|---|---|
| 原始封裝 | Knights-of-legend.zip；SHA-256 `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46` |
| 分析程式 | KOLTOWN-unpacked.EXE；SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| IDA資料庫 | KOLTOWN-unpacked-ida94-derived.i64；SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8` |
| 工具 | IDA Pro9.4，IDA linear；DGROUP基底0x2DED0。執行期狀態採DS相對位移 |
| 執行器 | dosgolem revision `968ca774705853276ae55c8d1163c369481ac71f` |
| 建置及GUI | Go1.26.7／Ebitengine2.9.9，既有image `eob-remake-go:1.26.7-ebiten2.9.9-audio`；image ID `3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa` |
| 收據與圖像 | 既有image `kol-package:20260907`及`tools/png_io.py`標準函式庫，程式雜湊隨收據鎖定 |
| 控制暫存 | 原名word_39A54，IDA linear 0x39A54，DS:BB84；用途會重用 |
| 返回次格 | 原始0x2A952 bytes `c70666000400`，0x2A958 bytes `c70668000400`；已證實無條件寫4,4 |
| 耗時讀取 | 原始0x29803 bytes `8b1e84bb`，0x2980A bytes `8a87b812`；0x2981F bytes `29067000`扣時 |

IDA暫存副本保持338函式，匯出672、128、58條有限指令及120筆直接暫存交叉參照。
0x21765仍在原始sub_1E3A8函式內，不另造函式名稱。直接交叉參照不排除間接寫入。
原有九筆分級定位與[索引](world-return-address-index.json)保持原雜湊，正式資料庫不改。

正常來源入口為`workplace/battle-exits-20261002/prepare_flee3.py`／`source_flee3.sh`
及`prepare_flee4.py`／`source_flee4.sh`。它們沿用固定dosgolem前綴與原版種子設定：
旅店招呼後及第一次正常規劃入口，DS:1600=1、DS:1602=0。沒有換種子或挑選結果。
全部原始資料、遊戲副本、程式與新PNG仍僅本機，不增加公開素材授權。

## 次格的有限正式修正

[SPEC-011 RE-185](../spec/011-wilderness-ambush.md)先完成READY審查，再於
startAmbush標記實際隨機戰鬥狀態。正式共同出口在搜刮及角色回寫後，把該場世界
次格設4,4並清除標記。任務、競技場及過期狀態不觸發置中；世界大格、時刻、
天候、亂數與保存格式保持既有契約。

| 驗收 | 結果 |
|---|---|
| 正式共同出口 | 三種結算乘三種來源共9組；一次性出口及重按鍵通過 |
| 規則與建置 | 全部internal／cmd通過；正式候選SHA-256 `9c64e34362cee13c5abbea95174f46f7c78f90d1f33968947c9ded5704c4d05f` |
| 正常雙語玩家路徑 | 各492鍵、493狀態與畫格；61回合後返回4,4，再正常移動四步 |
| 保持原結果 | 返回前各480狀態／畫格完全相同；61回合、1458戰中骰序及兩人完整34槽不變 |
| 返回畫格差分 | 九張英文各只改舊／新標記的8像素；繁中各72像素，沒有其他差異 |
| 正式保存相容性 | 新正式程式重載RE-182正常住宿存檔，再正常進門招募；雙語各16狀態／畫格完全相同 |

觀察副本SHA-256 `cd1f16f00f2d15b118ff21b5ca34aeb72dfc5e51fc71b79f4c014bbedda7fa1e`，
只在與原版相同的首次規劃入口固定測試seed1並記錄資料。正式程式沒有永久播種旗標。
三種結算的共同出口測試不冒充三條完整正常GUI；正常61回合仍是自然敗北。
本輪重載既有住宿存檔，沒有從新置中位置另走住宿保存鏈。

次格返回獨立切片為有限CONFORMED。產品時刻仍3320，原版本場返回3260；
扣時刻／亂數部分維持DRAFT，父SPEC-011／027及第6／44項維持開放。
下一閘門只補正常勝利的最終暫存，再審查延後扣時刻與必要的亂數接續。

## 收據與重驗

[verify_exits.py](../verification/battle-exits-20261002/verify_exits.py)及同目錄receipt.json
核對原版來源、正式規則、正常畫格、舊產品反例及重載。1309份RE-184共享輸入先封存，
`history_check.py`只映射已鎖來源，再用原歷史驗證器重算1516附件及下層結果。
歷史收據、驗證器與workplace輸出不覆寫。批末稽核另見
[audit_exits.py](../verification/battle-exits-20261002/audit_exits.py)及同目錄audit.json。

在專案根目錄重驗已封存結果：

```sh
test -d "$PWD/docs/verification/battle-exits-20261002" &&
timeout 90s docker run --rm --network none --memory 2g --cpus 2 --pids-limit 96 \
  --user "$(id -u):$(id -g)" -v "$PWD:/src:ro" \
  kol-package:20260907 python3 \
  /src/docs/verification/battle-exits-20261002/verify_exits.py
```

此命令核對雜湊並重算收據，不重播遊戲。重生時先讀上述source_flee入口、
`build_center.sh`、`build_probe.sh`、`normal_gui.sh`及`reload_gui.sh`，另設明確可寫輸出；
不可覆寫本輪輸出。全部Docker預設無網路且有限時、有限資源、UID/GID1000。
