# remake 畫面展示圖的產生規格。每行是「輸出路徑|參數|參數…」，欄位以 `|` 分隔。
# 由 tools/run_kol_shots.sh 在鎖定容器內一次跑完，再由
# tools/build_screenshot_manifest.py 產生 MANIFEST.json。
#
# 英文模式：320×200，與 DOS EGA 原版同解析度、同字模，是逐畫面對拍用的模式。
docs/screenshots/en-title.png|-lang|en|-title-frame|89|-shot|docs/screenshots/en-title.png
docs/screenshots/en-title-knight.png|-lang|en|-title-frame|66|-shot|docs/screenshots/en-title-knight.png
docs/screenshots/en-book-cover.png|-lang|en|-scene|cover|-shot|docs/screenshots/en-book-cover.png
docs/screenshots/en-book-contents.png|-lang|en|-scene|contents|-shot|docs/screenshots/en-book-contents.png
docs/screenshots/en-newchar-stats.png|-lang|en|-scene|script|-script|t e s t RET RET RET RET|-shot|docs/screenshots/en-newchar-stats.png
docs/screenshots/en-newchar-picture.png|-lang|en|-scene|script|-script|z e d RET RET RET RET DOWN RET|-shot|docs/screenshots/en-newchar-picture.png
docs/screenshots/en-book-menu.png|-lang|en|-scene|menu|-menu-selected|1|-shot|docs/screenshots/en-book-menu.png
docs/screenshots/en-book-roster.png|-lang|en|-scene|roster|-shot|docs/screenshots/en-book-roster.png
docs/screenshots/en-book-delete.png|-lang|en|-scene|delete|-delete-script|DOWN RET|-shot|docs/screenshots/en-book-delete.png
docs/screenshots/en-book-harbor.png|-lang|en|-scene|harbor|-shot|docs/screenshots/en-book-harbor.png
docs/screenshots/en-town-inn.png|-lang|en|-scene|inn|-shot|docs/screenshots/en-town-inn.png
docs/screenshots/en-town-inn-party2.png|-lang|en|-scene|inn|-inn|docs/re/town-inn-sample-recruited.json|-shot|docs/screenshots/en-town-inn-party2.png
docs/screenshots/en-town-outdoor.png|-lang|en|-scene|outdoor|-shot|docs/screenshots/en-town-outdoor.png
docs/screenshots/en-town-outdoor-moved.png|-lang|en|-scene|outdoor|-outdoor-camera|58,9|-outdoor-frame|3|-outdoor-marker|0,1|-outdoor-icons|none|-shot|docs/screenshots/en-town-outdoor-moved.png
docs/screenshots/en-town-door.png|-lang|en|-scene|outdoor|-outdoor-camera|57,8|-outdoor-frame|1|-outdoor-at-door|-shot|docs/screenshots/en-town-door.png
docs/screenshots/en-town-building.png|-lang|en|-scene|building|-building-town|7|-building-door|5|-shot|docs/screenshots/en-town-building.png
docs/screenshots/en-town-building-smithy.png|-lang|en|-scene|building|-building-town|7|-building-door|12|-shot|docs/screenshots/en-town-building-smithy.png
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
docs/screenshots/zh-town-building.png|-lang|zh-Hant|-scene|building|-building-town|7|-building-door|5|-shot|docs/screenshots/zh-town-building.png
docs/screenshots/zh-town-building-smithy.png|-lang|zh-Hant|-scene|building|-building-town|7|-building-door|12|-shot|docs/screenshots/zh-town-building-smithy.png
docs/screenshots/zh-town-inn-party2.png|-lang|zh-Hant|-scene|inn|-inn|docs/re/town-inn-sample-recruited.json|-shot|docs/screenshots/zh-town-inn-party2.png
docs/screenshots/zh-town-brettle-inn.png|-lang|zh-Hant|-scene|building|-building-town|1|-building-door|1|-shot|docs/screenshots/zh-town-brettle-inn.png
docs/screenshots/zh-town-olanth-bakery.png|-lang|zh-Hant|-scene|building|-building-town|6|-building-door|11|-shot|docs/screenshots/zh-town-olanth-bakery.png
docs/screenshots/en-town-edge.png|-lang|en|-scene|outdoor|-outdoor-camera|58,0|-outdoor-marker|0,1|-outdoor-frame|3|-outdoor-at-edge|-outdoor-icons|5,18|-shot|docs/screenshots/en-town-edge.png
docs/screenshots/zh-town-listen-ask.png|-lang|zh-Hant|-scene|play|-outdoor|docs/re/town1-outdoor-sample.json|-play-script|L R RET SPACE LISTEN|-shot|docs/screenshots/zh-town-listen-ask.png
docs/screenshots/zh-town-listen-answer.png|-lang|zh-Hant|-scene|play|-outdoor|docs/re/town1-outdoor-sample.json|-play-script|L R RET SPACE LISTEN SPACE|-shot|docs/screenshots/zh-town-listen-answer.png
docs/screenshots/en-world-map.png|-lang|en|-scene|world|-shot|docs/screenshots/en-world-map.png
docs/screenshots/en-world-town.png|-lang|en|-scene|play|-play-script|R R U U U U U U U U U RET R R R R R L|-shot|docs/screenshots/en-world-town.png
docs/screenshots/zh-world-map.png|-lang|zh-Hant|-scene|world|-shot|docs/screenshots/zh-world-map.png
docs/screenshots/zh-world-town.png|-lang|zh-Hant|-scene|play|-play-script|R R U U U U U U U U U RET R R R R R L|-shot|docs/screenshots/zh-world-town.png
docs/screenshots/en-town-sheet.png|-lang|en|-scene|play|-play-script|L R RET SPACE > RET|-shot|docs/screenshots/en-town-sheet.png
docs/screenshots/zh-town-sheet.png|-lang|zh-Hant|-scene|play|-play-script|L R RET SPACE > RET|-shot|docs/screenshots/zh-town-sheet.png
docs/screenshots/zh-book-contents.png|-lang|zh-Hant|-scene|contents|-shot|docs/screenshots/zh-book-contents.png
docs/screenshots/zh-newchar-stats.png|-lang|zh-Hant|-scene|script|-script|t e s t RET RET RET RET|-shot|docs/screenshots/zh-newchar-stats.png
docs/screenshots/zh-book-cover.png|-lang|zh-Hant|-scene|cover|-shot|docs/screenshots/zh-book-cover.png
docs/screenshots/en-town-items.png|-lang|en|-scene|play|-play-script|L R RET SPACE > RET > RET|-shot|docs/screenshots/en-town-items.png
docs/screenshots/zh-town-items.png|-lang|zh-Hant|-scene|play|-play-script|L R RET SPACE > RET > RET|-shot|docs/screenshots/zh-town-items.png
docs/screenshots/en-town-party-page.png|-lang|en|-scene|play|-play-script|L R RET SPACE > RET > > RET|-shot|docs/screenshots/en-town-party-page.png
docs/screenshots/zh-town-party-page.png|-lang|zh-Hant|-scene|play|-play-script|L R RET SPACE > RET > > RET|-shot|docs/screenshots/zh-town-party-page.png
