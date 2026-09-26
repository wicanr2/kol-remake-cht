# battle-rideaway 的重跑規格。收據驗的是軌跡上的機制斷言，畫格只作示意。
# tools/replay_battle_receipts.py battle-rideaway
# 兩行是同一段路的兩條分支。騎馬的隊伍在第 32 步時段變換，受控骰值依 RollAmbush 的擲骰順序
# 指定：遭遇成立、不落治療、TROLL、場景，最後一骰是騎走判定（`rand(100) > 5 × 移動率`）：
# 第一行擲 1 跑不掉、第二行擲 100 跑掉（使用者 2026-09-26 決定：野外擲點受控）。
# 換過的名冊在 workplace/（不進版控），由 tools/make_observation_inputs.py 重生。
workplace/tmp/rideaway-stood/step-000.png|-scene|play|-seed|141421|-load|workplace/saves/quest.json|-roster|workplace/observe/roster-mounted.json|-ambush-rolls|1,100,4,1,1|-play-script|R R U U U U U U U U U RET R R R R R R R R R R R R R R R R R R R R RET SPACE|-shot-dir|workplace/tmp/rideaway-stood
workplace/tmp/rideaway-gone/step-000.png|-scene|play|-seed|1|-load|workplace/saves/quest.json|-roster|workplace/observe/roster-mounted.json|-ambush-rolls|1,100,4,1,100|-play-script|R R U U U U U U U U U RET R R R R R R R R R R R R R R R R R R R R RET SPACE|-shot-dir|workplace/tmp/rideaway-gone
