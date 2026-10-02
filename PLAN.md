# Kế hoạch phát triển ESKhan (Elasticsearch & Universal API IDE)

Dự án xây dựng công cụ giao diện trực quan tất-cả-trong-một cho Developer viết bằng Golang:
- **Elasticsearch Workbench**: Tương thích ES 7.x, 8.x, OpenSearch, smart autocomplete, visual charts, linter.
- **gRPC Studio**: Dynamic protobuf invoker, Server Reflection, proto import, siêu nhẹ & mượt.
- **REST & Swagger Studio**: Postman-like client, proxy Go bypass CORS, import Swagger URL / OpenAPI v2-v3.

---

## 🌟 Lộ trình Version 1 (ĐÃ HOÀN THÀNH ✅)
- [x] **Giai đoạn 1**: Thiết lập Skill & Quy chuẩn làm việc ([GEMINI.md](GEMINI.md)).
- [x] **Giai đoạn 2**: Backend Golang Core & Giao diện IDE Cơ bản (ES 7.x/8.x/OpenSearch, Monaco Editor, CSV export).
- [x] **Giai đoạn 3**: Tính năng Premium (Linter, Context-Aware Autocomplete, Top Terms, Visual Charts, Analyzer Playground, Document CRUD, Safe Mode, Variables, Index Maintenance).

---

## 🚀 Lộ trình Version 2 (Universal Developer Studio)

### 📌 Giai đoạn 1 (ƯU TIÊN HÀNG ĐẦU): gRPC Client Engine & Dynamic Invoker (ĐÃ HOÀN THÀNH ✅)
*Mục tiêu: Gọi gRPC siêu mượt bằng Go core, không cần cài protoc, không cần BloomRPC/Postman.*
- [x] **1.1. gRPC Reflection & Discovery (`internal/grpcclient`)**:
  - Hỗ trợ gRPC Server Reflection Protocol (`v1` và `v1alpha`).
  - Tự động scan và liệt kê toàn bộ Packages, Services, Methods (RPCs).
  - Trích xuất Protobuf Descriptors của Request/Response messages.
- [x] **1.2. Dynamic Protobuf & Mock JSON Generator**:
  - Tự động sinh JSON mẫu (Mock payload) từ Message Descriptor (chuẩn hóa types: string, number, bool, enum, nested message, repeated array).
  - Parser 2 chiều: JSON (từ Monaco Editor) ➔ Protobuf Binary (để gửi đi) và Protobuf Binary ➔ JSON (để hiển thị).
- [x] **1.3. Dynamic Invoker & Connection Manager**:
  - Hỗ trợ kết nối **Plaintext (Insecure)** và **TLS** (kèm cờ `InsecureSkipVerify`).
  - Gửi gRPC Metadata (Headers, Authorization Bearer, Custom keys).
  - Thực thi **Unary RPC** với timeout và context cancellation.
  - Đo thời gian thực thi (Latency / Took ms).
- [x] **1.4. Proto File Import (Fallback khi server tắt reflection)**:
  - Cho phép người dùng upload / nạp nội dung file `.proto` để parse descriptors động (`internal/grpcclient/proto_parser.go`).
- [x] **1.5. REST API Endpoints Backend**:
  - `POST /api/grpc/reflect`: Khám phá services & methods từ gRPC target.
  - `POST /api/grpc/invoke`: Gọi RPC method động với JSON body và metadata.
  - `POST /api/grpc/proto/parse`: Parse trực tiếp chuỗi `.proto` sang services & mock payloads.
- [x] **1.6. Giao diện gRPC Studio trên Web**:
  - Mode Switcher trên Header: **⚡ ES Workbench** ⟷ **🔌 gRPC Studio**.
  - Thanh chọn Target (host:port, Plaintext, Insecure TLS), Service & Method selector.
  - Tabs Request: Monaco JSON Request Editor, Metadata Key-Value Editor, Settings (timeout ms).
  - Tabs Response: Monaco Read-Only JSON Viewer, Headers & Trailers viewer, Status Code Badge, Latency & Payload size.
  - Phím tắt `Cmd+Enter` / `Ctrl+Enter` để Invoke RPC; `Cmd+Shift+F` để format JSON.
  - Modal Import `.proto` linh hoạt.
- [x] **1.7. Automated Unit Tests & Mock gRPC Server**:
  - Kiểm thử Server Reflection, Unary Invocation, Input Validation, và Proto Parser (`internal/grpcclient/client_test.go` & `internal/api/grpc_handlers_test.go`).

---

### 📌 Giai đoạn 2: REST HTTP Client Engine & Proxy Backend
*Mục tiêu: Động cơ gửi HTTP request đa năng, giải quyết 100% lỗi CORS của trình duyệt.*
- [ ] Package [`internal/httpclient`](internal/httpclient): Đầy đủ methods (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`, `OPTIONS`).
- [ ] Headers, Query Params, Request Body (JSON, Form-data, x-www-form-urlencoded, Raw).
- [ ] Đo đạc chi tiết chỉ số mạng qua `net/http/httptrace` (DNS, TCP, TLS, TTFB, Total Duration).
- [ ] API endpoint: `POST /api/http/send`.
- [ ] Unit test cho HTTP engine.

---

### 📌 Giai đoạn 3: OpenAPI & Swagger Parser & Explorer
*Mục tiêu: 1-click test API từ tài liệu Swagger.*
- [ ] Package [`internal/openapi`](internal/openapi): Fetch URL Swagger hoặc parse file JSON/YAML (hỗ trợ Swagger 2.0, OpenAPI 3.0/3.1).
- [ ] Tự động trích xuất endpoints, tags, parameters, và tự động tạo Mock JSON Body từ Schema.
- [ ] UI Swagger Explorer: Cây danh mục API, tìm kiếm, nút "⚡ Open in Tab" để đẩy sang API Client.
- [ ] Unit test cho Swagger/OpenAPI parser.

---

### 📌 Giai đoạn 4: Giao diện IDE Hợp nhất (Multi-Protocol Studio)
*Mục tiêu: Trải nghiệm mượt mà, chuyển đổi tức thì giữa các phân hệ.*
- [ ] Sidebar Mode Switcher: 🔍 **Elasticsearch** | 🔌 **gRPC Studio** | 🌐 **REST/Swagger**.
- [ ] Multi-tab engine: Mở đồng thời các Tab ES Query, Tab gRPC Call, Tab REST Request.
- [ ] Tối ưu hóa UI Dark Mode, phím tắt thống nhất (`Cmd+Enter` / `Ctrl+Enter` để thực thi).

---

### 📌 Giai đoạn 5: Collections, Biến môi trường & Đóng gói Phân phối
*Mục tiêu: Hoàn thiện tính năng tiện ích hàng ngày và phát hành.*
- [ ] Lưu trữ Collections & History chung cho cả ES, gRPC và REST.
- [ ] Environment Manager (`{{base_url}}`, `{{grpc_host}}`, `{{token}}`, `{{$timestamp}}`, `{{$uuid}}`).
- [ ] Chạy toàn bộ test suite (`go test ./...`) và build cross-platform binaries.
