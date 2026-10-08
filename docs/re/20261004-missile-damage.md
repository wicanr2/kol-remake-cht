# RE-218：正常投石命中的力量加成

本輪入口為 `workplace/monster-ranged-rounds-20261004/`。正常啟動固定 seed 1，
接 RE-217 前綴，兩名隊員選休息，原版完成四回合。沒有人物、座標或戰局注入。
259 個正常狀態及 266 張 PNG 保留在本機。前 195 狀態與 201 張前綴 PNG 完全保持。

## 固定來源

| 輸入／工具 | 身分 |
|---|---|
| KOLTOWN.EXE | SHA-256 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221` |
| KOLTOWN-unpacked.EXE | SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| KOLTOWN-unpacked-ida94-derived.i64 | SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8` |
| IDA Pro 9.4 | 原始線性位址；DGROUP 基址 0x2DED0；一次性資料庫、UID 1000 |
| dosgolem | `968ca774705853276ae55c8d1163c369481ac71f`，Go 1.26.7 |

原始命中與 RNG 收據為 `native1/native-shop-private.json`；
IDA 匯出入口 `ida_damage.py`，輸出 `ida-damage.json`。
原始名稱、指令 bytes、xref 與位址保持，payload 僅本機。

## 首個差異與控制流程

- 已證實：第二回合執行入口兩側 seed 為 3278234827；第一發攻擊入口為 3956440006。
  原版第一發後 Damage 4、Hurt 12、右臂 16；舊重製 Damage 16、Hurt 24、右臂 4，觸發擊退。
- 已證實：第一發七次骰的 n 與 seed_before 相符。3d9 為 4、6、1，餘裕加成 1；
  原版護甲骰 2，加值 6，共吸收 8。原版原始傷害 12，護甲後 4。
- 已證實：`sub_26D76:0x27487` 先將力量加成局部變數清零。
  `0x2748C..0x27493` 在目前指令 0x23 時直接跳 0x274D6，跳過力量及疲勞加成。
  舊重製仍加入 `(218−70)÷8×70÷100 = 12`，因此原始傷害成為 24。
- 已證實：`0x27579..0x2758E` 只有持械指令 0x21 加攻擊式傷害。
  `0x27648..0x27665` 預設部位機率列 1，只有持械 0x21 查攻擊式部位表。
  本輪正常投石式為 0；正式窄修正先處理力量加成，不擴張其他式的驗收聲明。

## 驗收與停止線

`review_damage.py` 核對 IDA 指令與固定 EXE 的 MZ 搬移後 bytes，
保存原版三發命中前後的狀態及種子，建立 `normal-hit-cases.json`。
局部測試走正式 `shoot`，同時核對傷害、部位、Hurt、種子、位置及計數。
局部夾具以原版觀察狀態注入，只證明該結算；正常雙語 GUI 另以原版等價按鍵重播。

完整射後裝填別名仍遵循 RE-217 的 DRAFT 限制。後續回合若分歧，先停在首個差異，
不以成功的單元測試宣稱整場或整款遊戲 parity。人物存檔 v3、怪物資料 v2 保持。

## 正式窄修正結果

只改 `attack.go` 的射擊力量加成，新增正常三發反例測試。
334 份 Go、336 個程式輸入，完整 internal／cmd 測試通過。
雙語各 267 鍵／268 狀態，每語言四入口、五名有效參戰者共 760 欄及 40 計數器相符。
4564 次進場及 209 次戰鬥骰的 n／seed_before 與正常原版相同。
四回合傷害 0、8、13、34，怪物最後計數 10／4，第三隻疲勞 54；
原版最後正常畫格與重製最後畫格種子皆 2839412825。
正式／觀察抽樣 PNG 與非 audit trace 相同；SLEDGE 雙語既有 270 狀態／14 PNG 保持。

第一發結果句抽樣：原版 `native1/town-213.png`、英文 `damage-observed-en/step-222.png`、
繁中 `damage-observed-zh-Hant/step-222.png`。結果為右臂嚴重傷口，文字未裁切。
戰場視窗幾何仍有差異，沒有把這三張畫面宣稱像素 parity。

三平台由同一輸入重建，Linux 乾淨重建相同；Windows PE32+ AMD64、macOS amd64／arm64
及 universal 結構、最低 11.0 核對通過。本輪沒有執行 Windows 或 macOS。
來源與二進位完整雜湊在 `program-after-inputs.json`、`promotion.json`、`platform-checks.json`。
收據入口為 `receipt.json`、`comparison.json`、`verify.py`，擁有權與清理在 `audit.json`。

正常耗盡仍待驗。原版 next-planning 鉤子與重製規劃後畫格階段不同，Command、Flag1D
及第三隻 Facing 的終態差異保留，不把不同階段包成完整終態對拍。
完整射後裝填別名仍 DRAFT。父規格與工作項目 6／44 保持開放。
索引 `monster-ranged-address-index.json` 的 0x2748C 記錄固定映像鍵與兩份舊規格回鏈，
`verify.py` 檢查兩份勘誤標記，原始定位及舊形成紀錄保持。
