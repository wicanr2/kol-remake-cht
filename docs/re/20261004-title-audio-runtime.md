# RE-228：已驗收雷聲接正式標題播放

日期：2026-10-04。狀態：CONFORMED，限已驗收標題cue與執行期接線。

使用者明確回覆「已聆聽，可接受並接正式播放」，核准現有
`assets/audio/title-thunder.ogg`，SHA-256
`0e95e887d2513161a4f07d85fad3c64d7fc538d3b61864256f787d98db85d362`。
候選的單聲道、44100 Hz、0.961451秒、非循環與已驗收音量保持。
規格近似合成的權利與技術收據沿 SPEC-004，不散布原版錄音。

原版來源 TITLE.EXE SHA-256
`8a2879d924d9a9cb01e000bd548c348022c0f922288565015f7560ab8fc6fad6`。
固定序列 `docs/re/title-sequence.json` 的 primitive lightning 位址
IDA Pro 9.4 linear `sub_10C86:0x10C86`，呈現 BACKPICS[6] 後連放兩次雷聲。
SET 每圈一次，FACE 三張全螢幕閃光各一次；每輪預期23個 cue。
現有 Frame.Thunder 布林與估計停留不能表示閃電的第二次 cue。
已證實的是呼叫次數及順序；音訊長度與硬體 wall-clock 為規格近似，不升格硬體 parity。

## 預先登記

正式 runtime 只從核准 catalog 讀 OGG，驗 schema、穩定 cue ID、路徑、SHA-256、
非循環、取樣率與已驗收標記。缺檔、越界與解碼錯誤拒絕播放，不改讀原版 EXE。
音訊介面用鎖定 Ebitengine 2.9.9，44100 Hz context，解碼候選後逐 cue 播放。
標題畫格先呈現，再依已證實次數播放；等實際播放完成後才處理該格 tick 等待。
離開標題時停止播放器，聲音不進存檔，不消耗遊戲亂數。

驗收包含單／雙 cue、播放未完不換格、不重複觸發、正常任意鍵離開，
catalog／解碼負對照、全部 internal／cmd、固定標題畫格及雙語正常玩家路徑回歸。
實際容器音訊輸出另擷取非靜音與 cue 次數；技術輸出不冒稱跨平台實機聆聽。

工作區 `workplace/title-audio-runtime-20261004/` 保存來源快照、READY 審查、
建置、輸出擷取及收據，入口 prepare.py；後續腳本在本入口補索引。
`build.sh`在鎖定Go音訊image執行正式模組的離線相依解析、測試與Linux建置。
初次使用根目錄`go mod tidy`掃入研究工作區的其他專案副本；這是建置範圍問題。
改為正式`internal`與`cmd`套件，沿用Ebitengine鎖定的音訊相依，不載入研究副本。
`capture.sh`與`capture.py`從正式啟動、實際音訊輸出及鍵盤離開驗23次cue，雙語各一輪。
轉檔入口`tools/build_title_audio.py --out <不存在的輸出目錄>`在
`eob-remake-audio:20260901`執行，固定FFmpeg版本、參數與OGG serial；另輸出解碼比較收據。
`regression.sh`與`regression.py`使用RE-227的實際存檔，從正常開場重招兩人、離城，
比較雙語完整31狀態與各3張畫格。原版錄音與過往收據維持唯讀。
第二次英文擷取已播一輪23聲；腳本等候到下一輪並使用12毫秒快捷鍵，
Ctrl修飾鍵未跨越畫面更新，退出超時。分類為GUI腳本問題，保留輸出後
改以跨更新的正常keydown／keyup，第23聲播放中按SPACE驗停止，未改遊戲鍵位。
FFmpeg收到主動SIGINT正常返回255，鎖版控制實驗確認；錄音需另外驗完整解碼。
事件數與非靜音分開驗證。首次雙語23聲錄音含約100毫秒週期缺口，有效聲長不足。
軟體繪圖預設執行緒在2CPU配額下擠壓音訊；限制Mesa與Go執行緒後，同一binary首兩聲
控制實驗有效1.116190秒，正常第2聲中斷，CPU nr_throttled=0。完整雙語重跑後才驗收。
原版輸入唯讀，容器 UID/GID1000:1000、無網路、有界；不改舊RE封存。
同輪使用者另要求安排遊戲圖片與操作畫面HD化，另列工作清單，
不把未定視覺方向寫成正式 theme 或素材契約。

## 結果

雙語正常啟動各23次cue，首格雙聲；第23聲播放中SPACE離開，輸出尾端一秒為零。
實際非靜音時長另驗，全部internal／cmd通過，RE-227真實v3重招離城31狀態及6圖保持。
兩次固定參數新OGG雜湊相同，與原已驗收OGG解碼相同；候選原檔雜湊保持。
硬體時序為規格近似，不宣稱逐波形、全遊戲parity、三平台實機或已發布封包更新。
review_capture.py、finalize.py、seal.py與verify.py負責波形審查、目前狀態、封存及唯讀驗證。
舊失敗擷取與腳本分類保留，不當作驗收。未新增原版硬體逆向或發行聲明。
