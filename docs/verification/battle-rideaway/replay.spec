# battle-rideaway 的重跑規格。收據的 frames 逐 byte 比對這一批畫格。
# tools/replay_battle_receipts.py battle-rideaway
# 兩行是同一段路的兩條分支：第一行（種子 141421）騎不掉、接著開打，
# 第二行（種子 1）騎掉了。換過的名冊在 workplace/（不進版控），
# 由 tools/make_observation_inputs.py 重生。
workplace/tmp/rideaway-stood/step-000.png|-scene|play|-seed|141421|-load|workplace/saves/quest.json|-roster|workplace/observe/roster-mounted.json|-play-script|R R U U U U U U U U U RET R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R RET SPACE|-shot-dir|workplace/tmp/rideaway-stood|-shot-steps|88,89,90|-trace|workplace/tmp/rideaway-stood/trace.json
workplace/tmp/rideaway-gone/step-000.png|-scene|play|-seed|1|-load|workplace/saves/quest.json|-roster|workplace/observe/roster-mounted.json|-play-script|R R U U U U U U U U U RET R R R R R R R R R R R R R R R R R R R R RET SPACE|-shot-dir|workplace/tmp/rideaway-gone|-shot-steps|32,33|-trace|workplace/tmp/rideaway-gone/trace.json
