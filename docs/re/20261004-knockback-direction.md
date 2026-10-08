# RE-219：正常第五回合擊退方向

更新日期：2026-10-04。正式擊退索引已修正；八回合閘門未通過。
SPEC-009 維持 READY，不把整場射擊或完整玩家鏈升格。

## 固定來源

| 輸入或工具 | 身分 |
|---|---|
| KOLTOWN.EXE | c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221 |
| KOLTOWN-unpacked.EXE | 01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2 |
| KOLTOWN-unpacked-ida94-derived.i64 | 5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8 |
| IDA Pro | 9.4，原始線性位址；DGROUP 基址 0x2DED0 |
| dosgolem | 968ca774705853276ae55c8d1163c369481ac71f，Go 1.26.7 |
| 原版八回合收據 | bc9d69f1d1acb03e9dbc6d9c35ca776dde9e6331b87bbc95574e02d95d725a47 |

工具入口在 `workplace/monster-ranged-exhaustion-20261004/`。
`prepare.py`／`run_native.sh` 從正常住宿前綴以 seed 1 最多跑八回合或正常出口。
沒有 direct-entry、狀態注入、等待替換或規劃種子重設。
原版 316 狀態、323 PNG；RE-218 前 259 狀態及 265 前綴 PNG 完全保持。
兩名巨魔仍剩 8／3 次投石，未發生正常耗盡。

## 已證實的勘誤

`sub_277F4:0x27BE9..0x27C30` 讀兩人的 `+0x1F`，目標減攻擊者後
選 0／4／8。`0x27C43..0x27C81` 讀 `+0x20`，負差加 1、正差加 2。
它們分別對應正式 Actor.X／Actor.Y。RE-095 的軸交換說法由原始 operand
及正常第五回合座標否定，原註解與歷史證據仍可回查。
DGROUP:0x123A 的 33 bytes 保持；0x27D3A／0x27D5A 分別更新 +0x1F／+0x20。

`ida_knockback.py` 匯出 245 道指令與方向表，保留原名、operand、bytes、xref。
`review_knockback.py` 用 EXE 的 MZ relocation 核對載入 bytes，再審查為 READY。
原始指令及角色 payload 僅本機保存；附加語意在
[戰鬥位址索引](battle-order-address-index.json)。

正常第五回合攻擊者 32／30、目標 31／30；原版退至 30／29，舊版退至 30／30。
兩次擊退骰的入口 seed 1535194068、終 seed 3413147886 相同。
新增原版反例在舊程式確實失敗，修正後通過。
只改 KnockbackIndex、原版回歸及法術頁測試舊座標期待，共三檔。
人物存檔 v3 與怪物資料 v2 保持。

## 正常重播與剩餘差異

雙語各 324 正常鍵、325 狀態、觀察 142 PNG、正式 16 PNG。
正式 trace 去除觀察欄後相同，正式抽樣圖完全相同；RE-218 前四回合保持。
前五個執行入口每語言 950 欄與 50 計數器相符。
到第五回合擊退完成的 4564 進場骰、240 戰鬥骰相符，座標 30／29；
37 個既定人物欄相符。此抽樣不代表 50-byte 全記錄。

**八回合閘門未通過。** 第五回合下一次投石，原版 seed 48124921，
重製 2486282811。第六回合執行入口原版 4271209092，重製 4121420029。
八回合等價鍵序的原版傷害 34／42，重製 34／38；重製尚未完成第八回合。
不能用修正後局部通過取代預先登記的八回合驗收。
下一步定位該次投石的目標、路徑與碰撞，證據 READY 前不猜補。

全部 internal／cmd 與雙語 SLEDGE 270 狀態／14 PNG 回歸通過。
三平台同來源重建，Linux 乾淨重建逐 byte 相同；新 Windows／macOS 僅驗結構。
本輪 Wine 雙語四回合的 268 狀態／11 PNG 相同，驗的是修正前 RE-218 Windows，
不得把這份執行結果移用到新候選，也不代表 Windows 實機。

## 重驗與權利

本機 `receipt.json`、`verify.py`、`audit.json` 保存附件、唯讀重驗與批末清理。
唯讀重驗使用 `kol-package:20260907`，專案掛至 `/src:ro`，
執行 `python3 /src/workplace/monster-ranged-exhaustion-20261004/verify.py`。
UID/GID 1000:1000、無網路、120 秒、1 GiB、1 CPU、64 程序上限。
原始 bytes、人物與畫格僅本機研究；本輪未提交、推送、寫 Issue、打包或發布。
