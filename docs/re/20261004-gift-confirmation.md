# RE-220：已核准送禮確認流程的正式接線

更新日期：2026-10-04。SPEC-031 第17節的兩人巨劍窄鏈有限 CONFORMED。
人物存檔 v3、怪物資料 v2 保持；滿背包與披風提案仍未批准。

## 證據與核准

原版來源為 KOLTOWN.EXE，SHA-256
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。
固定 dosgolem `968ca774705853276ae55c8d1163c369481ac71f`、Go 1.26.7，
兩側正常啟動 seed 1，沒有 direct-entry、狀態注入或正式鎖定測試結果。
原版 201 鍵／207 狀態，六條正常鏈的來源、工具、原始記錄及畫格定位見
[原版研究](20261001-item-detail.md#2026-10-04原版正常送禮與逐人保存續核)。
本輪先重算既有 receipt.json 的全部附件及原型驗收，不重新開啟已證實 helper。

使用者明確選「採用原版確認流程（建議）」後，將第17節 DRAFT 升為 READY。
確認前保留物品明細與完成訊息，ENTER／SPACE 才搬移完整16-byte記錄，再回首欄。
核准範圍為五個正式檔案；唯讀 main.go 只作觀察 overlay，不進正式程式。
`ready-review.json` 保存核准、既有原型及來源收據 SHA-256。

## 正式修正與有限驗收

- 五檔基底 SHA 與原型來源完全相同才套用，不覆蓋後續異動。
- GiftConfirm／GiftTarget／GiftMessage 只為畫面暫態，不進 v3 保存。
- 更新既有送禮回歸，驗 ENTER、SPACE、圖示；等待中整份人物不變。
- 雙語各四案例，共292正常鍵／300狀態；正式與觀察 trace、300 PNG 完全相同。
- 全部300正式畫格與已批准原型相同，繁中確認訊息及巨劍明細可讀。
- 原版確認前後兩人的17物品槽共272次比較相符，完整原始16 bytes保持。
- 完整 v3 保存與原型只差 saved_at，第三口袋巨劍重載保持，52張瀏覽圖相同。
- 真實視窗雙語各108鍵，合計216鍵；26個預先固定完整畫格通過。
- 六個 GUI 程序以 WM_DELETE_WINDOW 正常關窗，退出碼0；重載不改保存。
- 全部 internal／cmd 通過；既有戰鬥雙語325狀態與16 PNG逐檔保持。

這完成兩人腰帶巨劍的確認、原始槽、UI、保存、退出及重載垂直鏈。
不宣稱六人、其他物品、野外、滿背包或完整原版畫面 parity。
RE-219 的第五回合後續投石差異保持，八回合閘門仍未通過。

## 候選與重驗

334 Go／336程式輸入，五正式檔案與一測試檔案。三平台同來源乾淨重建，
Linux重建逐byte相同，新Windows／macOS只驗結構，沒有真機驗收。
本機候選與完整SHA在 `platform-checks.json`，不定版或公開發布。

工作區 `workplace/item-give-confirmed-20261004/`：
`prepare.py` 審查核准及保存修改前來源；`build.sh`／`run.py` 重建與正常觀察；
`run_formal.sh`／`run_formal.py` 正式重播；
`prepare_gui.py`／`run_gui.sh`／`run_gui.py` 真實按鍵與畫格同步；
`prepare_outputs.py`、`cross_build.sh`／`mac_build.sh`／`check_platform.py` 跨平台中間候選；
`run_battle_regression.sh`／`run_battle_regression.py` 保留戰鬥既有行為；
`compare.py`、`finalize_docs.py`、`seal.py`／`verify.py` 為比較及封存入口。

唯讀重驗用 kol-package:20260907，專案掛 `/src:ro`，
執行 `python3 -B /src/workplace/item-give-confirmed-20261004/verify.py`。
UID/GID1000:1000、無網路、120秒、1GiB、1CPU及64程序。
輸入、原始畫格及角色資料僅本機研究，沒有新增散布權。

第一次全量檢查的 TestGiveItem 沿用立即轉交期待，已依核准契約更新。
GUI首次誤用截圖用 -scene flow，flow未初始化；失敗保留在 initial-gui-save-en。
改為正式預設啟動及重載只指定 -load，同image乾淨重跑通過，沒有產品修補。
容器、UID/GID、暫存與附件檢查在 audit.json；本輪無提交、推送、Issue寫入、打包或發行。
