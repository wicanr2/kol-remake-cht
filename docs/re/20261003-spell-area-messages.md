# RE-206：負面部位法術的傷害結果句

目前狀態：結果句有限CONFORMED。實作前已升READY；全部internal／cmd、正常雙語流程及新保存／重載通過。完整戰鬥仍開放。

工作入口：`workplace/spell-area-messages-20261003/spec-draft.md`、`review_evidence.py`。
來源為 RE-205 正常原版法術流程及本輪 IDA 共用傷害分支匯出。
父收據及共同輸入已封存於 `parent-preservation.json`、`parent-inputs/`。

正式修正契約為 `spec-ready.md`。負面部位法術結果使用共用傷害句，不改擲骰或傷害。
原版正常結果記錄已證實；重傷擊退與遠距部位法術骰序仍待驗，不由本輪補猜。

| 來源 | 身分與範圍 |
|---|---|
| 解包版KOLTOWN-unpacked.EXE | SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| 正式IDA資料庫 | SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8` |
| 工具 | IDA Pro 9.4，image `6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`；dosgolem `968ca774705853276ae55c8d1163c369481ac71f` |
| 位址 | IDA線性位址；DGROUP基底0x2DED0；檔案偏移另列在evidence-review.json |
| 已證實 | sub_2818D:0x285C9類別1、0x285D3呼叫sub_277F4、0x285DB跳返回0x28866；正面0x2869A才跳通用句0x287E4 |
| 正常原版 | RE-205的native1/purchased-effect-private.json，SHA-256 `8f2b0a1334acbad260471859e6078a5170019f7712a1fd91adb60262b391b597`；正常seed1開局，第4回合回答2，傷害11，回答前後seed2128043714 |

正式修正三份既有Go檔及一份新回歸。CastResult保存實際目標與Take結果，AnswerArea及Cast交給同一訊息組合器。沒有新增擲骰、重算傷害或改v3存檔。

驗證結果：

- 舊版雙語訊息反例失敗；新碼20個雙語結果案例、直接部位等待頁及正面／屬性／疲勞通用句通過。全部internal／cmd通過。
- 初版回歸的三個中文期望與既有譯文不同，分類為測試夾具問題。首輪log保留，僅修測試期望後同命令乾淨重跑。
- 英文／繁中各631正常鍵、632狀態與PNG。正式與只讀觀察版全部相同；對RE-205僅step-297訊息不同，其他631圖與觀察角色欄位保持。差異限底部訊息區，英文708像素、繁中1648像素。
- 正式傷害8、回答seed3969673063不變。正式進場亂數與原版仍不同，不宣稱整場同狀態。
- 全新住宿保存只排除saved_at後與前版及觀察版相同。正式重載各16狀態／圖保持，14持久欄位與34個16-byte槽還原。
- 隔離裝備原型各304正常鍵、305狀態，全部觀察角色及seed保持，只改step-304結果頁。傷害11、seed2128043714及結果句符合原版七筆非零記錄。原型不是正式四套格式接線。

目視檢查：繁中step-297顯示目標、胸口與嚴重傷口，文字位於卷軸內。其他結果分支只做規則／適配器回歸，不外推正常原版驗收。

可重驗入口：[verify.py](../verification/spell-area-messages-20261003/verify.py)與同目錄receipt.json。本機工作區含prepare.py、ida_messages.py、review_evidence.py、build_check.sh、run_normal.py／.sh、run_prototype.py／.sh與review_normal.py。原版payload僅本機；正式素材、授權分類與待決事項保持。

剩餘缺口：負面部位法術傷害至少16可能進入原版擊退流程，正式法術路徑尚未接線；遠距部位法術骰序另核；繁中注解位置、怪物資料格式、多法術捲動與完整戰鬥保存parity仍開放。清理與擁有權由本輪audit.json記錄。沒有提交、推送、遠端寫入、打包或發布。
