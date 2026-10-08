# RE-285：正常重載的法術與任務子頁

日期：2026-10-06。狀態：有限呈現抽樣完成，完整UI驗收保持。

RE-285正常重載法術與任務子頁：有效原版19狀態、雙語正式／唯讀觀察296狀態。Olear法術N兩次英文整頁除滑鼠零差；雙語兩次法術與六人任務共16次開頁，52行抬頭及肖像名牌保持，完整隊員資料除Speaker不變。任務頁六人均顯示已回報第6號徽章，但沒有本輪原版任務oracle。正式342來源／Linuxd43a6385、怪物v2／保存v3保持。

## 輸入與方法

原版沿RE-213正常住宿寫出的CHARDATA及三個人物圖片檔重載，SHA-256見parent-inputs.json。
CHARDATA為0b83960e316aef6c4f4fc8c31873b2da815cef9b86c4dfca83698bf1499c7c76。
以dosgolem 968ca774705853276ae55c8d1163c369481ac71f正常KOL→TITLE→BOOK→TOWN啟動，
SetScratch讀取實際既有保存，正常招募與數字鍵開人物頁；沒有角色、位置、入口注入或等待stub。
本輪沒有亂數規則比較，不重設重載中的seed。原始seed逐列保留。
原始EXE與資料SHA-256及大小在original-inputs.json，位址空間沿probe.go及原版私有收據。
DS:BB82發言者、DS:6A選取圖示、DS:6C圖示數量、DS:2DAE圖示表及六筆768-byte原始角色記錄僅唯讀觀察。

Go法術保存沿RE-213正常保存，任務保存沿RE-279現行雙語長任務完成保存。
人物v3含16個角色槽，party保存槽號，空槽與null spells依SPEC-023解析。
當前342來源／340 Go輸入在program-inputs.json；正式Linux SHA-256為
d43a6385a9c35486cc2cd383f78ffce23d70820f1fa113c98559d1760ff2b647。
雙語均以正式-scene play -load入口重載，沒有改保存、製造法術或注入任務結果。
法術頁選Olear，兩次開頁返回；任務頁逐一選六人、開頁、返回人物頁及城鎮。

原版Olear排第一，Go正常重招募後排第二；按鍵依同一人物映射，並非相同全隊排列。
原版與Go既有旅程不同，nutrition為10與8，本頁只顯示共同人物抬頭與法術N。
不宣稱完整角色記錄或整條RE-213旅程相同。

## 有效結果

- 原版法術入口第11、16觀察，圖示表縮成0f；Go英文第23、28步顯示單一法術N。
- 原版兩張完整320×200頁各72個原始差異，全在招募Click(202,182)留下的可見滑鼠游標。
  由原始圖確認游標位置，固定排除矩形200,182,14,14；其餘整頁零差。原始PNG與72像素差異保留。
- 雙語法術各兩次、六人任務各六次，子頁與同一人物前頁的上方52原始掃描線逐像素相同。
  52行包含完整肖像、名牌及人物抬頭；第52行起屬屬性框，不納入抬頭比較。
- 任務頁均使用AWD1PICS第6號徽章；各角色實際保存Progress第0組為1，回報旗標沿保存。
  只驗此一頁，其他任務組與多頁翻頁仍未驗。
- 有效原版人物頁區段全部六筆768-byte角色記錄不變；Go開頁段完整Party投影除Speaker不變。
  正式／唯讀觀察的完整軌跡剔除新增觀察欄後相同，所有296張Go PNG逐byte相同，載入保存雜湊不變。
- 已實際查看英文法術、繁中法術、繁中任務完整PNG。字串N及六名既有姓名是有限樣本，
  不取代最長字串、全部文字安全矩形、其他UI與HD驗收。

## 控制器修正與限制

首次原版控制器誤以為數字鍵後仍需SPACE確認，實際已開戶外人物頁，後續走到物品頁。
初次original/及plan.json完整保留，22狀態不得作法術頁對拍證據。
修正刪除多餘確認鍵，另存corrected-plan.json及original-corrected/；同一輸入、同一原版觀察binary乾淨重跑。
第一次修正wrapper把binary路徑一併替換而退出127，修復確切路徑後重跑，沒有產品程式變更。
初始資料讀取腳本遇到保存null槽／索引形狀錯誤，讀SPEC-023後修正，屬驗證工具問題。
沒有新原版任務頁oracle，未證實其他任務內容、多頁、野外或坐騎邊界。
本輪不重開已驗六人主資料頁，也不變更既有HD、送禮、受阻擊退等決策。

## 重生與唯讀稽核

工作區workplace/sheet-subpages-20261006/；上層sheet-subpages-20261006-prepare.py是準備入口。
build-all.sh編譯原版觀察器及僅新增序列化觀察欄的Go觀察器；observer-main.diff可審查。
run-original-corrected.sh從乾淨輸出重生有效原版；replay.sh／replay.py執行雙語正式及觀察版本。
analyze.py核對完整頁、抬頭、人物資料及觀察中立性，analysis.json保留結果。
seal.py封存附件、外部輸入、PNG來源與權利，verify.py作唯讀雜湊與中立性稽核。
Docker image為eob-remake-go:1.26.7-ebiten2.9.9-audio，固定image ID見receipt.json。
專案唯讀掛/src、network none、UID/GID1000、有界逾時與資源限制，在容器內執行
python3 /src/workplace/sheet-subpages-20261006/verify.py。只支援本輪固定保存與來源。
PNG及原版payload全部僅本機，不發布或提交。容器及檔案衛生見收據。
