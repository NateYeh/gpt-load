# GPT-Load 專案指南

**Last Updated**: 2026-04-28

---

## 專案概述

高效能 AI API 透明代理服務（Go + Vue 3），支援多供應商密鑰輪換、負載均衡與集群部署。

- **語言**: Go 1.24 / TypeScript (Vue 3)
- **框架**: Gin (HTTP) + dig (DI) + GORM (ORM)
- **資料庫**: PostgreSQL（生產）/ SQLite（預設）
- **快取**: Redis（可選，未設定時使用記憶體快取）
- **端口**: 47300（由 .env 設定）

---

## 目錄結構

```
gpt-load/
├── main.go                          # 程式入口（含子命令: migrate-keys）
├── Makefile                         # 建置與運行指令
├── .env                             # 環境變數（資料庫、認證、加密等）
├── data/                            # SQLite 預設資料目錄（生產環境用 PostgreSQL）
├── internal/
│   ├── app/                         # 應用生命週期（啟動/關機）
│   ├── channel/                     # 供應商適配層（OpenAI/Anthropic/Gemini）
│   ├── commands/                    # CLI 子命令（migrate-keys）
│   ├── config/                      # 配置管理（SystemSettings + Manager）
│   ├── container/                   # dig 依賴注入容器定義
│   ├── db/                          # 資料庫連線 + 遷移
│   │   └── migrations/              # 版本遷移腳本
│   ├── encryption/                  # 密鑰加密服務
│   ├── errors/                      # 錯誤分類（可忽略/不計入/重試）
│   ├── handler/                     # HTTP Handler 層（Gin）
│   ├── httpclient/                  # HTTP 客戶端管理
│   ├── i18n/                        # 國際化（zh-CN/en）
│   ├── keypool/                     # 密鑰池管理（輪換/驗證/Cron）
│   ├── middleware/                  # Gin 中間件（Auth/CORS/RateLimit/Logger）
│   ├── models/                      # GORM 資料模型
│   ├── proxy/                       # 代理核心（請求轉發/串流/重試）
│   ├── response/                    # 統一回應格式 + 分頁
│   ├── router/                      # 路由定義
│   ├── services/                    # 業務服務層
│   ├── store/                       # KV 存儲抽象（Memory/Redis）
│   ├── syncer/                      # 集群快取同步（Redis Pub/Sub）
│   ├── types/                       # 介面與型別定義
│   ├── utils/                       # 工具函數
│   └── version/                     # 版本資訊
└── web/                             # Vue 3 前端（Vite + TypeScript）
    └── src/
        ├── api/                     # API 呼叫層
        ├── components/              # UI 元件
        ├── views/                   # 頁面視圖
        ├── locales/                  # i18n 語言檔
        ├── router/                   # 前端路由
        ├── services/                 # 前端服務
        └── types/                    # TypeScript 型別
```

---

## 常用命令

```bash
make run          # 建置前端 + 啟動服務
make dev          # 開發模式（race detection）
make migrate-keys ARGS="--from old --to new"  # 密鑰遷移
```

服務透過 `scripts/services.sh` 以 screen 啟動：`START_GPTLOAD=1`

---

## 核心架構

### 依賴注入

所有服務透過 `internal/container/container.go` 的 `dig` 容器註冊，`App` 結構體透過 `AppParams` 接收全部依賴。

### 請求流程

```
Client → Gin Router → ProxyAuth → ProxyServer.HandleProxy → Channel（OpenAI/Anthropic/Gemini）→ 回應處理/串流
                                                                                ↓
                                                          RequestLogService.Record → 快取 → 定期 flush → PostgreSQL
```

### Master / Slave 模式

- **Master** (`IS_SLAVE=false`): 執行資料庫遷移、設定初始化、密鑰載入至快取、日誌清理服務、Cron 密鑰驗證、日誌 flush
- **Slave** (`IS_SLAVE=true`): 僅啟動代理服務，從 Redis 同步設定

### 日誌系統

1. **寫入流程**: 請求完成 → `RequestLogService.Record()` → 寫入記憶體/Redis 快取 → 每 `request_log_write_interval_minutes` 分鐘 flush 至資料庫
2. **當設定為 0 時**: 同步寫入模式，跳過快取直接寫入資料庫
3. **清理服務**: `LogCleanupService` 每 **2 小時**執行一次，刪除 `timestamp < now() - retention_days` 的記錄，僅 Master 節點啟動
4. **保留天數**: 由系統設定 `request_log_retention_days` 控制（當前值: 8）

### 資料模型

| 模型 | 表名 | 說明 |
|------|------|------|
| `SystemSetting` | system_settings | 鍵值對系統設定 |
| `Group` | groups | 供應商分組（standard/aggregate） |
| `GroupSubGroup` | group_sub_groups | 聚合分組 <-> 子分組關聯 |
| `APIKey` | api_keys | API 密鑰（含狀態/計數/加密） |
| `RequestLog` | request_logs | 請求日誌（含 Token 統計/回應體） |
| `GroupHourlyStat` | group_hourly_stats | 分組小時維度統計 |

---

## API 路由

| 路由 | 說明 |
|------|------|
| `POST /api/auth/login` | 登入（公開） |
| `GET /api/groups` | 列出分組（需 Auth） |
| `POST /api/groups` | 建立分組 |
| `GET /api/keys` | 列出密鑰 |
| `POST /api/keys/add-multiple` | 批量新增密鑰 |
| `GET /api/dashboard/stats` | 儀表板統計 |
| `GET /api/logs` | 查詢日誌（分頁/篩選） |
| `GET /api/logs/export` | 匯出日誌密鑰 CSV |
| `GET /api/settings` | 取得系統設定 |
| `PUT /api/settings` | 更新系統設定 |
| `ANY /proxy/:group_name/*path` | 代理轉發（Auth 由分組的 proxy_keys 控制） |
| `GET /health` | 健康檢查 |

---

## 環境變數（.env）

| 變數 | 說明 | 當前值 |
|------|------|--------|
| `DATABASE_DSN` | 資料庫連線字串 | PostgreSQL 192.168.77.140:50432 |
| `REDIS_DSN` | Redis 連線（空則用記憶體） | 未設定 |
| `AUTH_KEY` | 管理 API 認證金鑰 | ✅ 已設定 |
| `ENCRYPTION_KEY` | API Key 靜態加密金鑰 | ✅ 已設定 |
| `IS_SLAVE` | 是否為 Slave 節點 | false |
| `PORT` / `HOST` | 服務端口/監聽地址 | 47300 / 0.0.0.0 |

---

## 開發注意事項

- **修改模型後**: GORM AutoMigrate 僅加欄位，不刪除/改型別；遷移腳本放在 `internal/db/migrations/`
- **系統設定**: 新增設定項需在 `types.SystemSettings` 結構體加欄位（含 json/default/name/desc/validate tag），初始化時自動寫入資料庫
- **新增供應商**: 實作 `channel.Channel` 介面，並在 `channel.Factory` 註冊
- **日誌清理**: 僅在 Master 節點啟動；`retention_days ≤ 0` 時停止清理
- **前端建置**: `make run` 會自動 `npm install && npm run build`，版本號取自 `internal/version/version.go`
- **前端開發**: 修改 `web/src` 後需重新 `npm run build`，建置產物透過 `//go:embed` 嵌入二進位