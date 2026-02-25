# PROJECT_MAP.md — GPT-Load 專案導覽

> 快速定位專案結構，詳細架構請參閱 [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)

## 專案概述

高效能 AI API 透明代理服務，支援多供應商密鑰輪換、負載均衡與集群部署。

## 核心目錄

```
main.go                 # 程式入口
internal/
├── app/        # 應用生命週期
├── proxy/      # 代理核心（重點）
├── channel/    # 供應商適配層
├── keypool/    # 密鑰池管理
├── services/   # 業務服務層
├── handler/    # HTTP 處理器
├── middleware/ # 中間件
├── models/     # 資料模型
├── store/      # 緩存存儲
└── config/     # 配置管理
web/            # Vue 3 前端
```

## 常用命令

```bash
make run     # 啟動服務
make dev     # 開發模式
docker compose up -d
```