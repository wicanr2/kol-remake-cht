# 畫面一覽

remake 目前所有的畫面，左邊是**英文模式**、右邊是**繁體中文模式**。

兩個模式是同一個狀態機，只有畫法不同：

- **英文模式** 320×200，用原版字模，逐畫面與原版對拍。
- **中文模式** 960×600，16×18 中文字模格，**不對拍**——原版沒有中文畫面。
  倍率三是算出來的：原版最小行距 6 px，`6 × 3` 剛好等於字模格高（[SPEC-030](../spec/030-cjk-character-sheet.md)）。

表格裡的「—」表示那一側沒有截圖，通常是因為那張畫面在兩個模式下沒有差別，
或者那個情境只在其中一個模式跑過。

每一張都由 [`shots.spec`](shots.spec) 的一行指令重生，不是手動截的；
產生流程與驗證判準見 [README](README.md)。

## 標題序列

遊戲開場。這兩張是原版美術，沒有文字要換。

| 畫面 | 英文 | 繁體中文 |
|---|---|---|
| 暴風雨夜裡的騎士 | ![暴風雨夜裡的騎士](en-title-knight.png) | — |
| 戰士逼近 | ![戰士逼近](en-title.png) | — |

## 書本

原版的主選單長得像一本書：左頁是內容、右頁是清單。封面與目錄頁上的字是美術不是文字，兩個模式看到的是同一張圖，只有解析度不同。

| 畫面 | 英文 | 繁體中文 |
|---|---|---|
| 封面 | ![封面](en-book-cover.png) | ![封面](zh-book-cover.png) |
| 目錄頁 | ![目錄頁](en-book-contents.png) | ![目錄頁](zh-book-contents.png) |
| 角色選項 | ![角色選項](en-book-menu.png) | ![角色選項](zh-book-menu.png) |
| 角色清單 | ![角色清單](en-book-roster.png) | ![角色清單](zh-book-roster.png) |
| 刪除確認 | ![刪除確認](en-book-delete.png) | ![刪除確認](zh-book-delete.png) |
| HARBOR MASTER 訊息 | ![HARBOR MASTER 訊息](en-book-harbor.png) | ![HARBOR MASTER 訊息](zh-book-harbor.png) |

## 建立角色

擲屬性、挑種族職業、選圖像。

| 畫面 | 英文 | 繁體中文 |
|---|---|---|
| 擲屬性 | ![擲屬性](en-newchar-stats.png) | ![擲屬性](zh-newchar-stats.png) |
| 選角色圖像 | ![選角色圖像](en-newchar-picture.png) | ![選角色圖像](zh-newchar-picture.png) |

## 城鎮

走在城裡、進門、跟人說話。

| 畫面 | 英文 | 繁體中文 |
|---|---|---|
| 戶外地圖 | ![戶外地圖](en-town-outdoor.png) | ![戶外地圖](zh-town-outdoor.png) |
| 走了幾步 | ![走了幾步](en-town-outdoor-moved.png) | — |
| 走到地圖邊緣 | ![走到地圖邊緣](en-town-edge.png) | — |
| 站在門口 | ![站在門口](en-town-door.png) | ![站在門口](zh-town-door.png) |
| 建築內部 | ![建築內部](en-town-building.png) | ![建築內部](zh-town-building.png) |
| 鐵匠鋪 | ![鐵匠鋪](en-town-building-smithy.png) | ![鐵匠鋪](zh-town-building-smithy.png) |
| 旅店 | ![旅店](en-town-inn.png) | ![旅店](zh-town-inn.png) |
| 旅店（兩人隊伍） | ![旅店（兩人隊伍）](en-town-inn-party2.png) | ![旅店（兩人隊伍）](zh-town-inn-party2.png) |
| Brettle 的旅店 | — | ![Brettle 的旅店](zh-town-brettle-inn.png) |
| Olanth 的麵包店 | — | ![Olanth 的麵包店](zh-town-olanth-bakery.png) |
| LISTEN：問話 | — | ![LISTEN：問話](zh-town-listen-ask.png) |
| LISTEN：回答 | — | ![LISTEN：回答](zh-town-listen-answer.png) |

## 角色資料頁

旅店的 MIRROR：八個屬性、武器熟練度、金幣與傷勢／營養兩條長條。物品與隊伍立繪是**換頁**不是蓋在資料頁上（RE-126／RE-127）。

| 畫面 | 英文 | 繁體中文 |
|---|---|---|
| 角色資料頁 | ![角色資料頁](en-town-sheet.png) | ![角色資料頁](zh-town-sheet.png) |
| 物品（換頁） | ![物品](en-town-items.png) | ![物品](zh-town-items.png) |
| 隊伍立繪（換頁） | ![隊伍立繪](en-town-party-page.png) | ![隊伍立繪](zh-town-party-page.png) |

## 野外

出了城門之後的世界地圖。

| 畫面 | 英文 | 繁體中文 |
|---|---|---|
| 世界地圖 | ![世界地圖](en-world-map.png) | ![世界地圖](zh-world-map.png) |
| 走到城鎮入口 | ![走到城鎮入口](en-world-town.png) | ![走到城鎮入口](zh-world-town.png) |
