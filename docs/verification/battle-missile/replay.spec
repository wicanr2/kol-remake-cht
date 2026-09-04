# battle-missile 的重跑規格。收據的 frames 逐 byte 比對這一批畫格。
# tools/run_kol_shots.sh docs/verification/battle-missile/replay.spec
# 存檔與換過的資料檔在 workplace/（不進版控），由 make_observation_inputs.py
# 與 import_original_characters.py 重生；輸出寫進 workplace/tmp/missile。
workplace/tmp/missile/step-000.png|-scene|play|-seed|141421|-load|workplace/saves/warder.json|-play-script|R R U U U U U U U U U RET R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R U U U U U U U U U U U U U RET SPACE > RET X RET RET SPACE SPACE SPACE SPACE SPACE SPACE SPACE SPACE SPACE SPACE SPACE SPACE|-shot-dir|workplace/tmp/missile|-shot-steps|100,104,105,106|-trace|workplace/tmp/missile/trace.json
