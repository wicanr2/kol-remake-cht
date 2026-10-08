# RE-230：正式雷聲接線後的跨平台候選

日期：2026-10-05。狀態：有限跨平台建置與Wine路徑通過；真機及正式封包未驗。

RE-228已在Linux驗過OGG正式播放，現行336份Go、338份輸入保持。
本輪從同一來源離線重建Linux、Windows AMD64、macOS ARM64／AMD64及universal，
再用Wine預設圖形後端抽驗雙語正常城鎮／戰鬥及既有真實v3保存重載。
Windows音訊另走正常預設啟動、23聲序列、SPACE中斷與正常退出，不用腳本場景代替。

工作區為workplace/title-audio-platform-20261005/。prepare.py核對RE-228及RE-229來源，
cross-build.sh／mac-build.sh建置，check-platform.py核對檔案格式與架構。
本輪沒有正式版號、交付包或發布；候選只放本機工作區。macOS只能核對建置與結構，
不宣稱真機驗收。原版oracle與保存v3、怪物v2、HD待決版面保持。

輸入唯讀、UID/GID1000:1000，沿用鎖定Docker映像及離線module cache；不掛主機runtime。
封存入口同工作區seal.py／verify.py，操作與音訊收據只涵蓋實際完成範圍。

## 失敗定位

首次控制器單次以標題查視窗未命中，兩語正常回放已執行完，但音訊只到首聲。
第二次依可見960×600視窗等待，記錄名稱仍為武士傳說；視窗與鍵盤路徑完成。
錄音程序啟動後等待0.4秒不足以證明資料已開始寫出，繁中首段僅1.19254秒，
總有效聲長20.60651秒未達既定界限。錄音容器CPU節流為零，後續完整單聲約0.96145秒。
初次與trial2輸出、控制器來源及trial2-wave-diagnosis.json／trial2-failure.json保留。

trial3以同候選、同輸入與同工具映像乾淨重跑，先確認錄音檔已寫出PCM及至少一秒
完全靜音，再允許遊戲啟動。驗收增加首格雙聲完整性，不放寬總聲長。
review-windows.py通過。recording-readiness-controls.json保留負對照，原先英文首段
1.77667秒及繁中1.19254秒均拒絕，新英文首格雙聲1.91907秒通過。

## 有限驗收結果

| 項目 | 證據與範圍 |
|---|---|
| 程式來源 | 336份Go／338輸入與RE-228相符；不改正式程式或素材 |
| Linux | 乾淨重建SHA-256仍0aca34273a671a78eb6b6b5e93e4336fdead6e2db1aedc3aa8ca48ac24ef3496 |
| Windows | PE32+ AMD64；Wine9預設DirectX實際執行 |
| macOS | ARM64、AMD64與universal，兩個thin內容相符；最低11.0及系統相依庫通過 |
| 音訊依賴 | 四個thin/native候選均含Go1.26.7、Oto3.4.0、oggvorbis1.0.5及vorbis1.0.2 |
| 正常回放 | 雙語各184狀態，城鎮及戰鬥兩張PNG與RE-229相同 |
| 真實v3重載 | 雙語各31狀態、三張PNG與RE-228相同，既有實際保存雜湊保持 |
| Windows音訊 | 正常啟動各23聲，首格雙聲完整，第23聲中SPACE中斷，正常Ctrl-Q退出0 |
| 實際聲長 | 英文21.35415秒，繁中21.35701秒；輸出最後一秒完全靜音 |

平台結構收據為platform-build-checks.json及platform-build-info.json；實際Windows收據為
windows-review.json，通過的畫面、波形、軌跡及操作記錄位於trial3/。英文與繁中標題
閃電、跳過後書封已目視核對；繁中戰鬥及重載後正常世界畫面亦已核對。
Wine容器存在CPU節流，錄音容器沒有；軟體渲染耗時不作原版硬體時鐘對拍。

同區cleanup.py只移除本批三個Wine前綴及IPC／設定暫存，保留使用過的Pulse設定、
全部診斷、失敗與通過波形。receipt.json封存附件與靜態輸入；文件以documents/快照保存。
在eob-remake-go:1.26.7-ebiten2.9.9-audio有界、無網路、UID/GID1000:1000容器內，
唯讀掛專案至/src，執行python3 /src/workplace/title-audio-platform-20261005/verify.py。
來源更新後加--recorded-source，只驗封存附件，不宣稱現行回歸。

本輪沒有原版新增oracle、Windows真機聆聽、macOS啟動、完整通關或更新發行包。
下一閘門仍含HD版面與素材共同決策、其他玩家路徑、實機與版本／交付。
