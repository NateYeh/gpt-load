# 📝 專案維護心得：gpt-load 版本同步與編譯優化

## 1. 專案背景
- **專案名稱**：`gpt-load` (Multi-channel AI proxy)
- **維護目標**：將 Fork 的專案從 `v1.0.0` 升級至官方最新的 `v1.4.4`，並解決 UI 介面版本號顯示不一致的問題。

## 2. 遇見的問題
雖然已經手動將後端程式碼的版本號修改為 `1.4.4`，但在執行專案後，網頁前端介面仍然顯示：
> **「v1.0.0 - 有更新 [v1.4.4]」**

### 問題分析：
1. **原始碼限制**：前端程式碼 (`web/src/services/version.ts`) 預設抓取環境變數 `VITE_VERSION`，若未定義則預設為 `1.0.0`。
2. **編譯脫節**：原本的 `Makefile` 在編譯前端時，並沒有將後端的版本資訊傳遞給 Vite 編譯器，導致前端打包後內容固定在舊版本。
3. **快取機制**：前端有版本資訊快取機制，若版本號不一致，會持續觸發與 GitHub API 的比對檢查。

## 3. 解決方案
我們採取的不是手動去修改每一個前端檔案，而是透過**「自動化注入」**的方式解決：

### A. 自動同步 Upstream
首先確保代碼與官方同步：
```bash
git remote add upstream https://github.com/tbphp/gpt-load.git
git fetch upstream
git merge upstream/main
```

### B. 優化編譯指令 (Makefile)
修改根目錄的 `Makefile`，讓前端在 `npm run build` 時能自動抓取 Go 專案定義的版本號。
- **修改前**：`cd web && npm run build`
- **修改後**：
  ```makefile
  # 使用 shell 指令動態抓取 internal/version/version.go 裡的值
  cd web && VITE_VERSION=$(shell grep -oP 'Version = "\134K[^"]+' internal/version/version.go) npm run build
  ```

## 4. 技術心得
1. **Single Source of Truth (唯一事實來源)**：
   在開發全棧專案時，版本號應該只在一個地方定義（例如此處的後端 Go 檔）。透過編譯指令將該值傳遞給前端，可以避免兩邊同步出錯。
   
2. **自動化勝於手動修改**：
   如果手動去修改 `package.json` 或前端組件，下次官方更新時又會產生衝突。透過修改 `Makefile` 或編譯腳本，是一勞永逸的做法。

3. **環境變數的靈活運用**：
   Vite 支援 `VITE_` 開頭的環境變數注入。利用這一特性，可以在不更動程式碼邏輯的前提下，動態改變前端部署後的行為。

## 5. 後續建議
- **清理快取**：若重啟後發現版本沒變，可能是瀏覽器 `LocalStorage` 快取了舊的版本檢查結果。可以開發者工具清理快取或點擊 UI 重新整理。
- **CI/CD 自動化**：未來可以用相似的邏輯，在 GitHub Actions 中自動完成版本比對與注入。

---
*文件更新日期：2026-02-24*
