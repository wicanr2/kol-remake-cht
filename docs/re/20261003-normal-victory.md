# RE-196：正常勝利、戰利品與住宿重載

日期：2026-10-03。入口為[目前狀態](../../CONTEXT.md)、
[SPEC-011](../spec/011-wilderness-ambush.md)、[SPEC-027](../spec/027-battle-exit-wounds.md)
及[唯讀重驗](../verification/normal-victory-20261003/verify_final.py)。
研究工作區沿既有 `workplace/normal-victory-20261003/`。先驗既有戰鬥候選，
再修正常住宿暴露的世界時段保存檢查。

## 來源與可比範圍

| 項目 | 身分與限制 |
|---|---|
| 原版輸入 | `KOLTOWN-unpacked.EXE`，SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| 原版來源 | [RE-186](20261002-battle-victory.md)的 `workplace/battle-victory-20261002/source4/complete-source-private.json`；正常147鍵、4回合勝利，沒有直接入口、狀態注入或等待替換 |
| 原版工具 | dosgolem revision `968ca774705853276ae55c8d1163c369481ac71f`；沿用原收據，本批不挑種子或重擲 |
| IDA | IDA Pro 9.4，資料庫 SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`；沿既有證據，未改資料庫 |
| 位址空間 | IDA linear DGROUP `0x2DED0`；世界返回讀值位於 `0x29803/0x2980A`，原版 `DS:BB84` 是另一位址基準 |
| 建置工具 | Go 1.26.7／Ebitengine 2.9.9；`eob-remake-go:1.26.7-ebiten2.9.9-audio`，image SHA-256 `3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa` |
| 勝利比較前版 | `workplace/world-cost-resume-20261003/kol-world-cost-linux`，SHA-256 `f4c4f70b82159b943a555bd855b4a65ca568c5dd25471544df17293032cfe960`；原候選及318份歷史碼輸入保持 |
| 正常戰鬥觀察副本 | 既有 `kol-settlement-probe2-linux`，SHA-256 `24afa7803ea5b900b98adfbc0beed2899e007d53fc07a1e76826a3fe6352496d` |
| 住宿觀察副本 | `kol-victory-inn8-linux`，SHA-256 `2175404dd4d00f289e13d0a6e4531a1c7b02a37ad6dca5b2c4088a26beec5278`；只讀公開狀態再送正常玩家鍵 |

兩側首次規劃入口固定測試 seed 1；原版沿既有dosgolem設定，重製觀察副本以
`KOL_PROBE_PLAN_SEED=1`只設定首次規劃。正式候選沒有永久播種旗標。
每場從標題開始，正常招募、城鎮移動、離城與世界遇敵。四個執行入口比較同狀態；
完整世界進場亂數仍未對拍，不因 seed 數字相同就宣稱整段相同。

## 本場勝利驗收

| 比較點 | 已證實結果 |
|---|---|
| 正常GUI | 英文／繁中各269鍵、270狀態及270張PNG；前120個狀態與PNG保持RE-195正常前綴 |
| 執行入口 | 4回合原版抽樣欄位與疲勞／耐力全部相符，seed依序 `3403800452,2166683160,4200262242,2198482029` |
| 勝利返回 | 末回合勝利、返回暫存1；戰中時刻3360，天候後扣60，世界49,23、次格4,4、時刻3300 |
| 亂數 | 本場86戰中、12個rand19、1個rand100及1個天候rand128，共100；100個seed_before相符，n有兩筆差異，完整骰序不宣稱通過 |
| 天候接續 | spoil chance後seed `1502690714`，最後天候seed `1084083701` |
| 原始物品 | 正常接受戰鎚並選Olear；原版返回後兩人34個完整16-byte槽逐byte相同，包含手持及空槽殘留 |
| 報酬 | Olear／Theodoric金額1835／1067，冒險點數2891／749，傷勢0／0，符合原版人物記錄 |
| 反例 | `normal2-*`取消戰利品，Olear少戰鎚，不能通過本場完整物品驗收 |

`normal3-en/`及`normal3-zh-Hant/`保存成功正常軌跡。第266／267步繁中
戰利品與角色問句已目視，文字完整。PNG只作重製UI驗收，本批沒有宣稱全部原版逐像素相同。

## 明示差異

首次勝利終態只有敵人第6槽 `LastHurt` 差異，原版25、重製23；其他既定抽樣欄位
及計數器相同。原版角色記錄 `+0x2A` 由既有驗證器列為LastHurt；名稱是導航，
原始位移保持。本場倒地敵人不進存檔，這項差異仍開放，成因未知，沒有猜補正式規則。
有限返回／報酬／物品對拍不等於完整戰鬥角色記錄對拍。

完整預驗另發現第4回合兩筆骰面差異。下表索引從0起算，原版呼叫位址依
dosgolem原收據的IDA linear定位；seed與實際n已證實，造成LastHurt差異的關係仍未知。

| 呼叫索引 | seed_before | 原版n／呼叫定位 | 重製n |
|---|---:|---|---:|
| 73 | 1041368008 | 100，IDA linear `0x2736B` | 3 |
| 74 | 441037675 | 2，IDA linear `0x2755E` | 100 |

其餘98筆n與全部100個seed_before相符。這是已知產品骰面／順序差異，不列為原版
oracle未知。本批沒有按終態相同忽略差異，也沒有猜補攻擊規則。最初文件及進度的
「100次亂數相符」在預驗前寫得過廣，本勘誤與兩筆反例收錄於首次失敗及新驗證器。

## 正常住宿與重載

回城途中自然遭遇CLIFF TROLL，以正常REST規劃及逃走返回；Theodoric手持大劍
因此掉落。原版本場勝利的34槽比較在第一次返回完成，住宿保存另與實際住宿前
狀態比較。只有Theodoric第0物品槽變成16個零，其他33槽、戰鎚、金額、傷勢及
冒險點數保持。這段回城只驗重製可玩性，不外推原版全流程。

修正保存檢查後，雙語各507正常鍵／508狀態、27張檢查圖生成v3保存。住宿第一人
Olear付款56後離隊，金額1779／1067、冒險點數2891／749、傷勢0／0。保存16名冊槽，
保留實際住宿前34槽及Ticks3360／Part0。正式候選另起行程正常招募Olear，兩語各
16狀態／PNG與新觀察副本相同；原始物品、typed裝備、背包及完整資料頁還原。
`town.Good`僅省略picture／movement的零值，逐欄比較時只補這兩個零；非零及其他
欄位不放寬。兩語完整保存只排除saved_at後相同，繁中正式重載圖已目視。

## 驗證工具失敗與歷史保存

最初直接翻譯DOS按鍵序列未對齊選單游標及等待階段；第一版語意驅動器把WEAPON
找成35，正確是33。兩版完整失敗軌跡保留。之後正常勝利成功，但先取消戰利品，
另以正常Return三鍵完成領取，不把少物品的結果當通過。

住宿前幾版過早選發言者、以英文問句比較繁中，以及離門後讀舊Inside資料，均屬
私人驗證工具錯誤。修正為Mode／AskingWho後，第6版又把功能ID當畫面圖號：
`IconSprite(id)=id+3`，REST功能5對應圖號8，圖號5是DOOR。第7版已改研究碼，
建置仍引用第6版，兩個二進位SHA相同，未執行第7版。第8版明確核對輸入檔及
圖號後再跑，正式城鎮狀態機不改。第8版走到存檔卻失敗，第9版加唯讀SaveError
並在507步停止，證實錯誤是「世界時段與剩餘時刻不一致」。先前GUI只顯示保存失敗，
不能將這次當成驅動器錯誤。各失敗輸出與中間副本只留本機，沒有加大上限掩蓋卡住。

## 保存檢查的正式修正

RE-151原版正常收據的Left step=13由3420扣至3360，`DS:0x72`仍0，下一步才變1。
來源是 `docs/verification/world-clock-probe-20260927/world-clock-probe.json`，工具與
輸入身分沿上表；已證實正常時段可晚一步。此次回城進門也留下3360／0，保存驗證
卻要求Part立即等於扣時刻後的時段，因而拒絕合法狀態。

[SPEC-023勘誤](../spec/023-complete-save.md)先經證據審查升READY，再將保存檢查
改為 `0 <= Part <= partFor(Ticks)`，保留Ticks／Part及其餘內容原值。沒有讀檔更新
時段或補擲骰，沒有改v3格式。正式只修改 `internal/save/complete.go`，兩份既有
測試增加門檻往返、提前／超界反例及保留下次世界變換的檢查。

`build_save_fix.sh`先將凍結的舊檢查放入隔離副本，確認新回歸拒絕實際舊行為，
再乾淨重跑全部internal／cmd及保存目標測試。新正式候選SHA-256為
`71d66b3150c9dec68c0e3dd16fd3619a42e24d747f9e9861d35c8d513310433c`，
位置 `workplace/normal-victory-20261003/kol-victory-save-linux`；住宿觀察副本
`kol-victory-inn10-linux`為 `4147c6e83677e4fbca8a309a10efc9685d6be3af5fbfdcd6b2e8ebd2c2707620`。
`save-production-inputs.json`鎖定318份碼輸入，只有上述正式檔與兩份測試改變。

開工先封存1367共享輸入，RE-195最終3961附件及下層結果由封存視圖重算。
舊程式、來源及收據保持，不用現況文件覆蓋歷史。Python唯讀重驗使用
`kol-package:20260907`，3 GiB／2 CPU／256 PID及360秒上限：

```sh
test -d "$PWD"
timeout 360s docker run --rm --network none --memory 3g --cpus 2 \
  --pids-limit 256 --user "$(id -u):$(id -g)" \
  -v "$PWD:/src:ro" --entrypoint python3 kol-package:20260907 \
  /src/docs/verification/normal-victory-20261003/verify_final.py
