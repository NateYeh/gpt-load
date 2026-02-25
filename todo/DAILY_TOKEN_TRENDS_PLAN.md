# 任務藍圖：7天 Token 消耗趨勢統計 (Daily Token Trends)

**版本**：2.0 (對齊現有代碼結構)
**狀態**：規劃中
**負責人**：MartletMolt (MCP Server)
**目標**：在儀表板新增一個以「天」為單位的圖表，顯示過去 7 天的 Token 消耗趨勢，並將目前的即時 Token 計算優化為資料庫預聚合模式以提升效能。

---

## 🏗️ 架構設計規劃

### 1. 資料存儲層 (Persistence)
目前 `group_hourly_stats` 僅記錄請求數，而 Token 消耗是透過 `request_logs` 即時 `SUM` 計算，這在數據量大時會造成儀表板讀取延遲 (Lag)。
- **優化方案**：在 `group_hourly_stats` 新增 `total_tokens` 欄位（未來可擴充 `prompt_tokens`, `completion_tokens`）。
- **優點**：查詢速度提升 100x 以上，降低資料庫 I/O。

### 2. 後端 API 層 (Golang - Gin)
- **端點優化**：擴展 `GET /api/dashboard/chart` 端點，支援參數 `?range=7d`。
- **邏輯描述**：
    - 若 `range=24h` (預設)：返回 24 小時數據（由 `group_hourly_stats` 提供請求數，優化後也由其提供 Token 數）。
    - 若 `range=7d`：將 `group_hourly_stats` 按天 (Truncate to Day) 進行聚合後返回。
- **時區處理**：強制使用 `Asia/Taipei` 零點作為切分點。

### 3. 前端展示層 (Vue 3 - Naive UI)
- **組件更新**：修改 `Dashboard.vue` 或 `Chart` 相關組件。
- **UI 配置**：
    - 增加 [24H] | [7 Days] 切換按鈕。
    - 7 天趨勢下，X 軸顯示日期 (MM-DD)。

---

## 📝 執行步驟清單

### 第一階段：資料庫模型更新 (Migration)
- [ ] 修改 `internal/models/types.go`：在 `GroupHourlyStat` 結構體新增 `TotalTokens int64`。
- [ ] 建立遷移文件 `internal/db/migrations/v1_3_0_AddTotalTokensToStats.go`：
    - 新增 `total_tokens` 欄位。
    - **回填數據 (Backfill)**：從 `request_logs` 統計過去 7 天的數據並寫入 `group_hourly_stats`。
- [ ] 在 `internal/db/migrations/migration.go` 註冊新遷移。

### 第二階段：統計邏輯同步 (Service)
- [ ] 修改 `internal/services/request_log_service.go`：
    - 在 `writeLogsToDB` 函式中，更新 `GroupHourlyStat` 的 `Upsert` 邏輯，增加 `total_tokens` 的累加計算。

### 第三階段：API 邏輯優化 (Handler)
- [ ] 修改 `internal/handler/dashboard_handler.go`：
    - 重構 `Chart` 函式，讀取 `group_hourly_stats.total_tokens` 取代現有的即時 `SUM(request_logs)`。
    - 實作 7 天範圍聚合邏輯。

### 第四階段：前端開發與串接
- [ ] `web/src/api/dashboard.ts` 調整 API 參數。
- [ ] 儀表板頁面新增範圍切換按鈕與對應圖表渲染邏輯。

---

## ⚠️ 潛在風險與考量
- **歷史數據完整性**：若之前的 `request_logs` 沒記錄到 Token (舊版本)，回填數據會為 0。這反映了實際的使用紀錄，屬正常現象。
- **SQLite 併發鎖定**：大量的 `CreateInBatches` 回填動作需在交易中小心處理，避免鎖死資料庫。

---

**備註**：此任務為 Token 管理效能優化之關鍵步驟。
