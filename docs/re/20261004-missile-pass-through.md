# RE-221：投石穿過半途人物

更新日期：2026-10-04。第五回合投石差異已修正；完整八回合閘門仍未通過。

## 固定來源

| 輸入或工具 | 身分 |
|---|---|
| KOLTOWN.EXE | c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221 |
| KOLTOWN-unpacked.EXE | 01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2 |
| KOLTOWN-unpacked-ida94-derived.i64 | 5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8 |
| IDA Pro | 9.4；IDA 線性位址，DGROUP 基址 0x2DED0 |
| dosgolem | 968ca774705853276ae55c8d1163c369481ac71f；Go 1.26.7 |
| 原版正常八回合 | bc9d69f1d1acb03e9dbc6d9c35ca776dde9e6331b87bbc95574e02d95d725a47 |
| 本輪 Linux | 731b266f2770f0baa4ee101f7c00aae15f1ad9d1601ff578de7f5ab59d9137ff |

工作區為 `workplace/missile-pass-through-20261004/`，入口 `prepare.py`、`ida.py`、
`run_oracle.sh`、`review.py`、`run_negative.sh`、`build.sh`／`prepare_run.py`、
`run_normal.py`、`compare.py`、`run_regression.sh`、`prepare_platform.py`、
`cross_build.sh`／`mac_build.sh`／`check_platform.py`、`finalize.py`、`seal.py`／`verify.py`。
本輪只改 missile.go 與既有 missile_test.go。怪物 v2、人物保存 v3 保持。

## 已證實的控制流

正常原版第五回合投石顯示記錄 0x124「那裡沒有東西」。RE-220 正式版停在半途
巨魔，額外擲兩次攻防骰。兩側射擊入口 seed 3413147886 相同，第一骰後為 48124921；
舊版額外兩骰後為 2486282811。

| 原始定位 | 語意與證據等級 |
|---|---|
| sub_25F8A:0x260FC..0x26109 | 已證實：讀人物 +0x0B，無號除以 2，得到碰撞成本 |
| 0x2610C／0x26115 | 已證實：原因寫 2，再對 word_3237C 做 16 位元減法 |
| 0x2611E／0x2612B | 已證實：負值回到 0x260B1，在同一格重新尋找人物及扣除 |
| 0x26123／0x26137 | 已證實：非負設 var_6=1，再回傳 0 |
| sub_25D7A:0x25F62 | 已證實：上述回傳 0 讓外層彈道繼續 |

`ida-final.json` 的 648 條指令全部與固定 EXE 的載入 bytes 核對，保留原名、
operand、xref 與未知警示。原始 bytes 僅本機。附加語意在
[戰鬥位址索引](battle-order-address-index.json)。

dosgolem 局部測試實跑完整 sub_25D7A、sub_25F8A 與人物搜尋，只有飛行繪圖
sub_27FC1 被略過。四種體型 2／80／232／254，各測六種有號分數，共 24 組。
全部走到終點、原因仍為 2、seed=1 保持。體型 80、初分 30 的原版扣 820 次，
回繞為 32766。這份局部測試不能替代正常玩家路徑。

RE-102 與 SPEC-013 第 3／7 節舊「扣一次停下來結果相同」因此被否定。
歷史定位保留，現行契約見 [SPEC-013 勘誤](../spec/013-combat-missiles.md)。
局部測試先後遺漏可穿越地形與人物槽上限的失敗，保存於兩份 initial log；
補齊由 IDA 證實的前置狀態後，以同一容器命令乾淨重跑通過。

零成本且負分數時，原版無法退出。重製保留既有有界停止處理，不宣稱該狀態一致。
本輪不改樹／牆、規劃目標、攻防、格式、保存、操作或飛行動畫契約。

## 有限驗收與剩餘差異

- 舊正式來源在原版反例確實失敗；修正後 24 組對照及全部 internal／cmd 通過。
- 沿用原版 seed=1 的同一正常八回合與雙語各 324 等價鍵，不重設規劃種子、不注入
  戰局、不延長回合。各 325 狀態、142 觀察 PNG、20 正式 PNG；正式與觀察結果相同。
- 原有四回合前綴保持。前七個執行入口每語言 1330 個既定欄位與 70 個計數器相符；
  4564 次進場骰與前 365 次戰鬥骰的骰面、入口 seed 相同。
- 第五回合投石訊息及 seed 48124921 回到原版結果。已核准送禮雙語 300 狀態／300 圖、
  完整 v3 保存重載，以及 SLEDGE 各 270 狀態／14 圖保持。
- 同一來源三平台候選重建，Linux 乾淨重建逐 byte 相同。新 Windows／macOS 只驗
  架構與載入結構，未執行；macOS 最低 11.0。

**完整八回合閘門仍失敗。** 新的第一個位置差異是第七回合擊退：原版退至 29／29，
重製退至 30／28。隨後第八回合 AI 改變，首個骰面差異為原版 rand(39)、
重製 rand(100)，入口 seed 同為 2464785957。第八個執行入口有 14 項既定差異。
等價鍵序終態原版傷害 34／42、seed 3351160206，重製 34／68、seed 429210148，
兩側階段不同，不宣稱完整終態同狀態。下一步查擊退方向的座標更新表。
正常耗盡、完整射後別名、50-byte 全部欄位與其他戰局仍待驗。

## 收據與權利

本機 `receipt.json` 索引前後各 336 個來源輸入、指令輸出、測試、畫格、平台候選與文件。
`verify.py` 唯讀重算 SHA-256、PNG CRC／解壓與 UID/GID，`audit.json` 記錄批末清理。
重驗以 `kol-package:20260907`，專案唯讀掛至 `/src`，執行
`python3 /src/workplace/missile-pass-through-20261004/verify.py`；UID/GID 1000:1000，
無網路、120 秒、2 GiB、2 CPU、128 程序上限。
原版 payload、畫格及候選只在本機研究區。本輪未提交、推送、寫 Issue、打包或發布。
