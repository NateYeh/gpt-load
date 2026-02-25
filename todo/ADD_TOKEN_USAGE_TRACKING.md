# 實作藍圖更新：Token 消耗統計 (Token Usage Tracking)

## 1. 當前進度狀態 (Current Status)

### ✅ 已完成 (Finished)
- **資料模型**：`internal/models/types.go` 已新增 `PromptTokens`, `CompletionTokens`, `TotalTokens` 三個欄位。
- **類型定義**：建立 `internal/proxy/types.go` 統一管理 `usageInfo` 結構，解決 Package 內重複定義衝突。
- **請求攔截**：`internal/proxy/request_helpers.go` 已實作自動注入 `stream_options: {"include_usage": true}`。
- **回應解析器**：`internal/proxy/response_handlers.go` 已實作解析 Logic。
- **代理核心整合**：已修復 `internal/proxy/server.go` 的語法錯誤（修復了 `{}` 匹配問題並清理了重複代碼）。
- **前端對齊**：
    - `web/src/types/models.ts` 已新增 Token 欄位。
    - `web/src/locales/` (zh-CN, en-US, ja-JP) 已新增相關 i18n 鍵值。
    - `web/src/components/logs/LogTable.vue` 已增加顯示欄位與詳情模態框更新。
- **編譯測試**：後端通過 `go build` 與 `golangci-lint`。

### 🚧 進行中/損壞中 (In Progress / Breaking)
- 無。系統已恢復穩定編譯狀態。

### ❌ 未完成 (Pending)
- **環境驗證**：建議在帶有 Node.js 環境的機器上執行 `npm run build` 進行前端最終打包確認。

---

## 2. 變更說明
- **類型統一**：將 `usageInfo` 從 `server.go` 抽離至 `types.go`，確保 `proxy` package 下只有一份定義。
- **語法修復**：修正了 `executeRequestWithRetry` 中因 `replace_block` 失敗導致的語法毀損（缺失 `}` 以及重複的 URL 構建邏輯）。

---

## 3. 下一步建議
1. 視需要啟動服務並進行 E2E 測試，觀察日誌中是否正確記錄 Token。
2. 檢查資料庫遷移是否如預期運作（GORM `AutoMigrate`）。

---
*Last Updated: 2026-02-25 09:10 (MartletMolt/AI)*
