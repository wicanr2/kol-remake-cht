# RE-266：兩個戰鬥記錄空缺的來源核對

狀態：有限來源核對通過。+0x2C及+0x31的語意仍為未知，沒有正式實作。

本輪沿RE-265的未表示欄位，核對指令寬度與候選取址。正式Linux仍為
`1f31007c19ceac875a12df9b68d8629296764d01194983cf98d10cc8dcdbd274`，338份輸入未改。
工作入口為`workplace/actor-record-gaps-20261006/review.py`，本機輸出與來源雜湊見seal.json及audit.json。

## 已核實的範圍

| 項目 | 結果與證據等級 |
|---|---|
| 戰鬥記錄取址 | 已證實的既有來源：DS:0x6712＋slot×0x32；來源RE-208的0x20A5B至0x20A67，本輪重核bytes |
| 法術倒數+0x30 | 已證實：IDA 0x20EE1、0x20EF3及0x28730都只操作一個byte，不能據此將+0x31建模為倒數高位元組 |
| +0x2C候選取址 | sub_14CCF及sub_14D3A使用word_39A52與DS:BB88基址，並非已證實的50-byte戰鬥記錄欄位 |
| 其他位移候選 | 查詢338個已辨識函式，共42,940道指令，得到5個直接候選及18個立即數候選；候選不等於欄位讀寫 |
| 指令核對 | 275道唯一指令與EXE載入後bytes相同；套用2,281筆MZ重定位，載入segment為0x1000、檔案header為9,216 bytes |
| 原版樣本 | 既有seed2主動近戰及seed1休息首場，前九個槽共821個戰鬥狀態、7,389筆記錄；兩格觀察值均為零 |

兩場沿RE-264封存的dosgolem正常收據，本輪沒有重跑原版、改seed或注入狀態。
零值只能證明這些樣本，不證明保留欄位、固定初始化、跨場清除或不存在間接寫入。
未新增正常GUI、保存、HD、平台與封包驗收，完整50-byte及射後Command分歧仍保持開放。

## 工具、輸入與證據

- IDA Pro 9.4，固定image為`ida-pro-9.4-idapython:locked-v1`。
- EXE：`workplace/analysis/koltown/KOLTOWN-unpacked.EXE`。
  SHA-256為`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
- 資料庫：`workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64`。
  SHA-256為`5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`。
- 位址空間：IDA線性基址0x10000，DGROUP為0x2DED0；DS偏移與IDA線性位址分開記錄。
- 原始函式名、指令、運算元及bytes留在本機ida-gaps.json及ida-targets.json，沒有改名或修改正式資料庫。
- 法術來源為`workplace/spell-lifecycle-20261003/ida-lifecycle-private.json`。
  原版樣本路徑、雜湊及計數見本工作區review.json。
- 匯出合併兩份受版控位址索引，載入66及41筆。以已證實的0x25409註記作正對照。
  候選語意保持未知，不以自訂名稱證明語意。

## 重現與停止線

原版、正式資料庫與來源樹唯讀掛載，只有此工作區可寫，容器使用UID/GID1000。
先將資料庫複製到容器/tmp，再以`/opt/ida-pro-9.4/idat`執行ida-runner.py及ida-targets.py。
使用有界逾時、資源、network none與--rm；執行方式見run-ida.sh。
其後在既有Go image內執行review.py、seal.py及verify.py。verify.py只寫容器/tmp的審查輸出。

查詢腳本的ASCII讀取及語法失敗另存script-failures.json，修正後使用相同image與輸入重跑。
這些是工具腳本失敗，沒有遊戲或正式資料庫變更。
兩格沒有已證實的玩家消費端，不擴大至無關環境或執行期函式。
後續優先處理已有正常反例的射後Command來源；完整記錄限制仍如實保留。
