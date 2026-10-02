# Kế hoạch phát triển ESKhan (Elasticsearch Query IDE)

Dự án xây dựng công cụ giao diện trực quan cho Elasticsearch viết bằng Golang, tập trung vào trải nghiệm lập trình viên (Developer Experience), tốc độ cao và hỗ trợ gợi ý thông minh (Smart Autocomplete).

---

## 1. Trọng tâm tương thích Elasticsearch
- **Phiên bản mục tiêu chính**: Elasticsearch 7.x.
- **Khả năng tương thích mở rộng**: Tương thích hoàn toàn với Elasticsearch 8.x và OpenSearch (hỗ trợ HTTPS/TLS tự ký, Basic Auth, API Key, parse linh hoạt response `hits.total` dạng số hoặc object).

---

## 2. Lộ trình triển khai (Roadmap)

### Giai đoạn 1: Thiết lập Skill & Quy chuẩn làm việc (ĐÃ HOÀN THÀNH ✅)
- [x] Tạo [GEMINI.md](GEMINI.md): Quy định vai trò Senior Architect, tiêu chuẩn code Go, quy chuẩn giao diện Dark Mode, và phương pháp kiểm thử.
- [x] Tạo Skill [.agents/skills/es-workbench-expert/SKILL.md](.agents/skills/es-workbench-expert/SKILL.md): Kỹ năng tương tác ES 7.x/8.x, trích xuất mapping và tối ưu DSL.
- [x] Tạo Skill [.agents/skills/go-ui-app/SKILL.md](.agents/skills/go-ui-app/SKILL.md): Kỹ năng nhúng Web UI, Monaco Editor, và phân phối Go binary.

### Giai đoạn 2: Xây dựng Backend Golang Core & Giao diện IDE Cơ bản (ĐÃ HOÀN THÀNH ✅)
- [x] Khởi tạo module `go.mod`
- [x] Cấu hình lưu trữ cục bộ: profile kết nối, lịch sử query, snippets (`internal/config`, `internal/storage`)
- [x] Client tương thích ES 7.x/8.x/OpenSearch (`internal/es`): Ping, health, raw query, metrics
- [x] Bộ phân tích Mapping & Autocomplete Schema (`internal/schema`)
- [x] REST API Router (`internal/api`)
- [x] Monaco Editor với autocomplete DSL, JSON Tree, Table view, CSV export, phím tắt `Cmd+Enter`
- [x] Đóng gói vào 1 file nhị phân duy nhất `eskhan`

### Giai đoạn 3: Nâng cấp tính năng Premium & Thông minh (ĐÃ HOÀN THÀNH ✅)
- [x] **Smart Query Linter & Anti-pattern Detector**: Phát hiện truy vấn chậm (wildcard ở đầu, must thay vì filter, size quá lớn, unindexed fielddata) kèm hướng dẫn tối ưu.
- [x] **Context-Aware Autocomplete & Top Terms**: Gợi ý trường theo ngữ cảnh (`range` chỉ hiện ngày/số, `terms` hiện keyword) và tự động nạp các giá trị mẫu thực tế từ index.
- [x] **Visual Aggregations Chart**: Tự động vẽ biểu đồ trực quan (Interactive Bar Chart) khi query có `aggregations`.
- [x] **Analyzer & Tokenizer Playground (`_analyze`)**: Trực quan hóa luồng phân tích từ (tokens ribbon, offsets, types) của các bộ phân tích tiếng Việt/tiếng Anh.
- [x] **Document CRUD & Inspector**: Xem chi tiết document dạng drawer, chỉnh sửa hoặc xóa document trực tiếp từ bảng kết quả.
- [x] **Safe Mode (Read-Only Toggle)**: Nút gạt an toàn bảo vệ dữ liệu Production khỏi việc sửa/xóa nhầm.
- [x] **Biến môi trường & Tham số hóa**: Hỗ trợ cú pháp `{{variable}}`, `{{$timestamp}}`, `{{$uuid}}`, `{{$date}}`.
- [x] **Index Maintenance Hub**: Các thao tác nhanh trên Sidebar (`_refresh`, `_flush`, `_cache/clear`).
