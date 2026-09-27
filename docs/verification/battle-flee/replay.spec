# battle-flee 的重跑規格。收據驗的是軌跡上的機制斷言，畫格只作示意。
# 路上的時段變換以受控骰值 100 讓遭遇不成立（使用者 2026-09-26 決定：野外擲點受控，戰場照種子）。
# tools/run_kol_shots.sh docs/verification/battle-flee/replay.spec
# 存檔與換過的資料檔在 workplace/（不進版控），由 make_observation_inputs.py
# 與 import_original_characters.py 重生；輸出寫進 workplace/tmp/flee。
workplace/tmp/flee/step-000.png|-scene|play|-seed|3|-load|workplace/saves/quest.json|-quest-sites|workplace/observe/quest-sites-edge.json|-ambush-rolls|100,100,100,100,100,100|-play-script|R R U U U U U U U U U RET R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R R U U U U U U U U U U U U U RET SPACE < RET RET SPACE Y|-shot-dir|workplace/tmp/flee|-shot-steps|97,101|-trace|workplace/tmp/flee/trace.json
