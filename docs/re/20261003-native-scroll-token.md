# RE-200：正常入城、卷軸購買與信物頸部槽

## 原始來源與定位

固定 KOLTOWN-unpacked.EXE 的 SHA-256 為
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
使用 dosgolem `968ca774705853276ae55c8d1163c369481ac71f`、Go 1.26.7，
映像為 `eob-remake-go:1.26.7-ebiten2.9.9-audio`。
狀態偏移均為 KOLTOWN DGROUP 相對位移，角色基底 `0xBB88 + 0x300 × n`。
歷史 IDA Pro 9.4 線性定位 `sub_17318:0x17438／0x174BC` 見
[RE-121](RESEARCH-LOG.md#re-121scroll-是入會與買法術法術的母表就在建築-payload-裡)。
這輪沒有新的反組譯；正常原版執行與記錄差分是主要證據。

## 正常玩家路徑與已證實結果

正常 KOL → KOLBOOK 選 Olear、招募 Theodoric、出城，世界位置 40,23。
初次旅店招呼後固定 DGROUP:1600=1／1602=0，不改其他狀態。
按已驗證灰階地圖及 RE-068 規劃步行，正常確認 CAMPING，由 42,9 入城點7進城。
原版與重製有123個共同世界狀態，位置、次格、時間、時段和種子逐步相同。
原版入城後邊界為 X48…63、Y16…31，鏡頭62,27、標記0,0。
正常 L D L L 到 Hostel，加入 Blue Gem 並買 Kelnalon，原版164個狀態。
金額1777→1327→987，教派原始記錄為 `8132a501cad00000b4b1959a5ee4ce80`，
法術記錄為 `85958e0b1ee00c814000000006000a22`。

## 已證實的產品缺口

正常入會把原始信物 `a132a500d014b88201ae000002007000` 寫入角色 `+0x170`。
重製舊候選42cf8db0的 Token 名稱與記錄存在，但 Worn[7] 與 Carried[14] 為空。
`Session.join` 以獨立 Token 判斷槽位，只建立 Name／Record，沒有寫頸部裝備或同步物品。
這是玩家物品頁、裝備明細與原始保存槽的垂直鏈缺口。
RE-199的獨立 Token 欄保存結果保持，不能外推為原版物品槽已對拍。

## 範圍與入口

按 RE-121 的空頸槽優先、占用後找首個空背包槽、全滿不扣錢契約修正。
依 RE-075 已解的記錄欄位建立完整 Good，保留16 byte原始記錄，不新增資產或保存格式。
Token 舊欄保留為入會信物資訊，實際物品位置由 Worn／Backpack／Carried 表示。
原版正常 AREA、其他教派與占用頸槽的正常玩家分支仍待驗。
尚未宣稱本輪新保存或完整戰鬥對拍。

證據工作區：`workplace/spell-native-20261003/`。
正常原版收據為 `shop1/native-shop-private.json`，兩側完整按鍵、原始角色、PNG與驅動器同目錄。
原始角色payload僅本機，未移入Git或交付目錄。
新收據入口預定為 `docs/verification/native-scroll-20261003/verify.py`。

## 2026-10-03 驗收與草案勘誤

上文「尚未宣稱本輪新保存」記錄草案時點；本節追加已完成的實作與全新保存結果。
`token-draft-review.json` 與 `token-evidence-review.json` 保存DRAFT及READY審查雜湊。
正式候選SHA-256為
`68b4ca0e30bf0f4803fc6024bc89ac9b9c9a2e3069b74bcc33e9bfbf1b5be4e3`，
只讀觀察副本為
`dacde28903de8cfc1a2499675d05ae4f61c9f5dbe48badff9bcd18b5ffa586f2`。
319碼輸入中只改scroll.go與scroll_test.go；舊版反例、占用／全滿／格式錯誤及
全部internal／cmd通過。新版正常入城與購買的20個共同城鎮狀態逐槽驗兩人34槽，
世界123個共同狀態另驗位置、次格、時刻、時段與seed，未把世界追蹤宣稱為全欄位對拍。

雙語正常172鍵購買及183鍵頸部頁通過；正常AREA303鍵保持RE-199的實際規則與骰序。
各533鍵住宿生成全新保存，另起正式程式重載各16狀態／PNG，完整34槽、法術、
教派、頸部信物與持久欄位還原。獨立Token僅保留舊資訊相容，物品在Worn[7]與Carried[14]。
早期已寫出的空物品舊保存沒有自動補發，避免重複生成曾移走／丟掉的信物。

正常繁中頸部頁揭露Blue Chain漏譯：名稱工具未掃order.token。
依SPEC-033先DRAFT／READY審查，六個來源加入537鍵／519原名清冊，
四個缺譯原名的舊表反例拒絕。藍鍊、珍珠護符、紅色項鍊、石護符仍為暫譯；
Black Choker沿用黑色頸鍊。原版Red Neclace拼字保留，不改typed data或保存。
新字型來源輸入重生後2608字模及PNG完全相同，9675項安全矩形全過。
184張頸部頁只改一張名稱列431像素；藍鍊32×18字模框零差，目視可讀。
補譯後另跑繁中全新住宿保存／正式重載，語意、保存原名及完整記錄保持。

正常原版輸入固定seed1，只有初次旅店的兩筆seed寫入；未注入角色、座標、指令或等待stub。
缺region1..4與一次int21h AH=0E AL=38診斷保留，不宣稱模擬器全部DOS服務已支援。
原版正常AREA、其他教派、占用頸槽的正常玩家分支與完整進場亂數仍未知。
本輪未做音訊人耳、Windows或macOS真機驗收，也沒有打包或發布。
正式[收據](../verification/native-scroll-20261003/receipt.json)、
[唯讀重驗](../verification/native-scroll-20261003/verify.py)及同目錄audit.json為驗收入口。
