# SPEC-001：原版 Oracle 準備與位址契約

- 狀態：`DRAFT`
- 日期：2026-08-31
- 範圍：研究工具與證據輸入；不定義任何玩家玩法
- RE 來源：`docs/re/builds.md`、`docs/re/address-spaces.md`、`docs/re/toolchain.md`

## 目的

建立唯一、可重生、不可誤用 packed 位址的 `KOLTOWN` 研究 oracle。這份 spec 未達
`READY` 前，不得據此實作遊戲規則、資料 parser 或 runtime 行為。

## 輸入

1. 原始 ZIP SHA-256 必須是 `2b455e3c…9f46`。
2. ZIP 內 `KOLTOWN.EXE` SHA-256 必須是 `c2442c84…b221`。
3. 解包器來源 SHA-256 必須是 `0e1da11a…b03c`，並保留其 CC0 檔頭與 garbage-byte 限制。
4. 工具必須在無網路、非 root、資源受限 Docker 容器執行；原始輸入唯讀。

任何 hash 不同都必須失敗即關閉，建立新的 build 身分，不得沿用本 spec 位址。

## 產物

- 本機 ignored `KOLTOWN-unpacked.EXE`，SHA-256 `01f6f9a9…d1d2`。
- 本機 ignored IDA Pro 9.4 derivative `.i64`。
- 受版控的 build、位址空間、toolchain 摘要與不含原版 bytes 的 JSON。
- packed 與 unpacked 位址必須帶空間名稱；禁止裸位址跨空間引用。

## 驗收 gate

- [x] 原始 ZIP CRC 與逐檔 SHA-256 通過。
- [x] EXEPACK marker、解壓命令、relocation loop 與 final transfer 有 IDA 原始位址證據。
- [x] 解包輸出 header 為 `1b8f:001c`、`2b48:07d0`、2,281 relocations。
- [x] IDA 9.4 重載入口為 unpacked EA `0x2b90c`。
- [x] 強制建立入口函式後可辨識 337 個函式。
- [ ] 以第二個獨立方法驗證至少三個 unpacked file offset 對回 packed／runtime bytes。
- [ ] 審查解包器 garbage-byte 差異不影響 code、relocation 或玩家資料 consumer。

## READY 條件

上列兩個未完成 gate 關閉，且 RE 知識庫與 generated summary 可由乾淨輸入重生。READY
只授權「把此 derivative 當研究 oracle」，不授權任何玩法實作。

## 實作狀態

無。依使用者要求，RE 知識庫與 spec 審查完成前不進入 remake implementation。
