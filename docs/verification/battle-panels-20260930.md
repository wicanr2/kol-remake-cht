# 戰鬥左側兩框的同狀態比較（A10，2026-09-30）

## 結果

原版正常玩家路徑（書本選 Olear → 旅店招募 Theodoric → 次日野外遇 SLEDGE）打到第 5 回合
Olear 胸與右腿掉血，第 6 回合規劃段停下。在這一刻對 Olear、Theodoric 與敵人各呼叫原版的
左下框、傷勢遮罩與左上框裝備模式，記下交給繪製層的參數；remake 從同一刻的參戰者記錄建狀態，
用正式的 `BodyPanel`／`EquipmentPanel` 算同樣的東西，全部一致：

| 參戰者 | 部位（目前／上限） | 傷勢遮罩 | 裝備列 |
|---|---|---|---|
| Olear | 15/15 13/15 20/20 20/20 20/30 30/30 | 247 點，讀取座標與順序相同 | Elf Bow、Chain Coif、Chain Shirt、Chain Pants |
| Theodoric | 全滿 | 0 點 | Barbut Helm、Platemail T、Platemail L、腰帶 Greatsword |
| SLEDGE（槽 6） | 全滿 | 0 點 | War Hammer、Skull、Hide、Hide |

人形編號（男 0x47、敵人 0x49）、疲勞直條兩個矩形、裝備列畫在哪一列也相同。裝備名字是原版
`sub_1B94F` 自己解進 `DS:0x2D50` 再印出的那一段，不是測試端另外解碼。

這一隊的裝備都沒有耐久旗標（`+0x0F` 的 0x20），正常戰鬥裡不會壞。所以在同一刻另做一次
受控破損：Olear 的軀幹那一格改成可損、耐久 31，原版由自己的 `sub_22084(0x110, 1)` 扣到壞，
remake 以 `AbsorbArmour` 打到壞；之後兩邊的裝備框都只剩 Elf Bow、Chain Coif、Chain Pants，
Chain Shirt 那一列消失。

```
# 原版（容器內，/src 與 /repo 唯讀）
KOL_CHECK=book KOL_BOOK_SELECT_DOWN=2 KOL_BOOK_UNTIL_WOUND=1 KOL_BOOK_BATTLE=1 KOL_BOOK_WALK=1 \
KOL_BOOK_NEXT_DAY=1 KOL_BOOK_ENTER_FIGHT=1 KOL_BOOK_PLAN=1 KOL_BOOK_EXECUTE=1 \
KOL_BOOK_PLAN_KEYS=.,Return,Return,.,.,Return,.,.,.,Return KOL_BOOK_TIMEOUT=1200s \
bash /src/tools/oracle/check_updated_dosgolem.sh
# remake
tools/run_original_parity_tests.sh          # TestOriginalNormalBattlePanels
```

| 輸入 | SHA-256 |
|---|---|
| `KOLTOWN-unpacked.EXE` | `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| dosgolem | `~/cht/dosgolem-kolrand` `968ca77` |
| 收據 `workplace/battle-panels-20260930/battle-rounds.json`（只留本機） | `36bd46c58e0990149d24867a3e8d8a31114882fc02be5912a03dbdf6f1a20970` |

這份收據的回合邊界、擲骰與訊息與 `battle-first-round-20260928` 那一份逐項相同，只多了
`panels`；`run_original_parity_tests.sh` 的預設改指它，既有的回合、敵人規劃比較照跑通過。

## 不宣稱

- 疲勞直條這一刻三名都在最上緣（分界 84）；分界隨疲勞變化的十二組邊界由
  `TestBattleStatusBarOriginal` 驗，不在這一刻。
- 破損是受控的一步：原版正常戰鬥裡沒有會壞的裝備走到這裡。
- 比的是交給繪製層的參數與原版解出的名字，不是像素截圖。
