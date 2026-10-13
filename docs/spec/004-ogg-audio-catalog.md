# SPEC-004：OGG 音訊目錄與執行期契約

狀態：**CONFORMED**（RE-228；限已驗收標題OGG、cue次數與正式播放）
是否授權實作：**是**

## 目的

HD配樂試聽入口：[標題與旅店室內樂](../verification/hd-music-20261012/review.html)。
此批為重製版原創。使用者於2026-10-13回覆「這樣ok 那就繼續」，已接受標題與旅店試聽原檔；同日另接受探索與戰鬥。接入契約與[驗證入口](../verification/hd-music-runtime-20261013.md)見下方。舊標題雷聲CONFORMED範圍保持。


把原版可聽事件轉成可追溯的 OGG 資產目錄，並保證 remake 執行期只讀 OGG，不直接模擬
DOS 音訊硬體或讀取原版 executable。原版資料不進 runtime；轉換管線、合成器與 Ebitengine 音訊路徑依本規格實作。

## 已固定證據

- 固定 ZIP 沒有獨立的 WAV、VOC、MIDI 或已辨識音訊 archive。
- `TITLE.EXE` 的 `sub_10F75` 已證實控制 PC Speaker 與 PIT channel 2：共兩段音爆，每段
  更新 100 次除數；`sub_11953` 是 `rand() % N + 1`，`sub_11904` 是依全域校正常數忙等。
- `TANDYPIN.EXE` 入口執行 DOS `INT 21h`、`AH=31h` 常駐記憶體，保留 `0x0401`
  paragraphs；它不是獨立音訊素材或音效播放器。
- 六個 executable 均以兩個全新 IDA DB 窮舉已定義 code 的 x86 `IN`／`OUT`。只有
  `TITLE.EXE` 命中 `0x61`／`0x42`；其餘五個沒有音訊候選。六個保守未解析的 `DX`
  I/O 點經人工審查全是既有圖片 renderer 的 EGA graph-register writes。
- `sub_10F75` 的八個 caller 全在 `_main`、`sub_10AF0`、`sub_10C86`，且都緊接
  `sub_11133` 圖片 blit；單次 cue 路徑隨後進入鍵盤輪詢。這證實 cue 位於正常 TITLE
  presentation flow，但尚未把各張畫面推定成原版語意名稱。
- 依 DOSBox-X 官方 `DX-CAPTURE /A` 契約，在正常 `TITLE.EXE` 流程取得 48 kHz、stereo、
  PCM WAV。38.229333 秒觀察窗內有四段非靜音輸出，約為 0.935、1.057、0.942、0.910 秒；
  peak 為 -10.317 dB。這是 emulator machine capture，不冒充人耳驗收或原硬體時鐘 parity。
- 原始定位、bytes、caller 與工具版本見 `docs/re/ida-fingerprints/`；目錄見
  `docs/re/audio-inventory.json`。

## 資產契約

每個可交付 cue 至少記錄：穩定 ID、玩家觸發條件、原版函式／位址、證據等級、生成方法、
取樣率、聲道、時長、loop、音量基準、來源權利、OGG SHA-256 與人耳驗收結果。檔名只用
穩定 ID，不把推測性語意當作原版名稱。

執行期允許副檔名只有 `.ogg`；原版 EXE、archive、即時 PIT 合成與未列入 catalog 的檔案
一律失敗即關閉。技術解碼通過不等於人耳驗收。

## 硬體時序停止線

不再反組譯 DAC／PIT／busy-wait 的逐週期 wall-clock。若重建 TITLE cue，只依公開 PIT
規格與成熟模擬器契約求合理的 sample duration 與播放完成閘門，並標記為
`hardware-spec approximation`；不得宣稱逐波形或實機時鐘 parity。

## 已由使用者決定的路線

2026-09-01：TITLE 的可聽事件採**規格近似合成 OGG**，不散布原版錄音。

## 已固定的編碼與權利契約

