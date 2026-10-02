---
name: go-ui-app
description: >-
  Use this skill when developing, packaging, or troubleshooting Go applications that embed a Web UI,
  integrate Monaco Editor, or provide local developer tooling.
---

# Go Embedded Web UI Application Skill

Skill này hướng dẫn quy trình phát triển và đóng gói ứng dụng Go có nhúng giao diện Web (Single-Page Application).

## 1. Cơ chế nhúng tài nguyên (Embedded Assets)

Sử dụng thư viện chuẩn của Go `embed.FS`:
```go
package web

import "embed"

//go:embed all:dist
var Assets embed.FS
```

Lợi ích:
- Toàn bộ HTML, CSS, JavaScript, icons, Monaco Editor assets được nén gọn trong file binary duy nhất.
- Người dùng chỉ cần tải 1 file `.exe` hoặc nhị phân macOS/Linux và chạy ngay, không cần cài đặt Node.js hay web server phụ.

## 2. Tích hợp Monaco Editor cho Elasticsearch

1. **Khởi tạo Editor**:
   Sử dụng Monaco Editor (phiên bản CDN hoặc nhúng bundle) hỗ trợ cấu hình theme `vs-dark`.
2. **Đăng ký Autocomplete Provider**:
   Sử dụng `monaco.languages.registerCompletionItemProvider('json', ...)`:
   - Cung cấp gợi ý từ khóa Elastic DSL khi gõ trong block JSON.
   - Cung cấp gợi ý field names dựa vào index được chọn.
3. **Phím tắt**:
   - `editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, runQueryCallback)`: Thực thi query ngay tức thì.
   - `editor.getAction('editor.action.formatDocument').run()`: Tự động căn chỉnh JSON đẹp mắt.

## 3. Quản lý trạng thái và kết nối
- Tự động mở trình duyệt mặc định khi server khởi động:
  - macOS: `exec.Command("open", url).Start()`
  - Linux: `exec.Command("xdg-open", url).Start()`
  - Windows: `exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()`
- Cung cấp cờ dòng lệnh `--no-browser` cho các trường hợp chạy trên server/docker.
