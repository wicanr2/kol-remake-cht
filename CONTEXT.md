# 《武士傳說》重製目前狀態

更新日期：2026-08-31

## 目前真相

- 本專案是 *Knights of Legend* 的繁體中文乾淨重寫，正式 runtime 使用 Go／Ebitengine、
  PNG 圖像與 Ogg Vorbis 音訊。
- GitHub `wicanr2/kol-remake-cht` 已建立為 private；原版與衍生素材不得提交。
- remake 尚未開始。現階段是 M1「鎖定原版與研究來源」，不能宣稱可啟動或可玩。
- `Knights-of-legend.zip` 已在唯讀容器中通過 ZIP CRC 檢查，共 76 個中央目錄項目；可見
  6 個 MZ executable、安裝批次檔、城鎮／角色／物品資料及多組 `*PICS` 圖像候選。
- ZIP 多數遊戲檔時間為 1989，但 `KOLTOWN.EXE`、`CHARDATA`、`CHARPICS`、`SHLDPICS`、
  `STEEDPIC` 等顯示 1998。這份封裝是可運行 DOS 集合的 `強推論`，是否為未修改原始
  發行版仍是 `未知`。
- IDA Pro 9.4 `locked-v1` 已對六個 executable 完成第一層指紋。全部由 IDA loader 判定
  為 16-bit MS-DOS EXE；`KOL`、`KOLBOOK`、`MAKEMODE`、`TITLE` 出現同一個 Microsoft
  Quick C／MSC 5.1 runtime FLIRT 候選，這只證實可套用的 library signature 候選，不能
  鎖定精確 compiler 版本。`TANDYPIN` 尚未辨識函式，仍屬未知。
- `KOLTOWN` entry 已證實會把 `0x131f` bytes 由高位址向低位址複製到另一 segment，隨後
  `retf` 至新 segment 的 `IP=0x0038`；這是自我重定位／解包前置 stub 的強推論。檔案相對
  MZ 宣告長度多出的 81 bytes 已證實全為 `0x1A`，屬 DOS EOF padding，不是 overlay payload。
- 魔旅士 2010-04-16 的玩家心得與《CRPG 通鑑》中文版已進入 README；只作專案起源，
  不作玩法 oracle。

## 固定輸入

| 檔案 | Bytes | SHA-256 | 分類 |
|---|---:|---|---|
| `Knights-of-legend.zip` | 1,699,452 | `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46` | 僅本機原版封裝 |
| `knightslegend-manual.pdf` | 35,942,482 | `4c80f5d4c7221f9d3376546d249d52e2fc7fc0f73b48ee9daf6e20ce87d7086a` | 僅本機英文說明書 |
| `knightslegend-refcard.pdf` | 2,253,388 | `4815bc6c3a51ff7abaa610a071f08e05eb2dd180e5130b0674de8b4a94b2abd3` | 僅本機參考卡 |
| `KOL-MAP-001a4s2.jpg` | 92,661 | `9528ca50e14d09d746d2ecce523ac56f78baeaec5cdd4ba5e3351679bd52dff9` | 僅本機參考地圖 |
| `knightslegend.jpg` | 433,061 | `76b08e4b6a6da9b616d7f57ce3e77c2973591379c8897779d849126555defb9b` | 僅本機盒面 |
| `knightslegend-back.jpg` | 525,582 | `10d3e63358d6604b807d4afae3b53acf9a2962ce10871390dd815e325e53b194` | 僅本機盒背 |

## 文件入口

- 執行規範：`AGENTS.md`
- 唯一待辦：`WORKLIST.md`
- 工作歷程：`WORKLOG.md`
- 研究證據：`docs/re/RESEARCH-LOG.md`
- 重製計畫：`REMAKE-PLAN.md`
- 驗證矩陣：`docs/verification/VERIFICATION-MATRIX.md`
- 交接：`HANDOFF.md`

## 下一個唯一工作

在 IDA 一次性 DB 中沿 `KOLTOWN` 的 far-transfer 目標建立第二階段 code，辨識 relocation／
解壓完成後的真正入口與 packer 家族；不得把目前 32 條線性解碼當成完整控制流。
