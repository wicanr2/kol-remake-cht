# RE-213：修改法術後的原版住宿保存

日期：2026-10-03。已取得原版正常修改法術後的實際保存與重載，正式325份Go及56bfb4c8建置保持。
前批入口：[RE-212](20261003-spell-name-original.md)。
本研究沿用正常玩家試玩路由與reverse-engineer-retro-game-remake技能。
工作區為`workplace/spell-edit-inn-save-20261003/`，輸入與來源唯讀，輸出僅本機。

## 擷取過程

先封存338份共用輸入，保持前批9313附件與原始收據。
原版固定dosgolem revision 968ca774705853276ae55c8d1163c369481ac71f，
正常旅店招呼後seed1，編輯器／世界／住宿不再重設。
正常購買Kelnalon、改名N及Body→Strength、答應存錄後，沿既有世界返回路線回旅店。
`SetScratch`只使原版實際寫檔落地到本工作區，不就地修改來源，不注入角色、物品、座標或勝利。
首次擷取在離城第二次Left遇到3隻Cliff Troll，停在正常戰鬥入口。
後續擷取依實際圖示選休息、確認及逃離，再按Y；依實際世界返回位置繼續走回旅館。
戰鬥入口不重設seed，不以勝利、座標或直接呼叫常式替代玩家輸入。

開工時待驗：原版真正寫出的CHARDATA、法術名稱／效果／成本與其他持久欄位，以及正常再次招募。
沒有打包、發布、發行聲明或新產品決定。原版payload與未驗圖片維持僅本機。
工具入口：`prepare.py`、`prepare_native.py`、`prepare_native2.py`、`prepare_native3.py`、`prepare_native4.py`、`prepare_native_reload.py`、`run_native.sh`；命令與keys逐份保存在工作區。
重製按鍵規劃工具為`prepare_adaptive.py`、`adaptive_driver.go`及`run_adaptive.sh`。
規劃器只讀狀態並送正式按鍵；取得路線後仍須以未改動的正式程式重播雙語保存及重載。

驗證腳本限制：一般野外SPACE在互動版會被忽略，批次介面會報未知鍵。
先前靜態腳本在該鍵停止；另一份規劃器雖跑完488鍵，但將旅行鍵繼續送進戰鬥，沒有保存。
兩者均不算住宿完成。檔案複製缺少執行權限及188鍵前綴中止，另存環境／腳本紀錄。
`native-inn2`的120鍵逃離腳本止於實際caller IDA linear 0x1FBDF。
0x2114D只涵蓋舊執行段提示，不能套用到回合逃離圖示；修正後從相同前綴重跑。

重製版經`prepare_replay.py`生成379個實際正常按鍵，再由`run_adaptive_replay.py`及
`run_adaptive_replay.sh`重播。雙語各380保存狀態與16重載狀態，1584正式／觀察PNG及可比trace相同。
`review_remake.py`核對兩人34物品槽與14持久欄位，雙語保存資料一致。
原版實際保存及重載比較工具為`review_source.py`。擷取期間原版尚未完成；後續結果見下節，不將重製自洽寫成原版parity。

`native-inn3`有553狀態及四次正常逃離，但沒有寫檔。第544狀態Down進入40,23／次格3,0，
實際DOOR／U-TURN列為020f、選中0、等待caller IDA linear 0x117F5。
第545狀態Space保持該提示，第546狀態Down離開提示。進城失敗來自旅行腳本繼續送方向鍵。
`native-inn4`修為觀察到020f時立即送RET；次格0,0只是該批正常走法的目標，不是新進城規則。
前206狀態的角色、seed、鏡頭、次格、時刻、時段及隊伍人數與首次原版來源全部相同。

## 本輪結果與限制

證據等級：已證實，限實際正常寫檔、重新招募及所列持久欄位。

RE-213從正常啟動、招募、購買與修改法術，經自然遇敵逃離回旅館，取得原版CHARDATA實際寫檔。原版574狀態、4次正常逃離；從實際檔正常重載及招募7狀態。Olear法術N／Strength、護符與骰式、金冠601及冒險點數2486和雙語重製保存相符。營養欄原版10／重製8；原版三次CAMPING、重製兩次，每次實際+2。差異保留，不宣稱旅程同狀態。雙語正式／觀察各380保存與16重載狀態，1584PNG及可比trace相同；兩人34物品槽與14持久欄位重載相符。正常旅店招呼後seed1，世界、戰鬥與住宿不重設；原版4次遇敵，重製一次，不宣稱整段骰序、整場或整頁parity。正式325份Go／56bfb4c8不變，保存格式v3保持。原版payload只留本機。22未完成／84完成保持。繁中法術注解A／B、怪物四套資料格式及既有未決事項維持原閘門。入口docs/re/20261003-spell-edit-inn-save.md。

原版CHARDATA SHA-256：`0b83960e316aef6c4f4fc8c31873b2da815cef9b86c4dfca83698bf1499c7c76`。16個768-byte槽位，以實際名稱定位Olear槽2。
原版角色法術為角色記錄+0x200，每槽16 bytes；金冠+0x2A大端24-bit，冒險點數+0x26大端16-bit。
使用既有typed解碼器，欄位位址空間為角色檔內相對偏移；caller均為IDA linear，DGROUP基址2DED0。
工具Go1.26.7、dosgolem968ca774705853276ae55c8d1163c369481ac71f；執行packed KOLTOWN.EXE c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221。
原版SetScratch只管理實際檔案寫入；重載覆層逐份記錄輸入雜湊，沒有角色／座標／骰值注入。

工具續接`review_source.py`、`verify_receipt.py`、`audit.py`及`finalize.py`。正式收據與重驗入口為
[receipt](../verification/spell-edit-inn-save-20261003/receipt.json)及[verify.py](../verification/spell-edit-inn-save-20261003/verify.py)。
