# RE-263：受阻打暈的舊地面名稱原型

日期：2026-10-06。狀態：DRAFT，隔離原型，不接正式。

路由命中 remake 收尾，沿 `reverse-engineer-retro-game-remake` 的
`references/verification-and-handoff.md`、文件職責與 `grilling` 技能。
本輪承接 [RE-262](20261006-stale-ground-name.md) 的已證實來源。
受阻持物規則沿 RE-254 既有問題等待確認，不提出重複問題或升格 READY。

## 來源與隔離範圍

原始輸入 `KOLTOWN-unpacked.EXE` SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
IDA Pro 9.4 線性位址、載入基底 `0x10000`、DGROUP 基底 `0x2DED0`。
資料庫 SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`。
沿既有匯出 `drop-name-private.json` 的原始名稱與 bytes，未新增 IDA 執行。

成功強制掉落的 `sub_269BA:0x26A65` 寫 ReturnScratch，
`0x26A6B` 另寫 `word_30D02` 的 H 索引。
受阻分支不呼叫該函式，H 訊息消費端依 RE-262 讀目前地面槽。
空手、滿槽或未成功搬移時，原型保持既有槽號。

在 RE-260 的受阻持物候選上，只覆蓋三個既有程式檔：

- `state.go` 增加本場私有地面槽號。
- `items.go` 在成功強制掉落後記住槽號，隊員與怪物共用語意。
- `knockback.go` 在受阻持械句讀該槽的現有 typed 名稱。

槽被清空後顯示空白，不快取物品文字。正式人物 v3／怪物 v2 保持。
初始 H=0 限這兩條已觀察的首場；跨場共用 H、戰鬥中途保存及殘留空記錄
未由本輪證實。成功擊退滿槽訊息也不列入本原型完成聲明。

## 局部反例

四個測試組合為非空／已清槽，以及正常容量／後續滿槽。
先成功強制掉落完整 Broadsword 到槽 2，再受阻打暈仍持 Battle Axe 的隊員。
舊候選在兩個非空槽案例輸出空白而失敗；新原型四個案例通過。
兩個既有隊員／怪物完整持物反例及正常原版第五／第七回合局部回歸保持。
綠色局部測試不當原版正常戰鬥完成證據。

## 正常玩家路徑

固定 setup seed 2，沿既有計畫，沒有更換種子或注入狀態：

- RE-260 正常招募、旅行與主動近戰，雙語各 809 狀態。
- RE-254 正常受阻、返回世界、回城住宿，舊候選與新原型各跑雙語。
- 每個真正 v3 保存另開新程序，以正常名冊選人與外出操作重載。

使用 `eob-remake-go:1.26.7-ebiten2.9.9-audio`，UID/GID 1000:1000，
Docker 無網路，資源與逾時有界。來源 `/src` 唯讀，研究輸出 `/out` 可寫。
Xvfb 有 socket 檢查及 trap；每個玩家程序有獨立終態收據，容器以 `--rm` 清理。
正常輸入的已知等效轉換沿 RE-254，不聲稱原版方向或滑鼠操作相同。

工作區為 `workplace/stale-ground-prototype-20261006/`，入口 `prepare.py`、
`build.sh`、`run.sh`、`review.py`。正式 338 輸入與 Linux
`1f31007c19ceac875a12df9b68d8629296764d01194983cf98d10cc8dcdbd274` 保持。
本地原始資料與衍生比較畫格保持僅本機，不擴張公開或 private PNG 授權。
正式十九套件沿 RE-258，沒有把隔離原型單測冒充全量正式回歸。

完成核對後的有限結果、收據與原圖比較入口接在本節後。

## 有限驗收結果

候選SHA-256：`6b5c866ebd4a1888662220cb9f6251572974dbfd6fa2c6f7b80341304a97a60f`。

六次正常流程與四次新程序重載共4218狀態。主動戰鬥1618狀態與前版相同，只有每語言第733步的H名稱改為Broadsword／闊劍。605個人物位置、頭胸與496個規劃選單相符。

原版第730格與候選第733步的訊息矩形24／165／272／20，前版差597像素，候選差0。繁中正常畫格已目視核對闊劍與完整句，未宣稱與原版繁中逐像素相同。

正常回城保存前版／候選雙語共2548狀態，完整trace相同；完整v3只排saved_at。四次真正新程序重載共52狀態，兩版trace、PNG及保存SHA保持。50張PNG驗CRC、解碼與尺寸。

原型仍DRAFT，等待既有受阻持物規則確認。原版仍十四回合戰敗，勝利S結尾、完整50-byte、跨場H、戰鬥中途保存、完整UI及正式交付仍未完成。

收據：[receipt.json](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/stale-ground-prototype-20261006/receipt.json)。並列原圖：[review.html](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/stale-ground-prototype-20261006/review.html)。唯讀重驗：[verify.py](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/stale-ground-prototype-20261006/verify.py)。
