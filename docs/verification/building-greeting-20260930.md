# 進門後的發言者選擇、好惡與教派（#32，2026-09-30）

## 結果

多人隊伍進門後，remake 先問「Which one needs help?」，按 1…6 選發言者；被建築討厭的人會得到
訊息 14，被欣賞的得到訊息 15，全隊被討厭只印「Get out, scum!」就離開；SCROLL 建築拒絕
屬於別的教派的人。規則來自 `sub_11A93` 的靜態流程與三組原版關鍵點接手（RE-166），規格見
[SPEC-037](../spec/037-building-greeting.md)。

原版的問句畫面（門 21，隊伍 Olear 與 Theodoric）：訊息 16 三行加記錄 0x55，圖示列只有 DOOR，
兩人的框都是藍色。remake 原本接的是「What can I do for you?」，這條差異已消除（
[門 21 驗證](town7-door21-20260930.md)）。

## 原版關鍵點

探針 `internal/oracle/greet_test.go`，`KOL_CHECK=greet`：

```
mkdir -p workplace/greet-20260930
docker run --rm --network none --memory 3g --cpus 2 --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" -e KOL_CHECK=greet \
  -v "$PWD:/src:ro" -v "$HOME/cht/dosgolem-kolrand:/repo:ro" \
  -v "$PWD/workplace/greet-20260930:/out" kol-package:20260907 \
  bash /src/tools/oracle/check_updated_dosgolem.sh
```

| 輸入 | 版本或 SHA-256 |
|---|---|
| `KOLTOWN-unpacked.EXE` | `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| dosgolem | `~/cht/dosgolem-kolrand` `968ca77` |

| 關鍵點 | 原版位址 | 組數 | 比較 |
|---|---|---|---|
| 名稱換算 | `sub_14307` | 1,280 | 回傳值與 `town.CanonicalRecord` |
| 全隊被趕 | `sub_11A93:0x11D0B`…`0x11DF2／0x11E25` | 24,000（411 組被趕） | 走哪個出口與 `GreetRule.Expels` |
| 發言者好惡 | `sub_11A93:0x11E8E`…`0x11EF3／0x11F5D／0x12004／0x12017` | 10,896 | 被拒、有傷、被欣賞、一般四個出口與 `GreetRule` |

發言者好惡的測試要求四個出口各至少走到一次，全隊被趕的測試要求被趕的組數不為零。負對照：把
`CanonicalRecord` 的門檻 `race >= 2` 暫時改成 `race >= 3`，三支測試都在第一個矛盾處失敗。
日誌 `workplace/greet-20260930/greet.log`（只留本機）。

## 不宣稱

- `sub_14189` 的 5-bit 名稱解碼沒有在原版裡另外接手，關鍵點裡用回傳記錄號取代；有名稱的好惡欄
  （討厭 36 棟、欣賞 72 棟）全部能在記錄 0…38 找到同名的一筆，其餘寫 `None` 或空白（RE-166）。
- `sub_11722` 的數字鍵選人靠反組譯，沒有直接接手。
- 各功能做完之後回到「發言者判斷」的對應關係是強推論，remake 只在招募與住宿兩處重跑判斷。
- 滑鼠點隊伍肖像換發言者不含，另列工作項目。
- 訊息 14、15 的文字內容與各建築的好惡沒有逐棟對拍。
