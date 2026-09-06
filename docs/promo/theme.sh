# 推廣片的設計 token —— 這一部片專屬，不沿用別部片。
#
# theme 名字：**赤石與燭火**。
#
# 色票不是憑喜好挑的，是從實機截圖的色彩直方圖取出來的
#（`convert <圖> -resize 100x100 -colors 6 -format %c histogram:info:-`）：
# 城鎮與戰鬥畫面的最大宗是磚紅 `#A20506` 與近黑 `#0C0808`，野外夜景是
# `#000015`。所以深端取暗紅黑、淺端取磚紅，強調色取中世紀奇幻慣用的鎏金，
# 文字取骨白——與遊戲畫面同一個色系，截圖擺上去不會像貼在別人的底上。
THEME_NAME="赤石與燭火"

BG_DEEP='#140606'   # 背景漸層深端（近黑的暗紅）
BG_LITE='#3d0f0b'   # 背景漸層淺端（磚紅壓暗）
ACCENT='#c8992e'    # 鎏金：標題、框線
ACCENT_SH='#6e4f12' # 鎏金的陰影層
TEXT='#f0e6cf'      # 骨白：正文與字幕
DIM='#b08a5e'       # 次要文字

# 西方奇幻用襯線；黑體會變成手遊味。
FONT_TITLE=/usr/share/fonts/opentype/noto/NotoSerifCJK-Bold.ttc
FONT_BODY=/usr/share/fonts/opentype/noto/NotoSerifCJK-Regular.ttc

W=1280; H=720; FPS=25
PACE_CARD=6         # 標題卡秒數
PACE_SLIDE=5        # 截圖段秒數
