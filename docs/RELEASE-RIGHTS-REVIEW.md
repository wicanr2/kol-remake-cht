# 發行前的權利審查

## v.1.0.2-20261013 發行審查

本版接續使用者已核准的HD介面、頭像及四首原創室內樂，製作三平台完整版、引擎包與76秒影片。儲存庫已由主機gh核對維持private，不改公開方式。交付與檢查見[本版收據](verification/release-v1.0.2-20261013.md)。

- 完整版含原作衍生PNG、typed data、字型、核准HD與OGG，僅本機保存。
- 引擎包沿既有範圍：程式、工具、必要翻譯、RRSAL-1.0及匯入雜湊清單。不得包含assets、原版資料、PNG／OGG或完整手冊譯文；自製幾何app圖示例外另核。
- 四首配樂為專案原創MIDI，以FluidR3 GM音色庫渲染；RRSAL-1.0及MIT來源見assets/audio/hd/manifest.json。來源、接受紀錄與OGG原檔保持，沒有原版音樂取樣。原版目前僅證實標題雷聲，不把未找到背景旋律寫成不存在。
- 影片使用本版實際封包新畫格與已核准配樂／雷聲，含原作衍生畫面，僅本機保存。
- 新版外部發布只限wicanr2/kol-remake-cht的三個引擎包與SHA256SUMS.txt；舊tag／Release保持。使用者先選擇維持一色階並暫緩發布，修正通過後在明確的最後確認以go解除暫緩。四個附件的精確雜湊已登記，發布與遠端核對中。

本次審查只適用上述內容與目前private儲存庫，不授權公開素材、完整版或影片。

## v.1.0.1-20261012 發行審查

2026-10-12 使用者要求三平台完整版、release 與推廣影片。GitHub 已核對仍為 private，
本次不改可見度。新交付與完整驗收見[本版收據](verification/release-v1.0.1-20261012.md)。
同日使用者明確同意指定儲存庫的上傳與發布，已完成 private GitHub Release，
附件僅三個引擎包及 `SHA256SUMS.txt`；遠端完整性與 tag 已核對。

| 內容 | 本版處理 |
|---|---|
| 完整版 | 含原作衍生 PNG／typed data、字型、OGG 與首批 HD，僅本機，不上傳 |
| Release 引擎包 | 只有 remake 程式、工具、條款、執行所需翻譯及資料雜湊清單；無素材位元組 |
| 完整手冊譯文 | 只放本機完整版；不把舊待決事項解讀成散布授權 |
| 推廣影片 | 本版實際畫格，含原作衍生畫面，僅本機，不上傳 |
| GitHub Release | 僅三平台引擎包與其 SHA-256，維持 private；不宣稱具備公開散布所有內容的權利 |

HD 素材 manifest 明示「private GitHub 可保存；無公開散布權」，故沿原作素材處理。
Release 解包後必須沒有 `assets/`、PNG、OGG、PDF、原版 EXE 或完整手冊譯文；
只允許自製幾何 app 圖示。引擎本身的 `kol.exe` 是本專案編譯產物。
玩家須自行匯入合法持有、依 manifest 全部 SHA-256 相符的資料，才可遊玩或切 HD。
程式採既有 RRSAL-1.0；原版素材及翻譯的權利不因此改變。

以下為早期審查紀錄，不作本版封包內容或 GitHub 狀態的依據。

專案契約規定：**儲存庫轉 public、建立 public fork、發 GitHub Release，或公開
`dist-all/.../release/` 之前，必須先停下來重做一次權利審查**（`CLAUDE.md`）。
這一份就是那次審查的紀錄。審查對象是 `dist-all/<版本>/` 的三個包。

審查日期：2026-09-07　　審查對象：`dist-all/0.2.0/`

## 三個包的界線

| 包 | 內容 | 權利分類 | 可不可以公開 |
|---|---|---|---|
| `full/` | 執行檔 ＋ **原作衍生的 PNG／JSON 素材** ＋ 文件 | 原作衍生 | **不可以**。包裡有 `DO-NOT-DISTRIBUTE.txt` |
| `release/` | 執行檔 ＋ 文件 ＋ 抽取器，**一個原版衍生位元組都沒有** | remake 自有 | 可以（見下面的逐項） |
| `promo/` | 推廣片 ＋ 來源 metadata ＋ 條款 | 見「推廣片」一節 | 有條件 |

`release/` 的作法是：玩家拿**自己的**原版磁片跑包裡的抽取器產生素材。
這是老遊戲重製的慣例，也是這個專案從一開始就選的路。

## `release/` 逐項

