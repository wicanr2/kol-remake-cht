# 原版檔案清冊的外部查證（2026-09-11）

證據等級 `已證實`（外部來源交叉比對）。

## 為什麼查

追「露營之後走不動」時，dosgolem 回報缺檔 `region1.dat`…`region4.dat`。手上的
`Knights-of-legend.zip` 裡只有 `REGION0.DAT`，而且只有 **48 byte**。兩種可能：

1. 原版本來就只有一個 region 檔，遊戲去找其餘四個找不到是正常的；
2. **我們手上這份 ZIP 不完整**——那樣的話所有以它為基準的結論都要重看。

這兩者在專案內部分不開，只能對外查。

## 查到什麼

Internet Archive 的 MS-DOS 版 item（[msdos_Knights_of_Legend_1989](https://archive.org/details/msdos_Knights_of_Legend_1989)，
存取日期 2026-09-11）保存的 `Knights_of_Legend_1989.zip` 共 **76 個檔案**，
其中 REGION 開頭的**只有一個**：

| 檔案 | 大小 |
|---|---:|
| `REGION0.DAT` | 48 byte |

與本專案的輸入**逐項相同**。兩份的差別只有一個檔案：Archive 那份多了
`KOLSOLVE.txt`（32,160 byte，後人整理的攻略文字），不是原版遊戲檔案。

本專案的 `Knights-of-legend.zip`：**75 個檔案**，扣掉那份攻略之後與 Archive 版
一致，包含六個執行檔（`KOL.EXE`／`KOLBOOK.EXE`／`KOLTOWN.EXE`／`TITLE.EXE`／
`MAKEMODE.EXE`／`TANDYPIN.EXE`）、八座城鎮（`TOWN0.0`…`TOWN7.0` 與 `TOWN7.1`）、
對應的 `TWN?PICS.*`、`CHARDATA`、`ITEM.LST`、`MAGIC.0`、六個 `DISKA?` 磁片標記，
以及 `INSTALL.BAT`／`INSUB.BAT`。

## 結論

**我們的原版素材是完整的**，`region1.dat`…`region4.dat` 原版根本不存在。
那四個是**擴充地區**的檔名，由另外發行的地區磁片經原版內建的 REGION INSTALLER
裝入，Origin 沒有發行過——機制與一手證據見
[地區擴充片](20260911-region-installer.md)。遊戲開檔失敗是它自己在清點裝了幾個
地區，**不是**「露營後走不動」的原因——書本階段一樣缺這四個檔而運作正常。

這一條同時排除了「素材不全」這個假設，讓那個死迴圈的診斷範圍縮小到遊戲狀態本身。

## 不宣稱

- 沒有比對逐檔 SHA-256：Archive 那份沒有公開逐檔雜湊，只比了**檔名清單與 REGION
  的數量**。發行版本、修訂與 1998 timestamp 的成因仍是未知（同 RESEARCH-LOG 的既有記載）。
- `KOLSOLVE.txt` 是第三方攻略，權利未知，不引用、不納入專案。
- 本專案不散布原版檔案；這份紀錄只寫清冊與數量。

## 來源

- [Knights of Legend — Internet Archive（MS-DOS）](https://archive.org/details/msdos_Knights_of_Legend_1989)，存取 2026-09-11。
  ZIP 內容清單經 `view_archive.php` 取得。
- [Knights of Legend — Wikipedia](https://en.wikipedia.org/wiki/Knights_of_Legend)，存取 2026-09-11：
  Origin Systems 1989 年於 Apple II 與 C64 發行，MS-DOS 版 1990 年。
