"""封包驗證通過後同步穩定入口；遠端發布另由回讀收據確認。"""
from pathlib import Path
import json

R = Path('/src')
V = 'v.1.0.3-20261017'
D = R / 'docs/verification/release-v1.0.3-20261017'
receipt = json.loads((D / 'receipt.json').read_text())
assert receipt['technical_status'] == 'PASS' and len(receipt['packages']) == 6

p = R / 'README.md'
s = p.read_text()
# 現況與下載區換版；歷史截圖及其來源不改成新版收據。
s = s.replace('現行交付為 `v.1.0.2-20261013`', '現行交付為 `v.1.0.3-20261017`')
s = s.replace('從新開局正常完成全部任務至結局仍未驗證。開發版已完成13類畫面與操作介面的HD接入，見[完整驗收](docs/verification/hd-full-runtime-20261017.md)；新版三平台封包尚未重建。',
              '從新開局正常完成全部任務至結局仍未驗證。13類畫面與操作介面的HD已納入本版，見[完整HD驗收](docs/verification/hd-full-runtime-20261017.md)與[封包驗收](docs/verification/release-v1.0.3-20261017.md)。')
a, b = s.index('## 開始遊玩'), s.index('操作照原版參考卡：')
part = s[a:b].replace('v.1.0.2-20261013', V).replace('release-v1.0.2-20261013.md', 'release-v1.0.3-20261017.md')
part = part.replace('包含HD操作介面與頭像、四首原創室內樂、F9配樂及F10畫面切換。', '包含完整HD圖形與操作介面、四首原創室內樂、F9配樂及F10畫面切換。')
part = part.replace('三平台引擎包及校驗碼已發布至', '三平台引擎包及校驗碼的發布入口為')
part = part.replace('使用本版抽樣與遊戲內切換的新畫格，也只留本機。', '使用本版旅店、遊戲內切換及明示受控HD場景的新畫格，也只留本機。')
s = s[:a] + part + s[b:]
s = s.replace('旅店與設定截圖來自現行發行包。開發版另已補齊地圖、戰場、傷勢、天候、片頭及事件，以下對比分別標明來源。',
              '旅店與設定截圖來自本版實際AppImage，縮圖最長邊768。13類圖形及操作頁均已納入本版；下方各批對比保留原驗收來源。首次載入完整HD素材需要較長時間，之後切換沿用快取。')
s = s.replace('docs/verification/release-v1.0.2-20261013/inn-hd-controls.png', 'docs/verification/release-v1.0.3-20261017/inn-hd-controls.png')
s = s.replace('| 戰鬥資訊框的 HD 人物 | 遊戲內 HD 畫面設定 |', '| HD 戰場、傷勢與操作介面 | 遊戲內 HD 畫面設定 |')
s = s.replace('![戰鬥資訊框與七個HD指令](docs/verification/hd-interface-runtime-20261012/battle-hd.png)', '![HD戰場、傷勢與操作圖示](docs/verification/release-v1.0.3-20261017/battle-hd.png)')
s = s.replace('docs/verification/release-v1.0.2-20261013/game-menu-hd.png', 'docs/verification/release-v1.0.3-20261017/game-menu-hd.png')
s = s.replace('docs/verification/release-v1.0.2-20261013/manifest.json', 'docs/verification/release-v1.0.3-20261017/screenshots.json')
s = s.replace('### 開發版：', '### ')
s = s.replace('開發版已接入手繪書頁', '本版已接入手繪書頁')
s = s.replace('已沿同一銅色主題接入開發版', '已沿同一銅色主題接入本版')
s = s.replace('，尚未納入現行發行包', '，已納入本版')
s = s.replace('；尚未納入現行發行包', '；已納入本版')
s = s.replace('。尚未納入現行發行包。', '。已納入本版。')
s = s.replace('這批尚未納入現行發行包', '這批已納入本版')
s = s.replace('地圖、戰場、天候與未驗操作分支仍續作。', '地圖、戰場與天候已納入本版；未抽測玩法邊界仍依工作清單處理。')
s = s.replace('地圖、戰場、傷勢、天候、片頭與其餘事件仍續作。', '地圖、戰場、傷勢、天候、片頭與事件均已納入本版。')
s = s.replace('其餘圖形已接入上述完整HD開發版', '其餘圖形已接入上述完整HD')
s = s.replace('走路標記另行續作，其餘五城及世界背景見下節。', '走路標記已納入完整HD，其餘五城及世界背景見下節。')
s = s.replace('圖中的天候與城鎮頭盔仍沿既有圖形，後續一併HD化；世界事件、戰場、傷勢及片頭繼續。', '這批歷史截圖的天候與城鎮頭盔沿當時圖形；本版已補齊HD，見上方完整HD對比。')
s = s.replace('七段結局已接入開發版，沿用手繪、紅黑銅框與原生 HD 中文。現行 `v.1.0.2-20261013` 發行包保持，尚未包含本批。',
              '七段結局已納入本版，沿用手繪、紅黑銅框與原生 HD 中文。下方保留正式接入時的同主題比較與來源。')
