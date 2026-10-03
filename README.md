# ⚡ ESKhan - All-in-One Developer Workbench & API IDE (v2.0)

[![Release](https://img.shields.io/badge/version-v2.0.0-blue.svg)](https://github.com/Khan9Tran/eskhan)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8.svg)](https://golang.org)
[![Author](https://img.shields.io/badge/Author-Khan9Tran-green.svg)](https://github.com/Khan9Tran)

**ESKhan** là công cụ IDE & Workbench trực quan, siêu nhẹ và đa giao thức dành cho Developer viết bằng **Golang**, tích hợp 3 phân hệ cốt lõi trong một phần mềm duy nhất:
1. **⚡ Elasticsearch Workbench**: Tương thích hoàn hảo ES 7.x, 8.x và OpenSearch, Context-Aware Autocomplete, Top Terms, Visual Charts, Query Linter, Safe Mode.
2. **🔌 gRPC Studio**: Trình gọi RPC động (Dynamic Protobuf Invoker), tự động khám phá dịch vụ qua Server Reflection (`v1` và `v1alpha`), nạp file `.proto` dự phòng, sinh mock JSON tự động, gRPC metadata & latency.
3. **🌐 REST & Swagger Studio**: Giải pháp thay thế Postman siêu nhẹ và mượt mà, Go Backend Proxy giải quyết 100% rào cản CORS, đo đạc chi tiết mạng (`httptrace`), tích hợp Swagger / OpenAPI Explorer (v2/v3/v3.1) và quản lý Collections.

Toàn bộ ứng dụng được đóng gói vào **1 file nhị phân duy nhất** (~16MB, nhúng sẵn toàn bộ Web UI, Monaco Editor bằng `go:embed`). Khi khởi động, ESKhan sẽ tự động mở trình duyệt mặc định trên máy của bạn.

---

## 🌟 Tính năng Version 2.0 (Universal API Studio)

### 🔌 1. gRPC Studio (Dynamic RPC Invoker & Server Reflection)
- **Tự động quét Dịch vụ (Reflection Discovery)**:
  - Hỗ trợ gRPC Server Reflection Protocol chuẩn `v1` và `v1alpha`.
  - Tự động liệt kê tất cả Packages, Services và Methods (RPC).
  - Tự động sinh JSON mẫu (Mock Request) chuẩn xác theo kiểu dữ liệu Protobuf (scalar types, nested messages, repeated arrays, enums, maps).
- **Gọi RPC động không cần biên dịch stubs (Dynamic Protobuf Invoker)**:
  - Hỗ trợ cả kết nối **Plaintext (Insecure)** và **TLS** (kèm cờ `InsecureSkipVerify`).
  - Gửi gRPC Metadata (Headers, Authorization Bearer, Custom keys).
  - Đo thời gian thực thi (Latency / Took ms) và hiển thị gRPC Status Code (OK, NOT_FOUND, v.v.).
  - Monaco JSON Editor cho Request và Monaco Read-Only Viewer cho Response.
- **Nạp file `.proto` thủ công (Fallback)**:
  - Khi server gRPC tắt Reflection, bạn có thể dán trực tiếp nội dung file `.proto` để phân tích và sinh mock JSON ngay lập tức.

### 🌐 2. REST & Swagger Studio (Postman Alternative & OpenAPI Explorer)
- **Động cơ HTTP Client cực nhẹ & Proxy Go bypass CORS**:
  - Không gặp bất kỳ giới hạn CORS nào của trình duyệt.
  - Hỗ trợ đầy đủ phương thức: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`, `OPTIONS`.
  - Quản lý Parameters (Query Params) và Headers dạng bảng Key-Value linh hoạt.
  - Hỗ trợ nhiều loại Request Body: None, JSON (Monaco Editor với auto-format `Cmd+Shift+F`), x-www-form-urlencoded, multipart form-data, và Raw text.
- **Đo lường chi tiết chỉ số mạng qua `net/http/httptrace`**:
  - Trực quan hóa từng giai đoạn mạng dưới dạng thanh tiến trình (Progress Bars): **DNS Lookup**, **TCP Connect**, **TLS Handshake**, **TTFB (Time to First Byte)**, và **Total Duration**.
- **Swagger / OpenAPI Explorer (v2.0, v3.0, v3.1)**:
  - Nạp từ đường dẫn URL (ví dụ: `/v3/api-docs` hoặc `/swagger.json`) hoặc dán trực tiếp JSON/YAML.
  - Cây danh mục API phân theo Tag, tìm kiếm tức thì theo keyword / đường dẫn.
  - Tự động trích xuất Parameters và Schema payload ($ref resolver), bấm **1-click để nạp thẳng vào Request Builder** và gửi thử nghiệm ngay.

### 🌍 3. Quản lý Môi trường (Environments) & Biến số động
- **Environment Switcher**:
  - Chuyển đổi môi trường làm việc trực tiếp ngay trên thanh Header (ví dụ: `Local Dev`, `Staging`, `Production`, `No Environment`).
  - Bảng quản lý biến số: Khai báo cặp Key - Value (ví dụ: `base_url`, `grpc_target`, `auth_token`).
- **Nạp biến số linh hoạt trên toàn hệ thống**:
  - Sử dụng cú pháp `{{ten_bien}}` trong URL, Header, Query Param, Metadata, hoặc Body.
- **Hệ thống biến động tích hợp sẵn (Built-in Dynamic Variables)**:
  - `{{$timestamp}}`: Unix epoch (giây).
  - `{{$timestamp_ms}}`: Unix epoch (mili-giây).
  - `{{$uuid}}`: Chuỗi UUID ngẫu nhiên v4.
  - `{{$date}}`: Ngày hiện tại `YYYY-MM-DD`.
  - `{{$datetime}}`: Thời gian hiện tại chuẩn ISO8601 / RFC3339.
  - `{{$randomInt}}`: Số nguyên ngẫu nhiên 6 chữ số (100000 - 999999).
- **Monaco Autocomplete**: Gõ `{{` trong trình soạn thảo sẽ tự động gợi ý danh sách biến môi trường và biến hệ thống!

### 📂 4. Saved Collections & Lịch sử gọi đa giao thức
- Nút **💾 Save** trên thanh Request Builder và gRPC Studio cho phép lưu các lệnh gọi vào từng thư mục Collection.
- Phân mục Collections ngay trên Sidebar: Xem danh sách, nhấp vào để nạp lại đầy đủ URL, headers, params và body.
- Lịch sử truy vấn riêng biệt cho từng phân hệ (ES Query History, HTTP History, gRPC History) kèm mã trạng thái HTTP / gRPC status code và thời gian thực thi.

---

## ⚡ Các tính năng Elasticsearch Workbench sẵn có

- **Context-Aware Autocomplete**: Gợi ý thông minh trường theo kiểu dữ liệu (`range` -> date/number, `terms/aggs` -> keyword, `match` -> text).
- **Top Terms Discovery**: Gợi ý giá trị thực tế phổ biến nhất trong Index.
- **Smart Query Linter**: Cảnh báo rủi ro hiệu năng tĩnh (Leading Wildcard, Must vs Filter, Deep Pagination, Text Field Aggs).
- **Visual Aggregations Chart**: Vẽ biểu đồ cột trực quan từ kết quả truy vấn.
- **Analyzer Playground**: Thử nghiệm tokenizer và analyzer tiếng Việt / quốc tế.
- **Document CRUD & Safe Mode**: Xem, chỉnh sửa, xóa tài liệu ES với chế độ bảo vệ chống xóa nhầm trên Production.

---

## 🚀 Khởi chạy ứng dụng

### 1. Tải và chạy file nhị phân (Standalone Binary):
```bash
# macOS Apple Silicon
./bin/eskhan-darwin-arm64

# macOS Intel
./bin/eskhan-darwin-amd64

# Linux
./bin/eskhan-linux-amd64

# Windows
./bin/eskhan-windows-amd64.exe
```
Ứng dụng sẽ tự động mở trình duyệt tại `http://localhost:8989`!

### 2. Chạy từ mã nguồn Go:
```bash
go run ./cmd/eskhan
```

### 3. Tùy chọn CLI:
```bash
./eskhan --help
  --port int          Cổng HTTP (mặc định: 8989)
  --host string       Địa chỉ IP bind (mặc định: localhost)
  --no-browser        Không tự động mở trình duyệt (thích hợp cho server / docker)
  --config string     Đường dẫn file cấu hình tùy chỉnh
  --version           In thông tin phiên bản
```

### 4. Đóng gói cho tất cả nền tảng:
```bash
./build_all.sh v2.0.0
```

---

## 🧪 Kiểm thử tự động (Unit Tests)

Toàn bộ backend được kiểm thử toàn diện:
```bash
go test -v ./...
```
Tất cả các packages `grpcclient`, `httpclient`, `openapi`, `variable`, `config`, `storage`, `es`, `linter`, `schema`, `converter`, `api` đều đạt **100% PASS**.

---

## 📜 Bản quyền (License) & Tác giả

- **Tác giả**: [Khan9Tran](https://github.com/Khan9Tran)
- Giấy phép mã nguồn mở: **[MIT License](LICENSE)**.