- 編碼：Vorbis、44 100 Hz、單聲道、`ffmpeg libvorbis -q:a 5`。
- loop：TITLE 雷聲不循環；每次觸發播放一次。
- 音量基準：合成峰值對齊 DOSBox-X 擷取量測的 −10.317 dB。
- 權利分類：remake 自有、可公開——波形由公開的 PC Speaker／PIT 規格與已證實的演算法
  產生，沒有取樣或轉換任何原版音訊。
- 目錄與逐 cue 欄位在 `assets/audio/manifest.json`，技術驗收收據在
  `docs/verification/audio/`。

## READY 審查與執行期接線

2026-10-04使用者明確回覆「已聆聽，可接受並接正式播放」。已驗收候選SHA-256
`0e95e887d2513161a4f07d85fad3c64d7fc538d3b61864256f787d98db85d362`保持。
來源、格式、技術解碼及原版四段活動窗比對見既有收據；READY審查見
[RE-228](../re/20261004-title-audio-runtime.md)。

1. 原版閃電各連放兩次，SET每圈一次、FACE三次；同一92畫格序列每輪23 cue。
2. 畫格先呈現，cue逐次播放，不重疊；實際播放完成後才計算該格tick等待。
3. 音訊catalog只允許已驗收且SHA-256相符的OGG。缺檔、路徑、schema、取樣率、loop或解碼不符時拒絕，不讀原版資料或即時合成代替。
4. 音訊狀態不進人物v3或怪物v2，不消耗遊戲亂數。離開標題停止播放。
5. 生成PCM峰值目標−10.317 dB，既有Vorbis解碼峰值−7.5 dB；有損過衝及此次聆聽接受的音量保持，不重新正規化。

## 已驗收資產

`assets/audio/title-thunder.ogg`為`sub_10F75`的規格近似合成，時長0.961451秒。
與DOSBox-X四段活動窗平均差0.000581秒，僅證實結構與合理長度，不稱逐波形或硬體時鐘parity。
正式接線與正常輸出驗收由RE-228追蹤；READY本身不代表runtime已完成。

## 正式播放驗收

[正式收據](../verification/audio/title-runtime.json)。
雙語正常啟動各23次cue，首格雙聲；第23聲播放中SPACE離開，輸出尾端一秒為零。
實際非靜音時長另驗，全部internal／cmd通過，RE-227真實v3重招離城31狀態及6圖保持。
兩次固定參數新OGG雜湊相同，與原已驗收OGG解碼相同；候選原檔雜湊保持。
硬體時序為規格近似，不宣稱逐波形、全遊戲parity、三平台實機或已發布封包更新。

## HD原創配樂試聽

使用者選定A：中世紀室內樂，重視耐聽與遊戲音效；2026-10-13接受四首候選並要求接入。曲目原檔與獨立控制A案已核准。
原版為1989年DOS奇幻RPG，目前已證實的音訊是標題PC Speaker雷聲，沒有已證實背景旋律。
證據沿本檔、[音訊研究](../re/audio.md)及手冊[王國背景](../manual/chapters/08-realm-of-ashtalarea.md)，不新增硬體逆向。

- 玩家動作：組隊旅行、逐人規劃攻防、返回旅店整理與保存。
- 聲音材料：矮人石廳、林地與木造旅店、旅行者的皮革與金屬裝備。這是新配器意象，不宣稱原作音樂史實。
- 新作動機：D5、A5、C5、B♭4、E5、D5。節拍位置0、2、4、6、8、10，長度1.5、1、1.5、1、1.5、2拍。勝利意義取「活著回來」，保留懸念。
- 標題：72秒、80BPM、24小節4/4；三段各8小節。長笛提出主題，中提琴短答，大提琴開放音程，撥弦留空，少量手鼓只在中段。
- 旅店：80秒、72BPM、同網格24小節；主題低八度，弓弦與撥弦更輕，適合讀對話與久留。沒有酒館喧鬧人聲。
- 每首含立體聲OGG、MIDI、完整音符與三個等長分軌：旋律、和聲、脈動。尾韻折回同一小節長度，接點10ms輕窗；各軌共用增益，七種非空聲部組合均不削波。runtime候選仍只讀OGG。
- 禁用合唱、大片鼓、持續高潮、現成曲目取樣及指定作曲家模仿。不是原作旋律修復。
- 工具為`tools/compose_hd_music.py`與`tools/Dockerfile.music`；TiMidity++2.14.0-8.1、FluidR3GM3.1-5.3，音色庫MIT全文與SHA-256見[清冊](../verification/hd-music-20261012/manifest.json)。無外部生成服務、模型、訂閱或亂數種子。
- 四首與獨立控制已核准，接入依下方契約。危險、首領、勝敗與推廣片配樂仍為後續草案。