s = s.replace('| 原版 DOS，320×200 | 開發版 HD，1920×1200 |', '| 原版 DOS，320×200 | 重製 HD，1920×1200 |')
p.write_text(s)

p = R / 'CONTEXT.md'
s = p.read_text()
a, b = s.index('現行交付為'), s.index('本檔保存目前狀態')
s = s[:a] + '''現行本機交付為`v.1.0.3-20261017`，三平台六包與76秒1080p影片驗收完成，private Release四附件待發布。入口：[本版交付](docs/verification/release-v1.0.3-20261017.md)。舊版tag與Release保持。
本版包含目前玩家路徑13類HD圖形及操作頁、四首原創室內樂、F9與F10。正式430份Go／module及832份HD資產保持完整HD驗收來源；中文與場景依SPEC-040，不宣稱自然全任務通關或新增原版parity。
''' + s[b:]
s = s.replace('新版`v.1.0.3-20261017`三平台完整版、引擎包與推廣影片正在重建，入口為[本版交付](docs/verification/release-v1.0.3-20261017.md)。現行下載仍為v.1.0.2-20261013，待新版全部驗收後更新。',
              '新版六包、76秒影片及資料匯入驗收通過；唯一現行交付根目錄為`dist-all/v.1.0.3-20261017/`。完整版與影片只留本機，發布範圍限private repo的三平台引擎包及SHA256SUMS.txt。')
s = s.replace('目前目標為完成全量HD。**13類圖形與操作介面的開發版接入已通過受控驗收**，新版平台封包為下一交付閘門，工作清單保留該項。',
              '全量HD及新版平台封包已完成本輪驗收。**13類圖形與操作介面已接入本版**，工作清單的HD交付項結案。')
s = s.replace('現行交付、tag與Release保持`v.1.0.2-20261013`。', '舊版`v.1.0.2-20261013`的交付、tag與Release保持。')
lines = s.splitlines()
for i, line in enumerate(lines):
    if line.startswith('| 正式來源 |'):
        lines[i] = '| 正式來源 | 430份Go／module、832份HD資產；程式SHA c1b5969e…，本版與[完整HD收據](docs/verification/hd-full-runtime-20261017/receipt.json)相同；[封包輸入](docs/verification/release-v1.0.3-20261017/build-inputs.json) | 正式規則與v3未改；收尾工具另有版本快照 |'
    elif line.startswith('| 最近測試 |'):
        lines[i] = '| 最近測試 | 全cmd／internal通過；一次正常1046狀態與13新程序重載，trace及完整v3排除saved_at相同；13場景跨平台52畫格與實際封包20組EGA零差、HD最多一色階；F9／F10與實際音訊通過 | SPEC-040受控場景、Linux容器與Wine；vet兩則原有警示保留 |'
    elif line.startswith('| 跨平台 |'):
        lines[i] = '| 跨平台 | [本版六包驗收](docs/verification/release-v1.0.3-20261017.md)：Linux AMD64、Windows AMD64、macOS Intel／ARM universal；實際AppImage及Wine兩個ZIP的啟動、匯入、F9／F10與音訊通過 | macOS未執行、Wine不是真機、主機FUSE未驗；首次HD載入成本見本版收據 |'
    elif line.startswith('| HD |'):
        lines[i] = line.replace('新版平台封包下一閘門；', '本版六包通過；')
    elif line.startswith('| 交付 |'):
        lines[i] = '| 交付 | `v.1.0.3-20261017`：三平台full／release、76秒1080p影片；[交付收據](docs/verification/release-v1.0.3-20261017.md)及[收尾工具](tools/closeout_release.py) | private四附件待發布；完整版及影片本機；無真機驗收 |'
    elif line.startswith('| GitHub Issue |'):
        lines[i] = line.replace('16開放／92已核實完成不變', '本輪HD本機項結案；遠端Issue不變')
    elif line.startswith('| hd-art-ui |'):
        lines[i] = '| hd-art-ui | 已完成13類HD、中文、F10、保存回歸及本版六包驗收；真機仍依獨立項目 |'
