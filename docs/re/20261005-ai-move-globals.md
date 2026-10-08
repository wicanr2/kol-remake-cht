# RE-237：敵人AI舊移動選項與32位元權重

日期：2026-10-05。狀態：局部常式比較及正常重製回歸通過；原版正常GUI仍待驗。

本輪修正三個已證實的差異：高移動選項60、61、62的部位權重列、部位額度修正的byte截斷、
加權輪盤的有號32位元溢位。SPEC-008第8節已有忠實實作授權；先做隔離原型、DRAFT與證據審查，
本機spec-review.json記錄READY後才修改正式來源。沒有新增玩法、資料格式、美術或發行決策。

## 固定輸入

| 項目 | 固定值 |
|---|---|
| 原版輸入 | KOLTOWN-unpacked.EXE |
| SHA-256 | 01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2 |
| dosgolem | 968ca774705853276ae55c8d1163c369481ac71f，從固定git archive建置 |
| 設定種子 | 正常旅店招呼後設1，取得初次AI快照 |
| 樣本生成器 | math/rand種子20260930，每個樣本兩側執行前固定cases[].seed |
| IDA | Pro9.4，原版.i64唯讀，一次性副本；原始名稱與bytes保留 |
| 位址空間 | 下列程式位址均為IDA線性有效位址；DGROUP另標資料段相對位移 |
| 程式身分 | 前後338輸入／336 Go；before source與formal-source逐檔SHA保存 |

原版快照由正常標題、選人、旅店與野外遭遇取得。每個樣本Restore後改寫參戰者、怪物與計數器，
直接呼叫sub_2374C。本輪900組是同一300個輸入分別指定舊全域60、61、62，
另重生既有300組attack-only正對照。每組遊戲seed在執行前固定，沒有重擲或挑結果。
state_injection及direct_routine_entry均為true，normal_gui_acceptance為false。

## 原版證據與等級

- 已證實：sub_2590F:0x25943..0x2595B讀DGROUP:0x0FF4+10*style的七個byte。
  三個高列為`3 4 5 0 0 0 0`、`2 3 2 3 4 0 96`、`91 96 110 116 78 97 91`。
  原版執行時整份640-byte緩衝與既有IDA匯出相同；唯讀觀察增加後，1200組原版結果完全相同。
- 已證實：0x25A14使用AL存回部位權重，額度修正後只保留低8位元。
- 已證實：sub_25A3A:0x25AA0、0x25B5F與0x25C42呼叫sub_2DE8C，乘積只返回DX:AX。
  0x25B88／0x25B8C用add／adc加總兩個word。0x25BC4檢查高字符號，再決定是否總和大於65535。
  sub_2DDF0:0x2DDFB及0x2DE13檢查兩側高字符號，還原有號商，除法向零截斷。
  位元組與原operand見ida-weight.json／ida-weight-helpers.json／ida-location-scaling.json；追加語意已合併
  docs/re/player-gaps-address-index.json，未確認的其餘指令保留警示。
- 已證實：樣本119改Level為5，舊測試LoadLevel仍為1。原版負重50／花費300，
  舊夾具負重70／花費420。同步LoadLevel後既有300組正對照恢復。
- 未知：舊任意壓力的35、46、24是否由正常玩家流程留下；三組差異保持，不補猜測列。

## 反例與驗證

| 階段 | 結果 |
|---|---|
| 原碼與舊夾具 | 各批都在119失敗 |
| 僅清遠程槽的初解釋 | 仍在119失敗，已撤回原因判斷 |
| 同步負重等級 | 正對照通過，高列共9差異 |
| 高列與byte截斷原型 | 60、61通過；62的59仍不同 |
| 有號32位元原型與正式 | 高列900＋正對照300全部一致，原版呼叫失敗0 |
| 舊任意壓力300 | 41、59、166三個差異保留 |
| 正式全部cmd／internal Go測試 | 通過，缺原版環境的oracle測試會略過 |
| 最新正式雙語正常任務與新程序重載 | 14954任務＋26重載狀態；完整trace與前輪相同 |
| 完整人物v3與32張PNG | 僅saved_at不同；全部畫面解碼後相同 |

原版兩次各1200呼叫合計2400，失敗0。高列900組只有12次實際讀高列，三個值各4次。
本輪比較擲骰引數、亂數終態、指令／攻防選項、朝向、撤退及撿起旗標；未核全50-byte記錄。
正常任務沿RE-235既定7476鍵、seed1與200個明示遭遇骰100，沒有中途重播種。
正常DROP、取得Ship Wheel、114回合逃離、六人回報、Moraine付56及六人保存重載保持。
這是正常重製回歸，沒有借用兩人原版快照宣稱六人任務同狀態。
初批外層1800秒共用於建置、全部測試及雙語長鏈，繁中尚未結束即中止，退出124。
這是驗證wrapper時限錯誤。保留initial-zh-wrapper-timeout後，只以相同image、資源、程式、
seed及完全相同command重跑繁中；獨立外層2400秒。未改測試條件、按鍵或結果判準。
中止前11張PNG與成功重跑相同，沒有把中止當產品缺陷或改寫成通過。

## 入口與限制

本機工作區`workplace/ai-move-globals-20261005/`：
prepare.py／build.sh／run-original.py重生原版；prepare-context.py／build-context.sh加入唯讀觀察；
build-remake.sh、build-fixture.sh、build-prototypes.sh及prepare-arithmetic.py保留失敗與局部原型；
ida-weight.py／ida-weight-helpers.py／ida-location-scaling.py匯出最小運算證據；
spec-draft.md／spec-review.json保存草案與審查；promote.py記錄兩檔正式修改；
prepare-formal.py／build-formal.sh建置、測試、900＋300比較與雙語任務；
run-zh-retry.sh只重跑時限中止的繁中流程；首次完整建置建議外層至少4200秒，
兩個正常程序各有1800秒內層界限，Xvfb皆有trap與容器生命週期。
review-formal.py／update-docs.py／seal.py／verify.py保存結果、文件快照與雜湊。
既有image為eob-remake-go:1.26.7-ebiten2.9.9-audio與ida-pro-9.4-idapython:locked-v1。
所有工作在network none、UID/GID1000、有界Docker執行；原始輸入唯讀，明確工作區可寫。

正式Linux SHA-256為`38f86f38b1531c37f436e357aca251e83e5c85ff1236992dbc5e5add2337db16`，只改internal/battle/ai.go與既有original_ai_fuzz_test.go。
怪物v2、人物v3、HD手繪決定及尚未確認的畫面比例保持。沒有新版封包、公開散布或平台真機驗收。
#11、整場與全任務原版對拍、玩家可見撤退／撿武器與側翼GUI保持開放。
原始資料、反組譯與新PNG只留本機；receipt.json／audit.json記錄附件、輸入與清理狀態。
