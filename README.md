# ⚡ ESKhan - Elasticsearch Pro Query IDE

[![Release](https://img.shields.io/badge/version-v1.0.0-blue.svg)](https://github.com/Khan9Tran/eskhan)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8.svg)](https://golang.org)
[![Author](https://img.shields.io/badge/Author-Khan9Tran-green.svg)](https://github.com/Khan9Tran)

**ESKhan** là một công cụ IDE/Workbench trực quan, tốc độ cao được viết bằng **Golang**, chuyên dụng để truy vấn và làm việc với **Elasticsearch (tương thích hoàn hảo 7.x, 8.x và OpenSearch)**.

Toàn bộ ứng dụng được đóng gói vào **1 file nhị phân duy nhất** (nhúng sẵn toàn bộ Web UI, Monaco Editor bằng `go:embed`). Khi khởi động, ESKhan sẽ tự động mở trình duyệt mặc định trên máy của bạn.

---

## 🌟 Bộ tính năng Pro & Thông minh (Premium Grade)

### 1. Trình soạn thảo Monaco Editor thông minh (Smart Intellisense)
- **Context-Aware Autocomplete**:
  - Khi gõ trong mệnh đề `"range"`: Hệ thống tự động ưu tiên gợi ý các trường kiểu ngày (`date`) và số (`long`, `integer`, `float`).
  - Khi gõ trong mệnh đề `"terms"` hoặc `"aggs"`: Ưu tiên gợi ý các trường `keyword`.
  - Khi gõ trong mệnh đề `"match"`: Ưu tiên gợi ý các trường `text`.
- **Top Terms Discovery**: Tự động lấy 8-10 giá trị phổ biến nhất trong index để gợi ý giá trị khi bạn gõ điều kiện lọc cho trường keyword (ví dụ: gõ `status` -> gợi ý `"active"`, `"pending"`).
- **Hỗ trợ biến môi trường & tham số hóa**:
  - `{{$timestamp}}`: Unix timestamp hiện tại.
  - `{{$date}}`: Ngày hiện tại dạng YYYY-MM-DD.
  - `{{$uuid}}`: UUID ngẫu nhiên.
  - `{{custom_var}}`: Biến tùy chỉnh do người dùng định nghĩa.
- **Phím tắt**:
  - `⌘ + Enter` (hoặc `Ctrl + Enter`): Chạy truy vấn tức thì.
  - `⌘ + Shift + F` (hoặc `Ctrl + Shift + F`): Format / Prettify JSON.
- **Multi-Tab**: Mở nhiều tab truy vấn song song.

### 2. Smart Query Linter & Anti-Pattern Detection
- Tự động phân tích cú pháp tĩnh khi bạn gõ code và cảnh báo ngay lập tức:
  - **Leading Wildcard (`*abc`)**: Cảnh báo rủi ro quét toàn bộ index và đề xuất giải pháp.
  - **Must vs Filter**: Đề xuất chuyển các truy vấn chính xác (term, range) sang `filter` để tận dụng ES Node Query Cache.
  - **Deep Pagination / Large Size**: Cảnh báo khi `size > 1000` hoặc `from + size > 10000` và hướng dẫn dùng `search_after`.
  - **Text Field Aggregation**: Cảnh báo khi sort/agg trên trường `text` và gợi ý dùng `.keyword`.
- Nút cảnh báo **⚠️ X Tips** trên thanh công cụ cho phép mở bảng phân tích chi tiết.

### 3. Trực quan hóa Aggregations (Visual Charts)
- Khi kết quả trả về có `aggregations`: Tự động kích hoạt tab **📊 Aggregations Chart**.
- Vẽ biểu đồ cột (Interactive Bar Chart) trực quan cho các bucket `terms`.
- Hiển thị tỷ lệ phần trăm (%), số lượng tài liệu (`doc_count`), và cho phép click vào cột để lọc tiếp.

### 4. Text Analyzer & Tokenizer Playground (`_analyze`)
- Tab **🔬 Analyzer** trên Sidebar cho phép bạn kiểm tra cơ chế bóc tách từ:
  - Nhập văn bản mẫu (hỗ trợ kiểm tra tiếng Việt, tiếng Anh, v.v.).
  - Chọn Analyzer (`standard`, `whitespace`, `simple`, hoặc analyzer tiếng Việt tùy chỉnh).
  - Hiển thị trực quan luồng Tokens (Tokens Ribbon) kèm vị trí `position`, `start_offset`, `end_offset` và `type`.

### 5. Document CRUD & Inspector
- Trong **Table View**, click nút **👁️ Inspect** trên từng dòng để mở Document Drawer.
- Xem chi tiết toàn bộ các trường `_source`.
- Cho phép chỉnh sửa JSON và bấm **💾 Save Document Changes** (`POST /{index}/_update/{id}`).
- Nút **🗑️ Delete Document** (`DELETE /{index}/_doc/{id}`) kèm hộp thoại xác nhận.
- **Chế độ Safe Mode (Read-Only Toggle)** trên Header: Khi bật Safe Mode, tính năng sửa/xóa document và các thao tác bảo trì phá hủy sẽ bị khóa chặt để bảo vệ cụm Production.

### 6. Quản lý Index & Thao tác nhanh (Maintenance Hub)
- Nút tác vụ nhanh trên từng index card ở Sidebar:
  - `🔄 Refresh Index` (`POST /{index}/_refresh`)
  - `🧹 Clear Query Cache` (`POST /{index}/_cache/clear`)
  - `💾 Flush Index` (`POST /{index}/_flush`)

### 7. Hiển thị kết quả đa chiều & Xuất dữ liệu
- **JSON Tree View**: Đóng/mở node, copy JSON path.
- **Table Grid View**: Biến đổi mảng documents thành bảng dữ liệu, hỗ trợ xuất file **CSV**.
- **Raw JSON View**: Xem trực tiếp response gốc.
- **Metrics Bar**: Đo thời gian `took` của ES, độ trễ mạng `latency`, số lượng hits và shard info.

---

## 🚀 Khởi chạy ứng dụng

### Chạy trực tiếp từ file nhị phân đã build sẵn:
```bash
./eskhan
```
Ứng dụng sẽ khởi động tại `http://localhost:8989` và tự động mở trình duyệt mặc định của bạn!

### Hoặc chạy từ mã nguồn Go:
```bash
go run ./cmd/eskhan
```

### Các tùy chọn dòng lệnh (CLI Flags):
```bash
./eskhan --help
  --port int          Cổng HTTP (mặc định: 8989)
  --host string       Địa chỉ IP bind (mặc định: localhost)
  --no-browser        Không tự động mở trình duyệt (thích hợp khi chạy trên server/docker)
  --config string     Đường dẫn file cấu hình tùy chỉnh
  --version           In thông tin phiên bản
```

---

## 🧪 Kiểm thử tự động (Unit Tests)

Dự án tuân thủ tiêu chuẩn kiểm thử nghiêm ngặt:
```bash
go test -v ./...
```
Toàn bộ các package `config`, `es`, `schema`, `storage`, `api`, `linter` đều đạt 100% test pass.

---

## 📂 Cấu trúc dự án

```text
ESKhan9/
├── GEMINI.md                                  # Quy chuẩn làm việc, persona và tiêu chuẩn kỹ thuật
├── PLAN.md                                    # Kế hoạch phát triển chi tiết
├── README.md                                  # Tài liệu hướng dẫn sử dụng
├── .agents/skills/
│   ├── es-workbench-expert/SKILL.md          # Skill phân tích Elasticsearch & Mapping
│   └── go-ui-app/SKILL.md                    # Skill phát triển Go Embedded Web UI
├── cmd/
│   └── eskhan/main.go                         # Điểm khởi chạy ứng dụng (CLI, auto-browser, shutdown)
├── internal/
│   ├── config/                                # Quản lý profile cluster, preferences
│   ├── es/                                    # Client ES 7.x/8.x, CRUD, analyze, top terms, variables
│   ├── linter/                                # Query static linter & anti-pattern detection
│   ├── schema/                                # Mapping parser & context autocomplete engine
│   ├── storage/                               # Lưu trữ lịch sử query và snippets
│   ├── version/                               # Quản lý version, build metadata, commit info
│   └── api/                                   # REST API router & handlers
└── web/
    ├── embed.go                               # go:embed nén giao diện vào binary
    └── dist/                                  # Frontend (HTML/CSS/JS, Monaco Editor, Charts, Tokenizer)
```

---

## 📜 Bản quyền (License) & Tác giả

- **Tác giả**: [Khan9Tran](https://github.com/Khan9Tran)
- Dự án được phát hành theo giấy phép mã nguồn mở **[MIT License](LICENSE)**. Toàn quyền sử dụng, sửa đổi và đóng gói miễn phí cho mục đích cá nhân lẫn thương mại.

