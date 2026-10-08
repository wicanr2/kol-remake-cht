# RE-184：戰鬥返回後的世界時刻與次格

日期：2026-10-02。產品候選仍為 `7bc7d34a`，本輪只補來源證據。

## 已證實的正常返回

原版先完成戰鬥，再扣引發遭遇的步行耗時。戰鬥返回會將世界次格設為4,4。
耗時還受戰鬥覆寫的共用暫存影響，不能直接沿用進場前的移動方向。

| 等級 | 固定正常路徑的結果 |
|---|---|
| 已證實 | 原版戰鬥入口在49,23、次格2,3、時刻3360，移動率20；BB84為65535，代表有號值−1 |
| 已證實 | 61回合返回時，時刻仍3360；0x2A952及0x2A958將次格寫為4,4 |
| 已證實 | 本場搜刮後BB84為16，步行扣時函式使用DS:12B8加16處的byte30，耗時為2×(20+30)=100 |
| 已證實 | 0x2981F扣時刻後為3260；天氣擲骰位於戰鬥返回後、步行扣時刻前 |
| 已證實 | 新觀察的初態、351鍵、61回合邊界、1458戰中骰、訊息、戰後角色及出口計數與RE-182完整來源相同 |
| 未知 | 勝利、逃走及其他戰鬥出口會留下哪個BB84；不得把本例100耗時套用全部出口 |

## 固定來源與原始定位

| 項目 | 值 |
|---|---|
| 封裝 | Knights-of-legend.zip；SHA-256 `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46` |
| 程式 | KOLTOWN-unpacked.EXE；SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| IDA資料庫 | KOLTOWN-unpacked-ida94-derived.i64；SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8` |
| 工具與位址 | IDA Pro9.4 linear；DGROUP基底0x2DED0。執行期以原版DS定位，不將DGROUP位移當線性位址 |
| 執行器 | dosgolem revision `968ca774705853276ae55c8d1163c369481ac71f` |
| 時刻／次格 | DS:0070／0066／0068；IDA linear 0x2DF40／2DF36／2DF38 |
| 共用暫存 | DS:BB84；IDA原名word_39A54、linear 0x39A54。單一名稱不代表單一用途 |

| IDA linear | 原始bytes | 限定語意，全部已證實 |
|---|---|---|
| 0x297FB | 見私有IDA匯出 | 呼叫sub_29987，等待其中遭遇／戰鬥返回 |
| 0x29803 | 8b1e84bb | 扣時刻前重新讀取word_39A54 |
| 0x2980A | 8a87b812 | 依當下BX取DS:12B8表byte |
| 0x2981F | 29067000 | SUB將計算耗時扣自word_2DF40 |
| 0x2A952 | c70666000400 | 戰鬥返回後寫word_2DF36為4 |
| 0x2A958 | c70668000400 | 戰鬥返回後寫word_2DF38為4 |
| 0x21801／2181C／21821 | 見私有IDA匯出 | 清槽逐byte迴圈以word_39A54計數，完整16byte清除後停在16 |

以上是固定程式的讀寫契約，不為BB84建立永久方向名稱。120筆直接交叉參照包含戰場
初始化、指令與結算寫入；直接交叉參照不能排除間接寫入，也不能證明其他出口值。

## 最小追加實驗

- `workplace/battle-return-20261002/prepare_return.py`及`source_return.sh`沿用RE-183
  正常啟動／招募／出城，續送RE-182整場輸入。549個事件、162個輸入觀察；沒有
  direct-entry、角色／座標注入、常式或等待替換。四份初始化JSON保持原雜湊。
- 兩側沿用已登記的首次規劃seed1。本輪不重擲、不換種子，也沒有新重製GUI驗收。
  正常來源的完整seed設定與執行條件記於私有JSON；正式程式不鎖種子。
- `prepare_scratch_cases.py`及`scratch_cases.sh`對原版搜刮常式追加432組案例。
  執行前BB84分別設−1、0、1；seed固定1232439027，與既有重製LCG模型比較六份完整
  角色、骰數及最終seed。清過任何槽時結果16，未清時保持輸入，各組136保留、8變16。
  清除原本為空的槽也會留下16。這是常式邊界測試，不取代正常玩家路徑。

搜刮常式保留輸入值，不等於整場保留進場方向。戰鬥內其他寫入仍可能改變該值。
此差異會改變正常世界時間與返回位置，屬現行第6／44項閘門；不開硬體時序研究。

## 收據與下一閘門

[唯讀重驗](../verification/battle-return-20261002/verify_return.py)及同目錄receipt.json
保存來源結果、IDA定位、案例及歷史雜湊。分級定位另見
[位址索引](world-return-address-index.json)，由IDA匯出器自動合併，正式資料庫不改。
封存後另以[audit_return.py](../verification/battle-return-20261002/audit_return.py)核對
收據附件、擁有權、資料庫與批末控制面狀態；結果另存同目錄audit.json。
所有新原始payload、程式、遊戲資料及未驗PNG仍僅本機。

RE-183的1303份共享輸入封存於本輪parent-inputs，舊6130附件及RE-182結果依歷史
輸入重驗。新文件不替換舊收據中的現況檔案。產品程式未改，父規格維持READY；
SPEC-011返回子切片為DRAFT。下一步只確認正常勝利／逃走出口的最後暫存寫入，
再審查延後扣時刻、返回置中及世界／戰鬥亂數接續的實作契約。

在專案根目錄重驗已封存結果：

```sh
test -d "$PWD/docs/verification/battle-return-20261002" &&
timeout 90s docker run --rm --network none --memory 2g --cpus 2 --pids-limit 96 \
  --user "$(id -u):$(id -g)" -v "$PWD:/src:ro" \
  kol-package:20260907 python3 \
  /src/docs/verification/battle-return-20261002/verify_return.py
```

此命令重算來源與核對雜湊，不重播遊戲。需要重生原版時，先讀工作區
`source_return.sh`與`prepare_return.py`，沿其固定dosgolem revision、唯讀輸入及
明確可寫輸出建立新的工作區；不可覆寫本輪source1或其他歷史輸出。
