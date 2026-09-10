# 建築離開的告別條件

輸入：`KOLTOWN-unpacked.EXE`，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
工具：IDA Pro 9.4、`ida-pro-9.4-idapython:locked-v1`；位址基準全為 IDA linear。
查詢使用既有 `.i64` 的函式邊界及交叉參照，正式資料庫唯讀，沒有改名。
可重生工具：`tools/ida_probe_player_gaps.py`；完整指令與 bytes 僅本機保存於
`workplace/player-gaps-20260908/player-gaps-private.json`。

- **已證實**：`sub_11A93:0x12238` 的 DOOR 分支跳往 `0x13D64`；清除目前發言者、
  重繪隊伍後，`0x13D6F` 比較 `word_2DF2A` 與 1。
- **已證實**：該欄位是隊員數，沿用既有招募／住宿與獎勵證據；本輪 xref 包含
  招募 `sub_13D9A:0x13DC7` 加一、住宿 `sub_14DB1:0x15290` 減一，
  `sub_1E3A8:0x212BA` 分配獎勵時作除數。不是顯示模式。
- **已證實**：隊員數為 1 時跳 `0x13D8E`，不讀告別也不等待；其他情況在
  `0x13D7D` 推入建築 payload 偏移 `0x4C0`，由 `sub_14234` 讀文字，
  `0x13D89` 呼叫 `sub_144C3` 等待。
- **已證實**：既有文字抽取器起點 `0x80`、每則 `0x40` bytes，因此告別是
  `(0x4C0-0x80)/0x40 = 17` 號訊息。runtime 可直接使用既有 typed data。
- **已證實**：`sub_144C3:0x144CC` 清圖示，`0x144D1` 推圖示 1（SPACEBAR），
  `0x144E1` 呼叫原版圖示輸入；不是按完 DOOR 立刻移到地圖。

正常 dosgolem 的兩人離店畫面已有
`docs/verification/battle-normal-seed-20260908.md` 收據；本輪需另外抽驗單人不等待。
訂正先前稽核的過廣描述：不是所有離店都缺告別，**只有多人成隊的正常離店**
需要補這個等待。空隊伍沒有正常離店入口，不據此推定空隊伍的玩家功能。
