# 《武士傳說》逆向知識庫

這裡是 remake 的原版證據入口。知識庫先於 spec，spec 先於實作；任何玩家可見行為若
沒有本知識庫的原始證據定位，不得直接進 production。

## 查詢順序

1. 先在 [builds.md](builds.md) 確認輸入 SHA-256 與衍生映像身分。
2. 再依 [address-spaces.md](address-spaces.md) 選對位址空間；同一數值若基準不同，不得
   視為同一位址。
3. 從 [toolchain.md](toolchain.md) 排除 packer、compiler runtime、driver 與工具程式。
4. 到 [RESEARCH-LOG.md](RESEARCH-LOG.md) 查單一問題的證據、推論等級與勘誤。
5. 只有已審查的結論才可進 `docs/spec/`；未達 `READY` 不得實作。

## 固定顯示契約

每筆函式、位址或 operand 證據至少同列：

- 輸入檔名與 SHA-256；
- 工具版本與位址空間；
- 原始名稱／位址／operand／bytes 的本機證據位置；
- 附加語意與 `已證實／強推論／假說／未知`；
- 證據來源、writer／reader／consumer 與玩家影響；
- 未達已證實時的醒目限制。

含原版 executable bytes、完整 disassembly、`.i64` 或解包 EXE 的檔案只留本機 ignored
區；Git 只保存 hash、位址、統計、工具與結論摘要。

## 索引

### 主題文件

先看主題文件，那是整理過的機制與導覽；要逐條證據、推論等級與勘誤才進 RESEARCH-LOG。

| 主題 | 文件 | 涵蓋 |
|---|---|---|
| 固定輸入與衍生映像 | [builds.md](builds.md) | KOLTOWN packed／unpacked 已固定 |
| 位址空間與搬移映射 | [address-spaces.md](address-spaces.md) | stage 1／stage 2／unpacked |
| Packer／compiler／runtime | [toolchain.md](toolchain.md) | KOLTOWN 為 Microsoft EXEPACK 已證實；精確工具版本未知 |
| 圖片容器與壓縮 | [image-formats.md](image-formats.md) | 30 個邏輯 archive、1,022 個 payload 全數有 disposition |
| 文字與字型 | [text-and-fonts.md](text-and-fonts.md) | 字模、三個 UI 文字庫、城鎮的 5-bit 打包、代換碼、中文雙模式 |
| 角色建立 | [character-creation.md](character-creation.md) | 屬性 RNG、職業表、起始裝備、writer coverage、UI 邊界 |
| 戰鬥 | [combat.md](combat.md) | 各階段的函式與 RE 對照、六個容易踩的地方、還沒解的六項 |
| 音訊 | [audio.md](audio.md) | 整份原版只有一個音效；母體怎麼窮舉、remake 為什麼用合成而不是錄音 |
| 城鎮與世界 | [town-and-world.md](town-and-world.md) | `TOWNn.0` 的 payload 版面、地圖、門、建築內對話、世界地圖、道路、遭遇 |

### 機器可讀輸出

| 內容 | 檔案 |
|---|---|
| 全部研究時間線（RE-001…RE-090） | [RESEARCH-LOG.md](RESEARCH-LOG.md) |
| 音訊 cue 母體 | [audio-inventory.json](audio-inventory.json) |
| TITLE 音訊 oracle | [title-audio-oracle.json](title-audio-oracle.json) |
| 書本文字版面 | [text-layout.json](text-layout.json) |
| 城鎮的門與地圖 | [town-doors.json](town-doors.json)、[town-map-grid.json](town-map-grid.json) |
| IDA／runtime 安全摘要 | `ida-fingerprints/*.json`（含 bytes 的 `*-private.json` ignored） |
| 各 archive 的視覺探針 | `*-ega-visual-probe.json` |

## 還沒有主題文件的區塊

這幾塊的證據都在 RESEARCH-LOG，但還沒整理成主題文件：

- **角色記錄的欄位版面**（RE-034…RE-051、RE-074、RE-075）：768-byte 的
  `CHARDATA` 記錄裡哪一格是什麼。

## 目前的知識缺口

按「會不會擋住玩家路徑」排：

1. **音訊的人耳驗收**——SPEC-004 卡在這一關，過了才接得上 Ebitengine
   （見 [audio.md](audio.md)）。
2. **馬廄的商品表**——買不到坐騎，所以「全隊都有坐騎才問得出來」的那一頁
   （騎走，RE-112）在正常玩法上到不了。
3. **戰利品畫面沒有原版畫格**——版面已經靜態讀完（RE-115），但要拍到它得打贏
   一場而且擲中兩成的掉落，所以那一塊到現在只有靜態證據。
4. **敵人 AI 的最後缺口**——移動段繞不出去時的重試（`word_34A08`）與篩選遮罩要的
   參戰者記錄 `+0x05`。空手去撿那一條已經做完（RE-116），不過它在原版資料下
   走不到：三十二隻怪全部都有武器。
5. **`EDIT PICTURE`／`EDIT SHIELD`**——原版是滑鼠驅動，現有的鍵盤 oracle 走不完，
   要先讓 oracle 能送滑鼠事件。

規格狀態見 [../spec/README.md](../spec/README.md)；未達 `READY` 不得實作。
