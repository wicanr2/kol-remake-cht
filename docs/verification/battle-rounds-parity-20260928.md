# 正常戰鬥六個回合的執行段同狀態比較（2026-09-28）

## 結果

同一場正常遭遇（Olear、Theodoric 對一隻 SLEDGE）的六個回合，remake 從原版每一回合
執行段入口的狀態與亂數接手，跑完執行段之後與原版逐項一致：

| 回合 | 內容 | 擲骰 | 比對項目 |
|---:|---|---:|---|
| 1…4 | SLEDGE 走近、Theodoric 對空揮拳 | 各 4 次 | 全部一致 |
| 5 | SLEDGE 打中 Olear，累積傷害 0 → 10 | 9 次 | 全部一致 |
| 6 | 再打中一次，10 → 12 | 9 次 | 全部一致 |

比的是：每一次 `rand(n)` 的 n 與順序、回合結束時的 32 位元亂數狀態、三名參戰者的位置、
累積傷害、六個部位、挨打累計 `+0x2D`、招架次數 `+0x0F`、方向 `+0x28` 與疲勞。

```
tools/run_original_parity_tests.sh
```

`TestOriginalNormalRoundsExecution` 與既有五支（初態、地形、怪物裝備、規劃、規劃逐格）
一起跑。

## 原版那一側

`tools/oracle/book_startup_test.go`，`KOL_CHECK=book` 加上
`KOL_BOOK_ENTER_FIGHT=1 KOL_BOOK_PLAN=1 KOL_BOOK_EXECUTE=1`，
`KOL_BOOK_PLAN_KEYS=.,Return,Return,.,.,Return,.,.,.,Return`，`KOL_BOOK_TIMEOUT=1200s`。
規劃入口的 seed 固定 1（同 planning-parity-20260911）；之後每一名都把白框移回第 0 格按
Return（YES）確認，Theodoric 第一回合選徒手、之後重複，Olear 不選指令。

| 輸入 | SHA-256 |
|---|---|
| `KOLTOWN-unpacked.EXE` | `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| dosgolem | `~/cht/dosgolem-kolrand` `968ca77` |
| 收據 `workplace/battle-first-round-20260928/battle-rounds.json`（只留本機） | `84e93b90bc7cc6a51c6ba6ac4deb620448aeceb19e25c5fecc6b78992bf8c507` |

## 抓到並修正的三處

1. 近戰對手防禦選項為 0 時，remake 多擲一次防禦骰。
2. 擊退：remake 每次命中都判；原版只在傷害 16 以上時判，而且判的時候一定擲骰。
3. 疲勞、挨打累計與招架次數：remake 在各人出手後結算；原版在回合收尾依索引逐名結算。

細節與位址見 RE-158。修正前第 5、6 回合的擲骰序列、傷害與位置都不同，修正後一致；
這也是這支測試的反向對照。

## 不宣稱

- 規劃段不在這裡比：每一回合都從原版的決策與站位接手。原版規劃的「按 YES 才換人」與
  remake 不同，另列 `planning-yes-confirm` 待使用者決定。
- 只有近戰命中、未招架、未擊退、未倒下的情形；射擊、招架成功、擊退成立、力竭與倒下
  沒有出現在這一場。
- 訊息記錄有收在收據裡，但這一支沒有逐句比。