| 項目 | 來源 | 判定 |
|---|---|---|
| `kol`／`kol.exe`／`kol-macos` | 本專案 Go 原始碼 | 自有 |
| `LICENSE` | RRSAL-1.0，著作權人 Wang Chun-Yu | 自有 |
| `README.md` | 本專案 | 自有 |
| `docs/manual/` | 說明書的**繁體中文譯文**（Markdown），不含原始 PDF、不含插圖 | 衍生著作，見下面的「說明書譯文」 |
| `docs/i18n/` | 原版文字的繁體中文譯文，以**記錄號**索引，不含原文 | 同上 |
| `docs/re/*.json` | remake 自己的 typed data（標題序列、城鎮取樣） | 自有 |
| `tools/*.py` | 抽取器與工具 | 自有 |

機器檢查：`build_dist.sh` 第 7 節與 `verify_dist.sh` 都跑一次 deny-list，
`release/` 裡不得出現 `*.png`／`*.ogg`／`*.EXE`／`*.zip`／`*.pdf`／`*.jpg`。
上一次建置通過。

## 說明書譯文（要使用者決定的那一項）

`docs/manual/` 是 1989 年那本說明書的逐頁繁體中文翻譯。

- **不含**原始 PDF、不含掃描圖、不含插圖。
- 但翻譯是**原文的衍生著作**：原文的著作權仍在原權利人手上，翻譯權也是。
  「只放譯文不放原文」降低了風險，**不等於取得授權**。

同樣的形狀在華人老遊戲圈很常見（攻略、說明書翻譯長年流通），實務上少有爭議，
但那是實務觀察不是法律結論。

**這一項要使用者決定**：`release/` 要不要帶 `docs/manual/`。
兩個選項都做得到——不帶的話 `build_dist.sh` 改一行，包裡放一個指向線上文件的連結。

## 推廣片

- **畫面**：全部是 remake 自己的實機截圖（`docs/verification/` 的可重播收據），
  沒有原版畫面、沒有 mockup。remake 的畫面是原版畫面的重製，形狀上與
  `full/` 的素材同一個問題——但**這些 PNG 是 remake 執行時畫出來的**，
  不是原版檔案的轉換產物。
- **音訊**：沒有背景音樂。全庫普查過，KOL 的 DOS 版只有 TITLE 的 PC Speaker
  音爆，**原版沒有可用的配樂**。片中兩下雷聲是
  `assets/audio/title-thunder.ogg`，依公開的 PIT channel 2 規格產生，
  manifest 分類為「remake 自有；可公開」——不是原版音訊的取樣。
  所以**沒有散布他人著作權音樂的問題**（`rulebook/93` 的但書不成立）。
- **商標**：片中出現 `Knights of Legend` 與 `Origin Systems` 兩個名稱，
  用途是指明本作重製的對象（指示性合理使用的典型形狀）。片頭沒有使用原版
  logo 字體或盒繪。

## 結論與待決

1. `release/` 的執行檔、文件、抽取器與譯文之外的部分：**可以公開**。
2. `full/`：**不可以公開**，任何情況下都不進 Release、不進 public 分支。
3. `docs/manual/` 譯文要不要進 `release/`：**待使用者決定**。
4. 儲存庫目前是 **private**，尚無任何 Release。建立第一個 Release 是
   **對外、不可逆**的動作，依專案契約與代理邊界，要使用者明確同意才做。

## 這一份要跟著版本走

每次要發行都重跑一次：包的內容變了、抽取器帶的檔案變了、promo 換了素材，
上一次的結論就不自動成立。

## 儲存庫可見度閘門（2026-09-30）

`tools/build_dist.sh`、`tools/package_dist.sh` 與 `tools/verify_dist.sh` 在主機端開頭都先跑
`tools/check_repo_visibility.sh`，用 `gh api repos/<owner>/<name> --jq .visibility` 問 GitHub 目前的可見度：

| 結果 | 動作 |
|---|---|
| `private` | 放行 |
| 其他值（`public`、`internal`）、空輸出 | 停下，結束碼 4，訊息指向本文件 |
| 沒有 `gh`、沒登入、沒網路、API 出錯 | 停下，結束碼 4；查不到就不放行 |

沒有略過旗標。儲存庫要轉 public，要先重做本文件的權利審查、修改這支腳本的判定並留下紀錄。
回歸測試是 `tools/test_check_repo_visibility.sh`（假的 `gh` 走完五種輸出與沒有 `gh` 的情形）。
`release/` 的素材 deny-list 原本就是失敗即關閉（`verify_dist.sh` 找到 PNG、OGG、EXE、ZIP、PDF、JPG 就退出）。

這道閘門只擋自家腳本；在 GitHub 網頁或 `gh` 直接改可見度、建立 Release，腳本管不到。
發 GitHub Release 的流程（#43）建立時要先呼叫同一支腳本。
