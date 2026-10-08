# RE-202：正常伏擊的進場位置校正

日期：2026-10-03。前批入口：[RE-201](20261003-purchased-spell-effect.md)。
本輪只追同一正常購買、出城、兩步伏擊的地形及怪物位置，不要求全場骰序一致。
本機研究輸出位於 `workplace/ambush-placement-20261003/`。

## 輸入與定位

- `KOLTOWN-unpacked.EXE` SHA-256：
  `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
- 正式衍生 `.i64` SHA-256：
  `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`。
- IDA Pro 9.4，映像 `ida-pro-9.4-idapython:locked-v1`，映像 SHA-256
  `6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`。
  原始名稱 `sub_1E3A8` 保留；位址皆為 IDA linear，DGROUP 基準 `0x2DED0`。
- dosgolem 固定修訂 `968ca774705853276ae55c8d1163c369481ac71f`。
  原版只在正常旅店招呼後設定 seed1；重製版以 `-seed 1` 執行同一正常按鍵。
  不注入角色、物品、位置、回傳值或骰序，不略過等待。

## 原始證據

`ida-placement.json` 保存指令、原始 bytes 及直接交叉參照。
`ida-placement-tables.json` 保存兩張位移表、原始 bytes 與站位分支。
間接 `[bx+1Fh]`／`[bx+20h]` 寫入由資料流及正常執行斷點補證，不用直接 xref 缺項否定。

| 定位 | 附加語意 | 推論等級 |
|---|---|---|
| `0x1EA13..0x1EA22` | 每隻敵人擲 RND(8)，減一作起點 | 已證實 |
| `0x1EA2B..0x1EA78` | 從當時的位置加兩張位移表，形成候選座標 | 已證實 |
| `0x1EA84..0x1EAFB` | 掃描有效參戰者；候選已占則跳過 | 已證實 |
| `0x1EAFE..0x1EB56` | 以 `Y*64+X` 讀地形；非 `0x0A` 才寫回位置 | 已證實 |
| `0x1EB59..0x1EB78` | 方向加一並環繞；完整走八個方向，成功後仍續查 | 已證實 |
| DGROUP `0x0C1C`／`0x0C2C` | X／Y 位移，原 bytes 留在表匯出 | 已證實 |

## 規格草案

DRAFT 已依正常擷取與原始 EXE bytes 審查升為 READY。
審查保存於本機 `evidence-review.json`：175 狀態及 PNG 保持，24 次候選檢查、
4096 地形格、校正前後種子與三隻怪物位置一致；814 道指令符合原始 EXE bytes。

正常伏擊在確認遭遇圖後，以野外擲完數量、等級、陣型的亂數狀態生成地形。
同一亂數來源接著依敵人槽位校正位置，再交給既有規劃流程。
每隻敵人只擲一次八面骰。每一步以當時的位置計算候選，跳過已占格與 `0x0A`。
成功就立刻寫回，然後繼續下一方向，總共八次。隊員位置不動，重排不再校正。
任務與競技場雖共用原版區段，本輪正式接線先限正常野外伏擊。

既有正常原版骰序的局部重播得到：位置校正前 seed `2052235245`，
校正後 `4058265942`；三隻敵人由 `(36,32)`、`(36,30)`、`(36,34)`
變成 `(35,32)`、`(37,29)`、`(36,34)`。新正常擷取已逐步核對此局部重播。
原版後續 `0x1EBEA` 的三次四面骰另有裝備資料用途，本輪不以丟棄骰值補齊。
完整首次規劃、怪物裝備變體及戰鬥後保存仍由父閘門追蹤。

正常驅動器第一次使用不存在的 `SS()` API，編譯失敗；第二次在核對後發現
`StackWord` 以 SP 為基準，於進入玩法前停止。第三次使用 `Regs().SS:BP-2`，
在同映像、同命令及同 seed 重跑通過。前兩次輸入與紀錄留在 `native1/`、`native2/`。
審查初次錯把既有 `int 21h AH=0E AL=38 ×1` 診斷要求為空，已核對與 RE-201 相同。
這些是驗證工具問題；未新增產品缺陷，也沒有挑選 seed。

## 正式驗收與剩餘閘門

正式候選 `eed01abbbff335e4a0944d670a61c668d0be5ce5a93fbea12f2e9561e273b043`
通過舊版反例、正常已有地形／確認時生成兩條回歸與全部內部／主程式測試。
只讀副本 `b5d8912e17b2461880ab5740aa158c19e95af8b9bc580fc46e9a2cd0f9fef1f4`
只觀察正式狀態，沒有規劃播種或物品／座標注入。

英文／繁中各 183 正常鍵、184 狀態及 PNG。正式版各 184 PNG 與觀察副本相同。
4096 地形格、兩名隊員及三隻敵人位置符合本輪正常原版。
英文前 183 張畫格保持；完整 144×144 戰場視窗 20,736 像素零差異。
繁中以實際三倍 PNG 取像素塊中心後有 55 像素不同，未用遮罩，也不稱繁中逐像素相同。
本場正常伏擊的地形交接與位置校正有限 CONFORMED；SPEC-011 父範圍維持 READY。

完整首次規劃及怪物裝備變體仍未完成。首次隊員同為 Theodoric；首次規劃 seed
原版 `2535178238`、重製 `1323149826`。下一步比較初始化後的玩家可見屬性與裝備。
不追全場亂數呼叫數，不丟棄骰值補齊，不換 seed 挑結果。
任務／競技場、本輪新住宿保存、其他戰鬥分支及負面 AREA 未由本場驗收。
319 正式碼輸入只改四份；315 份與前批保持，沒有新增正式資料格式。
本機 #6／#44 仍開放，22 未完成／84 完成；沒有提交、推送、遠端寫入、封包或發布。

唯讀入口：[verify.py](../verification/ambush-placement-20261003/verify.py)。
[receipt.json](../verification/ambush-placement-20261003/receipt.json)保存輸入與產物雜湊；
[audit.json](../verification/ambush-placement-20261003/audit.json)保存批末 Git、Docker、擁有權與暫存檢查。
稽核可由 [audit.py](../verification/ambush-placement-20261003/audit.py)唯讀重跑。

兩支 Python 工具只在 Docker 執行，將本專案唯讀掛到 `/src`，沿用
`kol-package:20260907`，設定 `--rm --network none --user 1000:1000 --memory 1536m`
及 `--cpus 2 --pids-limit 128`，外層 `timeout 240s`。
執行入口分別為 `python3 /src/docs/verification/ambush-placement-20261003/verify.py`
與 `python3 /src/docs/verification/ambush-placement-20261003/audit.py`。
深層歷史的初次重算另以十一份 `parent-inputs/` 檔唯讀覆掛原 `/src` 位址，
防止舊驗證器直接讀主路徑；結果及控制面保存在本輪 `parent-check.json`、`audit-control.json`。
一般重驗核對所有封存輸入、附件及本輪結果，不重新執行原版或修改凍結的歷史。
