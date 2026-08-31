# 《武士傳說》重製計畫

## 法律與完整性邊界

- 原版檔案由使用者提供、不散布，所有分析都唯讀。
- 存檔、抽取與改造資料使用明確可寫 overlay／輸出目錄。
- 乾淨重寫使用 Go／Ebitengine；runtime 只讀 PNG 與 OGG。
- private GitHub 也不提交原版或未獲授權的衍生素材。

## 清冊

| 類別 | 檔案／版本 | 已知格式 | 證據 | 狀態 |
|---|---|---|---|---|
| executable | ZIP 內 6 個 `.EXE` | DOS MZ | header magic 與逐檔 SHA-256 | 已盤點；工具鏈未知 |
| 城鎮／事件 | `TOWN0.0`…`TOWN7.1`、`KOLDATA.0`、`BOOKDATA.0` 等 | 未知 | ZIP manifest | 待解 |
| 角色／存檔候選 | `CHARDATA`、`CHARPICS`、`SHLDPICS`、`STEEDPIC` | 未知 | 1998 timestamp，與多數 1989 檔不同 | 身分未知 |
| 圖像／字型候選 | `*PICS*`、`AWD1HERC.0`、`AWD1PCGA.0` | 未知 legacy raster | 檔名與 ZIP manifest | 僅是假說，待 render／consumer 證實 |
| 音訊 | 尚未辨識 | 未知 | 尚無證據 | 待盤點 driver／資料 |
| 說明文件 | 79 頁 manual、5 頁 reference card | PDF 1.6 掃描件 | Poppler 22.12.0 metadata | 已固定雜湊，待翻譯 |

## 架構

```text
cmd／平台／UI → Knights of Legend 適配器／規則 → typed data
              → 繪圖／輸入／音訊／儲存／亂數原語
```

抽取工具與 runtime 完全分離；抽取器讀固定雜湊原版並產生 manifest＋PNG／OGG，runtime
不得連結或開啟原版 archive、EXE 或專有資料格式。

## 垂直切片

| 切片 | 原版 oracle | Parser／規則／UI／存檔路徑 | 驗證 | 狀態 |
|---|---|---|---|---|
| 啟動至第一個可操作畫面 | 待建立 DOS oracle | 未開始 | 同狀態截圖＋正常輸入 | 待辦 |
| 城鎮一項互動 | 待鎖定 | 未開始 | 狀態／畫面／存讀檔 | 待辦 |
| 一場完整戰鬥 | 待鎖定 | 未開始 | 規劃、疲勞、傷害、逃跑 | 待辦 |

## 工作項

以 `WORKLIST.md` 為唯一狀態來源，本檔只保存架構與里程碑，不複製逐輪進度。

## 刻意差異

目前無。戰鬥加速、簡化確認、現代化 UI 或其他玩家體驗差異都尚未經使用者決策。
