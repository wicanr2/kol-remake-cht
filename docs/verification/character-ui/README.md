# 角色建立 UI 的正常 DOS 玩家路徑收據

更新日期：2026-09-01

本目錄保存 KOLBOOK 角色建立流程在正常玩家路徑下的動態 oracle 收據。所有結論來自
`KOL.EXE` 啟動、鍵盤操作、逐畫面擷取與三個角色檔的前後 SHA-256，不使用除錯入口、
座標注入或測試旗標。

## 產生方式

```
tools/run_character_ui_oracle.sh <輸出目錄> --scenario <情境>
tools/compare_character_ui_runs.py <第一輪目錄> <第二輪目錄> --out cross-run/<情境>.json
```

- 容器映像 `rich2-dosbox-dbg:latest`（DOSBox-X 2026.06.02 SDL2），`docker run --rm
  --network none`，`--memory 1g --cpus 1 --pids-limit 256`。
- 原版 ZIP 唯讀掛載，SHA-256 不符即失敗關閉；遊戲目錄每次由 ZIP 重建於容器 tmpfs。
- 機器設定固定為 `machine=ega`、`memsize=4`、`core=normal`、`cputype=386`、
  `cycles=fixed 3000`，設定檔本身的 SHA-256 記在每份收據內。
- 按鍵以 X11 XTEST 送出（`tools/x11_key_inject.py`）。送鍵時機不是盲打時序，而是
  等畫面靜止到指定秒數才送下一鍵，因此原版磁碟解碼時間的變動不會讓序列走偏。
- 畫面由 Xvfb 的 `-fbdir` framebuffer 取樣，依 DOSBox-X 視窗實際幾何裁切，逐畫格
  去重後存成 PNG。

## 畫面 PNG 的保存位置

收據記錄每個畫格的 PNG SHA-256、尺寸與時間軸；PNG 本身目前只留本機，未進版控。
原版執行畫面的截圖是與 `assets/original-derived-png/` 不同的素材類別，2026-09-01 的
private repo PNG 特例沒有涵蓋它，需另行權利決策後才可入庫。

## 情境與結果

| 情境 | 檢查點 | 兩輪相同 | 三檔是否變動 |
|---|---:|---:|---|
| `name-length` | 26 | 26 | 否 |
| `blank-name` | 10 | 8 | 否 |
| `escape-layers` | 28 | 22 | 否 |
| `reroll` | 20 | 13 | 否 |

兩輪之間的差異全部落在三類本來就不可能逐 byte 相同的畫面，差異量已量化：

| 差異來源 | 檢查點 | 不同 byte 比例 |
|---|---|---|
| TITLE 標題序列（循環動畫，固定等待落點不同） | `blank-name` 的第一個檢查點 | 13.98% |
| 名字欄閃爍游標 | 各情境的名字層畫面 | 0.0125% |
| 屬性頁數值（BIOS tick 播種的擲骰） | 屬性頁及其後仍顯示數值的畫面 | 0.175%–0.280% |

其餘檢查點在兩次全新執行之間逐 byte 相同。

## 已證實的行為

- **名字欄容量 16 個字元，且輸入第 16 個字元會直接前進到種族層。** 連續送出
  `ABCDEFGHIJKLMNOPQ` 時，畫面在第 15 個字元顯示 `abcdefghijklmno`，第 16 個字元
  顯示 `abcdefghijklmnop` 並同時開啟種族選單；第 17 個字元與其後兩次 Backspace
  都落在種族層、對畫面沒有作用。26 個檢查點兩輪逐 byte 相同。
- **空名字不會前進。** 在名字欄不輸入任何字元直接按 Return，畫面停在名字層。
  這與 KOLBOOK 靜態控制流「輸入 buffer 首 byte 為零時回到同一輸入點」一致。
- **各層取消的返回目的地**：種族→名字、性別→種族、職業→性別、屬性→職業，
  再往上依序回到性別、種族、名字層。回到名字層時，名字與右頁摘要欄位都已清空。
- **屬性頁的 REROLL STATS 會重新產生整組數值。** 同一個 HUMAN／MALE／BARBARIAN
  角色連續兩次 reroll，力量 88 → 85 → 89、體型 74 → 72 → 73、健康 75 → 75 → 74，
  九個顯示值均重擲。
- 上述四個情境全程沒有接受屬性或角色圖像，`CHARDATA`、`CHARPICS`、`SHLDPICS`
  的前後 SHA-256 皆未變動。

## 尚未涵蓋

- 接受屬性、選擇角色圖像與實際保存的寫入邊界。
- 16 槽全滿時的訊息與退出行為。固定 ZIP 的 `CHARDATA` 十六槽都不是全零，前十四槽
  有名字，第 15、16 槽名字為空但記錄非零；槽位是否可用的判準尚未由 UI 確認。
- 名字欄接受的字元集合、大小寫如何保存，以及畫面字型顯示為小寫是字型或是資料。