s = '\n'.join(lines) + '\n'
s = s.replace('全量圖形接入已通過；新版封包依[SPEC-039](docs/spec/039-hd-art-ui.md)交付閘門，不重新開啟已確認選項。',
              '全量圖形與新版封包已依[SPEC-039](docs/spec/039-hd-art-ui.md)交付閘門通過，不重新開啟已確認選項。')
p.write_text(s)

p = R / 'WORKLIST.md'
s = p.read_text()
s = s.replace('最新交付入口為 `dist-all/v.1.0.2-20261013/`，包含HD介面、頭像、四首配樂及F9／F10，驗收範圍見\n[本版收據](docs/verification/release-v1.0.2-20261013.md)。前版保持。',
              '最新交付入口為 `dist-all/v.1.0.3-20261017/`，包含13類HD圖形與操作頁、四首配樂及F9／F10，驗收範圍見\n[本版收據](docs/verification/release-v1.0.3-20261017.md)。前版保持。')
p.write_text(s)

p = R / 'docs/verification/hd-completion-20261014.md'
s = p.read_text().replace('沿用1920×1200、手繪、紅黑銅框及遊戲內 F10。現行 v.1.0.2-20261013 只有部分 HD，不能宣稱全部完成。',
                         '沿用1920×1200、手繪、紅黑銅框及遊戲內F10。現行v.1.0.3-20261017已納入下表13類HD，見[本版交付驗收](release-v1.0.3-20261017.md)。')
s = s.replace('下表依後續開發版驗收更新，發行包仍為v.1.0.2-20261013。', '下表依後續正式驗收更新，已納入v.1.0.3-20261017。')
s = s.replace('新版封包尚未重建。', '新版六包、中文、F9／F10與一色階驗收通過。')
p.write_text(s)

p = R / 'docs/spec/039-hd-art-ui.md'
s = p.read_text().replace('新版平台封包為下一交付閘門；', '新版平台六包已依[交付驗收](../verification/release-v1.0.3-20261017.md)通過；', 1)
# 最後現況結論同步；前面各批時間序列保留。
s = s.replace('平台封包仍為下一交付閘門。', '新版平台封包已依[本版收據](../verification/release-v1.0.3-20261017.md)通過。')
p.write_text(s)

p = R / 'docs/worklist.json'
j = json.loads(p.read_text())
item = next(x for x in j['items'] if x['id'] == 'hd-art-ui')
item.update(status='done', resolved='2026-10-17：13類HD接入與本版三平台六包完成，固定seed1正常保存重載、Linux／Wine實際F9／F10、52受控畫格跨平台及20個封包畫格一色階通過。真機與原版未知仍由原工作項追蹤。')
item['body'] = item['body'].replace('現行v.1.0.2-20261013保持；本項僅剩新版平台封包閘門，原版oracle與真機另列。', 'v.1.0.3-20261017新版六包與影片完成；舊版保持，原版oracle與真機另列。')
item['verify'] = dict(kind='json_len', path='docs/verification/release-v1.0.3-20261017/receipt.json', field='packages', max=5,
                      note='已完成項回歸訊號：正式交付收據須保留六個已驗封包；原版未知與真機不算此項缺口。')
j['updated'] = '2026-10-17'
p.write_text(json.dumps(j, ensure_ascii=False, indent=2) + '\n')
print('CURRENT DOCS UPDATED; private publication still READY')
