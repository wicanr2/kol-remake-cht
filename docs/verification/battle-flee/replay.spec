# battle-flee 的重跑規格。收據的 frames 逐 byte 比對這一批畫格。
# tools/run_kol_shots.sh docs/verification/battle-flee/replay.spec
# 存檔與換過的資料檔在 workplace/（不進版控），由 make_observation_inputs.py
# 與 import_original_characters.py 重生；輸出寫進 workplace/tmp/flee。
workplace/tmp/flee/step-000.png|-scene|play|-seed|3|-load|workplace/saves/quest.json|-quest-sites|workplace/observe/quest-sites-edge.json|-play-script|R R U U U U U U U U U RET R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R U U U U U U U U U U U U U RET SPACE < RET RET SPACE SPACE SPACE SPACE SPACE SPACE Y SPACE SPACE SPACE SPACE SPACE SPACE|-shot-dir|workplace/tmp/flee|-shot-steps|97,101|-trace|workplace/tmp/flee/trace.json
