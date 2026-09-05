# battle-pickup 的重跑規格。收據的 frames 逐 byte 比對這一批畫格。
# tools/replay_battle_receipts.py battle-pickup
# 兩行是同一段路的兩種回答：第一行答是（撿起來），第二行答否（留在地上）。
# 存檔在 workplace/（不進版控），由 import_original_characters.py 重生。
workplace/tmp/pickup-yes/step-000.png|-scene|play|-seed|141421|-load|workplace/saves/quest.json|-play-script|R R U U U U U U U U U RET R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R > > > > RET RET SPACE SPACE SPACE SPACE SPACE > > RET RET SPACE SPACE SPACE SPACE Y|-shot-dir|workplace/tmp/pickup-yes|-shot-steps|110,112,120,122|-trace|workplace/tmp/pickup-yes/trace.json
workplace/tmp/pickup-no/step-000.png|-scene|play|-seed|141421|-load|workplace/saves/quest.json|-play-script|R R U U U U U U U U U RET R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R > > > > RET RET SPACE SPACE SPACE SPACE SPACE > > RET RET SPACE SPACE SPACE SPACE N|-shot-dir|workplace/tmp/pickup-no|-shot-steps|122|-trace|workplace/tmp/pickup-no/trace.json
