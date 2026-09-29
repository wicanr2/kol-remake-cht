# battle-pickup 的重跑規格。收據驗的是軌跡上的機制斷言，畫格只作示意。
# tools/replay_battle_receipts.py battle-pickup
# 兩行是同一段路的兩種回答：第一行答是（撿起來），第二行答否（留在地上）。
# 伏擊照種子自然發生（第 129 步時段變換遇到 OGRE；路上先有路過的通知與治療僧侶，答應治療），不用受控骰值。
# 存檔在 workplace/（不進版控），由 import_original_characters.py 重生。
workplace/tmp/pickup-yes/step-000.png|-scene|play|-seed|141421|-load|workplace/saves/quest.json|-play-script|R R U U U U U U U U U RET R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R RET R R R R R R R R R R R R R R R R R R R R R R R R R R R R SPACE > > > > RET RET RET SPACE SPACE SPACE SPACE SPACE > > RET RET RET SPACE SPACE SPACE SPACE Y SPACE SPACE SPACE SPACE|-shot-dir|workplace/tmp/pickup-yes
workplace/tmp/pickup-no/step-000.png|-scene|play|-seed|141421|-load|workplace/saves/quest.json|-play-script|R R U U U U U U U U U RET R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R RET R R R R R R R R R R R R R R R R R R R R R R R R R R R R SPACE > > > > RET RET RET SPACE SPACE SPACE SPACE SPACE > > RET RET RET SPACE SPACE SPACE SPACE N SPACE SPACE SPACE SPACE|-shot-dir|workplace/tmp/pickup-no
