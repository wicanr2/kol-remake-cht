# 原版固定 seed 正常戰鬥路徑追查

日期：2026-09-08。狀態：**進行中，尚未達到戰鬥同狀態驗收**。

## 已完成

`tools/oracle/book_startup_test.go` 增加可選正常路徑探測：
KOL_BOOK_SELECT_DOWN=2、KOL_BOOK_BATTLE=1 選 Olear 進旅店後，固定原版 seed=1，
再用正常滑鼠與確認鍵招募 Theodoric。設定位置沿用 RE-132：
KOLTOWN-unpacked.EXE，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`，
IDA 線性 0x2F4D0／0x2F4D2，DGROUP 基底 0x2DED0，故 DS 相對位移為
0x1600／0x1602。這是既有已證實定位的應用，沒有新增亂數內部 RE。
正常執行實際 DS 為 1EFD；種子設定點是旅店招呼後，與 remake 啟動時 seed=1
不是相同時點，尚不能宣稱兩方戰鬥輸入條件可比。

`workplace/battle-normal-seed-20260908/mouse-recruit/` 的原版測試通過（56.44 秒），
收據來源未變。人工檢視 battle-path-02／03，確認候選名冊及 Theodoric 入隊；
兩名隊員為 Olear／Theodoric，與 remake 已驗正常路徑相同。

## 保留的失敗與訂正

- 根目錄第一輪：TypeKeys 不支援 `<`，測試明確失敗；未當成遊戲缺陷。
  改用原版 (202,182) 招募圖示點擊，再按確認；屬正常滑鼠操作。
- `walk/`：離店後原版先顯示告別訊息，未確認就送方向鍵，路徑沒有按預定出城。
  即使 Go 測試返回 PASS，也只表示探測程式正常結束，**不代表已到戰鬥**。
  已補 Space 確認，在 `walk-confirm/` 重跑。
- remake 的 `leaveDoor` 直接回到地圖；原版告別訊息與確認步驟列為新發現的
  玩家可見差異，尚未完成訊息選擇與各建築分支規格審查，不自行猜補。

## 下一閘門

後續 `walk-confirm/` 已確認離店、出城與首日晚間露營，測試 51.90 秒，來源未變。
`next-day/` 延續同一 seed 狀態，確認露營後走到第二天，畫格
`next-day-029.png` 顯示 `YOU HAVE BEEN AMBUSHED BY 'SLEDGE'.`；測試 74.57 秒，
來源未變。後續方向鍵未關閉伏擊畫面，`next-day-099.png` 仍停在相同問候，
沒有送戰鬥攻擊或強制勝利。`fight-entry/` 另加 Space 進入規劃畫面，人工檢視
`fight-entry.png` 確認目前角色 Theodoric、身體框、戰場及規劃圖示均已出現。
TestNormalBookEntry 通過（91.08 秒），49 個輸出，來源變動旗標 false，擁有權
全部 1000:1000。與 remake 先前 Olear 先手的畫格並非同一比較狀態，不能直接判通過。
本輪所有一次性容器與背景程序均已結束，專案容器清單為空。

確認原版出城、遭遇及戰鬥入口，再記錄隊伍、裝備、傷勢、敵人、地形、站位與
共同 seed 設定點，建立實際可比狀態後才驗命令、射擊／近戰及回寫。
不追逐自然 RND() 的跨全流程精確骰序，不以固定 seed 數字相同代替狀態比較。
