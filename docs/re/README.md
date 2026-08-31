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

| 主題 | 文件 | 目前狀態 |
|---|---|---|
| 固定輸入與衍生映像 | [builds.md](builds.md) | KOLTOWN packed／unpacked 已固定 |
| 位址空間與搬移映射 | [address-spaces.md](address-spaces.md) | stage 1／stage 2／unpacked 已建立 |
| Packer／compiler／runtime | [toolchain.md](toolchain.md) | KOLTOWN 為 Microsoft EXEPACK 已證實；精確工具版本未知 |
| 全部研究時間線 | [RESEARCH-LOG.md](RESEARCH-LOG.md) | RE-001 至 RE-032 |
| 圖片容器與壓縮 | [image-formats.md](image-formats.md) | COMBPICS、KOLPICS、KOLMPICS、TWN 全包、TITLE SET／FACE／BACK、ENDPICS、BOOKPICS、WEATPICS、AWD1PICS、SHLDPICS、HANDPICS、WEAPPICS、HEADPICS／TORPICS／LEGSPICS、MONPICS、PERPICS、MISCPICS 與固定記錄表已分類 |
| 音訊 cue 母體 | [audio-inventory.json](audio-inventory.json) | 六個 executable 直接 I/O 母體已關閉；只有 TITLE PC Speaker cue |
| TITLE 音訊 oracle | [title-audio-oracle.json](title-audio-oracle.json) | DOSBox-X 正常流程非靜音擷取已證實；人耳驗收待辦 |
| IDA／runtime 安全摘要 | `ida-fingerprints/*.json` | build、startup 與逐 byte oracle 摘要；含 bytes 的 `*-private.json` ignored |

## 下一個知識缺口

SPEC-002 與 SPEC-004 仍為 DRAFT。圖片 theme 範圍等待使用者決定；音訊直接 I/O 母體與
正常流程機械擷取已關閉，下一步是人耳驗收及 TITLE cue 的 OGG 處理決策。兩份規格未達
READY 前不得接進 runtime。
