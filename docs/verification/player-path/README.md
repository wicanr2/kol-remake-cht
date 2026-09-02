# 正常玩家路徑收據

這裡的截圖序列全部由 `cmd/kol -scene play -play-script …` 產生，經過 `town.Session`
這個狀態機——沒有 direct-entry、沒有座標注入、沒有測試旗標。輸入 token 對應原版鍵位：
`L R U D` 是方向鍵、`A S Z X` 是參考卡的斜走鍵、`<` `>` 移動圖示列白框、`RET` 執行。

```
tools/run_kol_shots.sh workplace/play/path.spec
python3 tools/verify_player_path.py --plan docs/verification/player-path/town.json
```

驗證器檢查三件事：

1. 每一步的截圖存在、解得開、尺寸正確；
2. **相鄰兩步不得完全相同**——每個 token 都要有可見效果。
   這一條抓到過一個真的 bug：門口圖示列的白框位置沒有畫出來，
   按 `>` 之後畫面完全沒變；
3. 指定的檢查點與**已通過對拍**的截圖逐 byte 相同。

## `town`：走動、站到門口、進建築

`L R > RET L R RET RET R R D D U`，14 張截圖：

| 步 | 操作 | 畫面 |
|---|---|---|
| 0 | （起始）| 戶外地圖，隊伍站在半途之家門口的格子上 |
| 1 | `L` | 往西一步，離開門的格子 |
| 2 | `R` | 走回門的格子 → **門口狀態**（地圖換成門面圖，圖示剩 DOOR／U-TURN）|
| 3 | `>` | 白框移到 U-TURN |
| 4 | `RET` | U-TURN → 回到戶外地圖 |
| 5–6 | `L R` | 再走回門口 |
| 7 | `RET` | DOOR → **進入建築**（肖像、標題框、八行敘述）|
| 8 | `RET` | 出來，回到戶外地圖 |
| 9–13 | `R R D D U` | 往東兩步、往南兩步、往北一步 |

檢查點：第 2 步與 `remake-parity/shots/town-door/inn-door.png` 逐 byte 相同、
第 7 步與 `remake-parity/shots/town-building/inn-entry.png` 逐 byte 相同。
**可玩路徑與對拍證據因此直接連上**——玩家走出來的畫面，就是通過逐像素對拍的那批像素。

## 這份收據不宣稱什麼

- 不宣稱遊戲可玩到結局。目前接上的只有城鎮這一段：走動、門口、進建築。
- 不宣稱鏡頭跟隨規則已證實。dead zone 是由 22 張原版畫格反解的**強推論**，
  規則本身沒有從程式碼讀出來。
- 走到地圖邊界目前是停住。原版會切到野外地圖，那條路徑還沒實作。
- 建築內按圖示之後的多輪對話還沒接。
