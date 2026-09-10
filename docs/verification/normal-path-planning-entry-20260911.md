# 正常路徑修復與規劃入口之後的第一批狀態（2026-09-11）

## 結果

從書本選角一路走到**戰鬥規劃畫面**的正常路徑重新通了，`TestNormalBookEntry`
69.88 秒通過。這一場與 2026-09-08 的兩份收據**逐位元組相同**：

| 檔案 | 與 `battle-terrain-normal-fixed-20260908/` |
|---|---|
| `battle-state-records.json` | 相同 |
| `battle-alignment.json` | 相同 |
| `battle-terrain.json` | 相同 |

三次獨立執行到達同一個狀態。這一輪多拿到的是**規劃入口之後**的逐鍵狀態
（`battle-planning.json`），那正是 `battle-normal-seed-alignment` 一直缺的東西。

輸出在 `workplace/battle-planning-20260911/`。原版測試 seed=1，設定點在旅店招呼後，
DGROUP=`0x1EFD`。

## 路徑原本斷在哪

兩個獨立的原因，兩個都不是遊戲的問題：

**一、每一步跑固定 80,000,000 道指令當「這一鍵處理完了」。** 書本第四步那個
Space 要等城鎮載入完才會被讀走，而那要跑多久不是常數。量到的後果是已讀=5／
待讀=1——最後那一鍵一直留在佇列裡，店主招呼沒發生，畫面停在建築外觀。改成
**等 `KeysPending()` 歸零**（每輪 20M，上限 20 輪，超過就 fail）之後，四步的
已讀／待讀與 2026-09-08 那次完全一致。

**二、旅店序列被加了三步 `Down`／`Return`／`Space`。** 那三步是在
`KOL_BOOK_SELECT_DOWN=12` 的情境下加的，而 12 是 Maia、在 Trollsbane——
**另一座城**。在 Halfway House 這條路徑上它們是多餘的，加了之後招募點擊落空。
序列還原成 `recruit-click` 起的六步。

順帶更正一個參數：**Olear 是角色清單的第 2 個**（第 0 個是 Theodoric）。
`SELECT_DOWN=0` 選到的是 Theodoric，而他就是要招募的對象，於是隊伍只有一人、
走不出城——症狀看起來像「招募失敗」，實際上是選錯人。

## 規劃入口之後的逐鍵狀態

欄位依 SPEC-010：`+0x22` 規劃旗標、`+0x23` 攻擊式、`+0x25` 指令碼；
參戰記錄 DS:0x6712、步幅 50；目前槽位 DS:0x9E5A。

| 步 | 按鍵 | 目前槽 | +0x22 | +0x23 | +0x25 | IP |
|---:|---|---:|---:|---:|---:|---|
| — | （進場後，未送鍵） | 1 | 0 | 0 | `0x00` | `0AE9:1780` |
| 0 | Return | 0 | 0 | 0 | `0x00` | `0AE9:153D` |
| 1 | Right | 0 | 0 | 0 | `0x00` | `0AE9:17CE` |
| 2 | Return | 5 | **1** | 0 | `0x00` | `0AE9:13A7` |
| 3 | `.` | 5 | 1 | 0 | `0x00` | `0AE9:17DA` |
| 4 | Return | 0 | 0 | 0 | `0x00` | `0AE9:17D6` |
| 5 | Return | 1 | 0 | 0 | `0x00` | `0AE9:17BA` |

`plan-00.png` 是 Olear 的規劃畫面：左上肖像、左下身體圖、右側戰場、下方六個指令圖示。

## 不宣稱

- **這是逐鍵觀察，不是對拍。** 按鍵序列還沒與 remake 對齊，也沒有宣稱任何一步的
  結果與 remake 相同。
- 「目前槽」在六步裡是 1→0→0→5→5→0→1。DS:0x9E5A 讀到的值是不是「正在規劃的
  角色槽」還沒定案——跳到 5 的那兩步同時也是 `+0x22` 立起來的那兩步，兩者的關係
  未解，不從單一序列推語意。
- 沒有送出結算命令，戰後回寫與保存鏈不在這一輪。

## 重跑

```
KOL_CHECK=book KOL_BOOK_SELECT_DOWN=2 KOL_BOOK_BATTLE=1 \
KOL_BOOK_WALK=1 KOL_BOOK_NEXT_DAY=1 KOL_BOOK_ENTER_FIGHT=1 KOL_BOOK_PLAN=1 \
  tools/oracle/check_updated_dosgolem.sh
```

`.` 這一鍵需要 dosgolem 的主鍵區標點掃描碼（同日補上，值取自 DOSBox-X 的
`KEYBOARD_AddKey1`，scancode set 1）。
