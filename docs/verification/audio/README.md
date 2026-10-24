# 音訊驗收

更新日期：2026-10-05

原版 TITLE 的可聽事件只有一種：`sub_10F75` 的喇叭雷聲。固定 ZIP 沒有獨立的 WAV、VOC、
MIDI 或已辨識的音訊 archive，六個 executable 中也只有 `TITLE.EXE` 碰到埠 `0x61`／`0x42`
（SPEC-004 已固定的證據）。

使用者於 2026-09-01 決定這個 cue 採**規格近似合成**：波形由公開的 PC Speaker／PIT
channel 2 規格與已證實的演算法產生，不散布原版錄音。因此 `assets/audio/title-thunder.ogg`
裡沒有任何原版音訊資料，權利分類是 remake 自有。

## 合成依據

`sub_10F75` 的演算法已由兩個全新 IDA DB 證實（RE-047 與 `TITLE-speaker-summary.json`）：
兩段音爆，每段開喇叭後更新 100 次 PIT channel 2 除數再關喇叭；每一步的除數是
`500 × 段次 + rand()%10000 + 1 + 1000`，方波頻率為 `1193182 / 除數`；兩段之間有 50%
機率多等一個單位。

**長度不能靜態讀出**：步進間隔是用開機期校正值算出的忙等，而校正迴圈（`sub_118D1`，
數的是「一個 BIOS tick 內跑得完幾圈」，迴圈裡有 `int 1Ah`）與延遲迴圈的指令不同，
換一台機器就換一個長度。因此總長度取自 DOSBox-X 擷取的量測值，並標為
`hardware-spec approximation`。

## 目前結果

`title-thunder.json`：

| 項目 | 值 |
|---|---|
| 編碼 | Vorbis、44 100 Hz、單聲道、`-q:a 5` |
| 時長 | 0.961451 秒 |
| 原版擷取四段非靜音視窗 | 0.935、1.057、0.942、0.910 秒（平均 0.960870） |
| 與平均值的差 | 0.000581 秒 |
| 解碼 | 通過 |
| 低於 −50 dB 且超過 50 ms 的靜音段 | 0 |
| 峰值／平均 | −7.5 dB／−10.3 dB |

## 人耳試聽入口

[離線試聽頁](../../../workplace/audio/listening-review-20261004/audition.html) 包含同一候選與
四段原版模擬器捕捉，所有音訊內嵌，不讀網路。OGG候選保持原SHA-256；WAV備用只作解碼，
沒有改音量。原版片段保留前後各0.1秒的既有樣本，不補零、不重錄。

材料、來源及處理參數見[本機收據](../../../workplace/audio/listening-review-20261004/receipt.json)。
原捕捉的RIFF長度欄位為28、data為0，Python嚴格讀取器拒絕；FFmpeg建立的研究副本與
原檔offset44後的PCM逐位元組相同，原檔未改。失敗及副本保留在同一工作區。

原版錄音與內嵌錄音的HTML只留本機，不提交GitHub或公開。
使用者已於2026-10-04實際聆聽並核准正式播放。原版捕捉未固定seed，
只供聆聽音色，不作亂數或同狀態對拍。

## 正式播放

[正式收據](title-runtime.json)與[RE-228](../../re/20261004-title-audio-runtime.md)。
雙語正常啟動各23次cue，首格雙聲；第23聲播放中SPACE離開，輸出尾端一秒為零。
實際非靜音時長另驗，全部internal／cmd通過，RE-227真實v3重招離城31狀態及6圖保持。
兩次固定參數新OGG雜湊相同，與原已驗收OGG解碼相同；候選原檔雜湊保持。
硬體時序為規格近似，不宣稱逐波形、全遊戲parity、三平台實機或已發布封包更新。

轉檔入口為 tools/build_title_audio.py --out <不存在的輸出目錄>，在
eob-remake-audio:20260901 固定image執行。參數與兩次輸出雜湊見RE-228收據。
新OGG固定serial且解碼相同，保留原已驗收OGG，不以新封裝位元組取代。
title-thunder.json保留最初技術量測，其當時人耳尚未執行由此次正式收據補足。

## Windows輸出續核

[RE-230](../../re/20261005-title-audio-platform.md)以同來源重建Windows及macOS候選。
Wine9預設DirectX正常啟動，雙語各23次cue，首格雙聲完整；第23聲中正常SPACE離開，
實際有效聲長21.35415／21.35701秒，尾端一秒為零，正常Ctrl-Q退出0。
收據為[windows-review.json](../../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/title-audio-platform-20261005/windows-review.json)。
錄音須先確認已寫出靜音PCM才啟動遊戲，舊開頭截斷的兩份錄音均由首格完整性閘門拒絕。
這是實際Wine輸出驗證，Windows真機聆聽與macOS啟動仍待驗，現有封包未更新。
