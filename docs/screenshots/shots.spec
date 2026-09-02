# remake 畫面展示圖的產生規格。每行是「輸出路徑|參數|參數…」，欄位以 `|` 分隔。
# 由 tools/run_kol_shots.sh 在鎖定容器內一次跑完，再由
# tools/build_screenshot_manifest.py 產生 MANIFEST.json。
#
# 英文模式：320×200，與 DOS EGA 原版同解析度、同字模，是逐畫面對拍用的模式。
docs/screenshots/en-title.png|-title-frame|89|-shot|docs/screenshots/en-title.png
docs/screenshots/en-title-knight.png|-title-frame|66|-shot|docs/screenshots/en-title-knight.png
docs/screenshots/en-book-cover.png|-scene|cover|-shot|docs/screenshots/en-book-cover.png
docs/screenshots/en-book-contents.png|-scene|contents|-shot|docs/screenshots/en-book-contents.png
docs/screenshots/en-newchar-stats.png|-scene|script|-script|t e s t RET RET RET RET|-shot|docs/screenshots/en-newchar-stats.png
docs/screenshots/en-newchar-picture.png|-scene|script|-script|z e d RET RET RET RET DOWN RET|-shot|docs/screenshots/en-newchar-picture.png
docs/screenshots/en-book-menu.png|-scene|menu|-menu-selected|1|-shot|docs/screenshots/en-book-menu.png
docs/screenshots/en-book-roster.png|-scene|roster|-shot|docs/screenshots/en-book-roster.png
docs/screenshots/en-book-delete.png|-scene|delete|-delete-script|DOWN RET|-shot|docs/screenshots/en-book-delete.png
docs/screenshots/en-book-harbor.png|-scene|harbor|-shot|docs/screenshots/en-book-harbor.png
docs/screenshots/en-town-inn.png|-scene|inn|-shot|docs/screenshots/en-town-inn.png
docs/screenshots/en-town-inn-party2.png|-scene|inn|-inn|docs/re/town-inn-sample-recruited.json|-shot|docs/screenshots/en-town-inn-party2.png
docs/screenshots/en-town-outdoor.png|-scene|outdoor|-shot|docs/screenshots/en-town-outdoor.png
docs/screenshots/en-town-outdoor-moved.png|-scene|outdoor|-outdoor-camera|58,9|-outdoor-frame|3|-outdoor-marker|0,1|-outdoor-icons|none|-shot|docs/screenshots/en-town-outdoor-moved.png
docs/screenshots/en-town-door.png|-scene|outdoor|-outdoor-camera|57,8|-outdoor-frame|1|-outdoor-at-door|-shot|docs/screenshots/en-town-door.png
docs/screenshots/en-town-building.png|-scene|building|-building-town|7|-building-door|5|-shot|docs/screenshots/en-town-building.png
docs/screenshots/en-town-building-smithy.png|-scene|building|-building-town|7|-building-door|12|-shot|docs/screenshots/en-town-building-smithy.png
#
# 中文模式：640×400，書本美術以最近鄰放大，文字改用 16×18 中文字模格。這個模式不對拍。
docs/screenshots/zh-newchar-picture.png|-lang|zh-Hant|-scene|script|-script|z e d RET RET RET RET DOWN RET|-shot|docs/screenshots/zh-newchar-picture.png
docs/screenshots/zh-book-menu.png|-lang|zh-Hant|-scene|menu|-menu-selected|2|-shot|docs/screenshots/zh-book-menu.png
docs/screenshots/zh-book-roster.png|-lang|zh-Hant|-scene|roster|-shot|docs/screenshots/zh-book-roster.png
docs/screenshots/zh-book-delete.png|-lang|zh-Hant|-scene|delete|-delete-script|DOWN RET|-shot|docs/screenshots/zh-book-delete.png
docs/screenshots/zh-book-harbor.png|-lang|zh-Hant|-scene|harbor|-shot|docs/screenshots/zh-book-harbor.png
docs/screenshots/zh-town-inn.png|-lang|zh-Hant|-scene|inn|-shot|docs/screenshots/zh-town-inn.png
docs/screenshots/zh-town-outdoor.png|-lang|zh-Hant|-scene|outdoor|-shot|docs/screenshots/zh-town-outdoor.png
docs/screenshots/zh-town-door.png|-lang|zh-Hant|-scene|outdoor|-outdoor-camera|57,8|-outdoor-frame|1|-outdoor-at-door|-shot|docs/screenshots/zh-town-door.png
docs/screenshots/zh-town-inn-party2.png|-lang|zh-Hant|-scene|inn|-inn|docs/re/town-inn-sample-recruited.json|-shot|docs/screenshots/zh-town-inn-party2.png
