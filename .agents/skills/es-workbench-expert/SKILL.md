---
name: es-workbench-expert
description: >-
  Use this skill when interacting with Elasticsearch clusters (7.x, 8.x, OpenSearch),
  extracting index mappings, generating query DSL snippets, or optimizing search performance.
---

# Elasticsearch Workbench Expert Skill

Skill này cung cấp kiến thức nền tảng và các bước chuẩn hóa để tương tác, tối ưu hóa và phân tích Elasticsearch trong ESKhan.

## 1. Tương thích các phiên bản Elasticsearch (7.x, 8.x, OpenSearch)

### Điểm khác biệt cần lưu ý:
- **Elasticsearch 7.x**:
  - Không còn mapping types (type mặc định là `_doc`).
  - Response field `hits.total`: Có thể là số nguyên nếu không bật `track_total_hits`, hoặc object `{"value": 10000, "relation": "gte"}`. Parser cần fallback parse cả hai trường hợp.
- **Elasticsearch 8.x**:
  - Bảo mật (Security/HTTPS) được bật mặc định khi cài đặt. Cần hỗ trợ chứng chỉ tự ký (`InsecureSkipVerify`) và API Key.
  - Endpoint `_cat/*` trả về định dạng text hoặc JSON (`?format=json`).
- **OpenSearch**:
  - Tương thích hoàn toàn với các REST API cơ bản của ES 7.10.2. Header version có thể chứa `opensearch`.

## 2. Quy trình phân tích Mapping để sinh Autocomplete

Để sinh autocomplete trường dữ liệu (field names) chính xác cho Monaco Editor:
1. Gọi API `GET /{index}/_mapping`
2. Đệ quy qua object `properties`:
   - Nếu một field có type (`keyword`, `text`, `long`, `boolean`, `date`, `nested`), ghi nhận vào danh sách gợi ý kèm type.
   - Nếu một field có sub-properties (ví dụ nested object), tạo đường dẫn dot-notation (ví dụ `user.address.city`).
   - Ghi nhận `fields` con (ví dụ `title.keyword`).
3. Xuất ra định dạng JSON cho Autocomplete Provider của Monaco Editor:
   ```json
   [
     {"label": "user.id", "type": "keyword", "detail": "Keyword field"},
     {"label": "user.email", "type": "keyword", "detail": "Keyword field"},
     {"label": "created_at", "type": "date", "detail": "Date field"}
   ]
   ```

## 3. Các mẫu Query DSL chuẩn (Snippets)

### Boolean Search Template:
```json
{
  "query": {
    "bool": {
      "must": [
        { "match": { "field_name": "query_text" } }
      ],
      "filter": [
        { "term": { "status": "active" } },
        { "range": { "created_at": { "gte": "now-7d/d" } } }
      ]
    }
  },
  "sort": [
    { "created_at": { "order": "desc" } }
  ],
  "size": 20
}
```

### Aggregations Template:
```json
{
  "size": 0,
  "aggs": {
    "by_status": {
      "terms": {
        "field": "status.keyword",
        "size": 10
      },
      "aggs": {
        "avg_price": {
          "avg": { "field": "price" }
        }
      }
    }
  }
}
```