```

原始payload、未授權畫格、觀察程式與存檔僅本機，不加入Git或公開交付。
父規格及第6／44項保持開放；其他戰鬥分支、未知表外返回、完整進場亂數及
上述LastHurt差異仍未完成。本場勝利返回、報酬、物品及新保存鏈只列有限CONFORMED。

批末稽核入口為同一重驗器的 `--audit`，以唯讀根目錄重算收據、歷史結果、318碼
輸入及UID/GID；檢查全專案root-owned與.md同名目錄，並核對主機Git／Docker控制面。
結果保存於同目錄 `audit.json`。工作清單正反測試、核實及掃描通過，22未完成／84完成。

初封重驗器直接比較Python元組與JSON陣列造成稽核失敗，產品內容保持。
九份初封文件已封存；初封receipt.json及verify.py不改。現行verify_final.py先核對
初封附件，再將重算結果轉成JSON標準形比較，另封receipt2.json。原始slot／n／seed
數值沒有換值或忽略欄位；兩筆骰面及LastHurt差異照常明列。

## 2026-10-04：現行進場條件追加

RE-197已修本篇歷史兩筆徒手骰序與LastHurt差異，不重新開案。
現行RE-202接續世界亂數後，舊269鍵在第4回合未勝；首次規劃順序已不同。
成因是正式版仍少逐怪物四套裝備選擇。沿RE-203的當前來源隔離原型
雙語恢復四回合勝利、九個狀態點已映射欄位與報酬相符，詳見
[RE-203追加](20261003-monster-equipment.md#2026-10-04現行四回合續核)。
原版首次seed重設在排序後，本輪不以相同seed掩蓋排序前條件。
資料格式DRAFT、正式來源保持，沒有本輪新住宿／重載，完整戰鬥驗收保持開放。