最終OGG檢查：44.1kHz雙聲道、Vorbis q6；標題−21.2LUFS、真峰值−5.9dB，旅店−21.4LUFS、−6.3dB。
三輪檔、雷聲共存示範與分軌均在試聽頁。機器檢查不代表已完成人耳疲勞、辨識或音效清晰度驗收。

## 遊戲內配樂接入

四首曲目接入及控制為**有限CONFORMED**，見[正式收據](../verification/hd-music-runtime-20261013.md)。使用者已選[控制比較](../verification/hd-music-control-20261013/review.html)A：配樂預設開啟，F9及下緣按鈕獨立切換。換EGA／HD保持目前配樂位置；開關只在本次遊玩有效，不存v3。
這是使用者已核准的新作音樂，沒有原版BGM或位址parity聲明。接入保持已確認OGG的位元組與SHA，曲目不重新配器或正規化。

- runtime資產放`assets/audio/hd/`，版本`kol-hd-music/1`，限`title-chamber.ogg`、`inn-chamber.ogg`、`exploration-chamber.ogg`與`combat-chamber.ogg`，各自記錄44.1kHz、雙聲道、72／80／90／60秒、循環、來源、授權與本輪使用者接受。
- 播放前驗完整目錄、SHA、實際聲道、取樣率與解碼時長。缺件或損壞拒絕整包，只關閉背景樂並顯示中文錯誤；雷聲、遊戲、目前畫面與v3保持，可用F9重試。
- 標題曲只在正常標題序列播放；旅店曲只在Halfway House內，包括整理角色子頁。探索曲在離城後的世界探索播放；戰鬥曲在遭遇圖、戰場與撤退確認播放。任務說明、戰利品、書本、城內街道、其他建築及結局停止背景樂。
- OGG解成記憶體中的浮點立體聲，接Ebitengine循環播放器。相同場景每幀不重建、不重播；場景切換約0.3秒淡出／淡入，原雷聲保持原音量、cue次數及完成閘門。
- 圖像切換不重建音樂播放器。預設`-music hd`，`-music off`關閉；批次截圖仍靜音。按鈕在960邏輯座標744,576至840,600，HD乘二；保留既有F10位置、樣式與輸入語意，錯誤用現有提示層。
- 只在互動模式播放；批次腳本保持決定性與靜音。不取遊戲亂數、不進snapshot或v3，關閉程式釋放播放器。
- 驗收包括正常標題與旅店輸出、離開停止、循環接點、圖像切換不中斷、缺件負例、對應中文、同輸入狀態／保存不變。工具為既有Go與音樂Docker，來源見前批清冊。

探索與戰鬥已由使用者於2026-10-13回覆「兩首都可接受，接入遊戲」核准，接入為**有限CONFORMED**。探索放慢並拉長新動機，戰鬥縮短動機形成輕脈動；保持室內樂、音效空間與三個同步分軌。
入口為[探索／戰鬥試聽](../verification/hd-adventure-music-20261013/review.html)，各90／60秒、64／96BPM，24小節4/4。
工具為`tools/compose_hd_music.py --tracks exploration combat`及`tools/verify_adventure_music.py`，原兩首音符與已接受OGG保持。
