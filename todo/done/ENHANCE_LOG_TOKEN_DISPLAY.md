# 任務藍圖：優化日誌表格 Token 顯示 (Enhance LogTable Token Display)

## 1. 需求描述 (Requirement)
目前日誌表格 (`LogTable.vue`) 雖然已新增 Token 相關欄位，但配置上較為分散。為了方便用戶直接在列表視圖中掌握資源消耗情況，需調整欄位順序與可視性，將 `TotalTokens` (總 Token) 設定為預設顯示的高優先級欄位。

## 2. 變更範圍 (Scope of Changes)

### 🎨 前端組件：`web/src/components/logs/LogTable.vue`
- **欄位配置調整**：
    - 將 `total_tokens` 欄位移動到 `status_code` 之前。
    - 調低 `prompt_tokens` 與 `completion_tokens` 的預設可視性 (`defaultVisible: false`)，改以 `total_tokens` 為主。
    - 增加 `total_tokens` 的顯眼度。
- **欄位定義優化**：
    - 設定 `total_tokens` 的 `required: true` 或確保其在預設顯示清單中。

## 3. 執行步驟 (Execution Steps)
1. **修改全域列配置 (`allColumnConfigs`)**：
    - 調整欄位順序。
    - 更新 `prompt_tokens` 與 `completion_tokens` 的 `defaultVisible` 屬性。
2. **重置本地緩存測試**：
    - 提醒開發者或測試時需清除 localStorage 的 `log-table-visible-columns` 以套用新預設。
3. **驗證**：
    - 進入 `/logs` 頁面，確認「總 Token」欄位是否直接出現在表格中。

---
*Created: 2026-02-25 | MartletMolt*
