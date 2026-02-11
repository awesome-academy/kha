# Foods & Drinks API

Hệ thống quản lý đồ ăn và đồ uống với Golang.

## Công nghệ sử dụng

- **Language**: Go 1.21+
- **Database**: MySQL 8.0+
- **ORM**: GORM
- **Migration**: golang-migrate
- **Config**: Viper

## Cấu trúc Project

```
.
├── cmd/
│   ├── server/          # Main server application
│   ├── migrate/         # Database migration tool
│   └── seed/            # Database seeder
├── internal/
│   ├── config/          # Configuration management
│   ├── models/          # Database models
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic layer
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # HTTP middlewares
│   └── routes/          # Route definitions
├── pkg/
│   ├── database/        # Database connection
│   └── utils/           # Utility functions
├── migrations/          # SQL migration files
├── config.yaml          # Configuration file
├── config.example.yaml  # Example configuration
├── Makefile            # Build and run commands
└── go.mod              # Go module file
```

## Setup

### 1. Clone và cài đặt dependencies

```bash
git clone <repository-url>
cd kha
go mod download
```

### 2. Cấu hình

Copy file cấu hình mẫu và chỉnh sửa:

```bash
cp config.example.yaml config.yaml
```

Cập nhật thông tin database trong `config.yaml`:

```yaml
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "your-password"
  dbname: "foods_drinks"
```

### 3. Tạo Database

```sql
CREATE DATABASE foods_drinks CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 4. Chạy Migration

```bash
# Chạy tất cả migrations
make migrate-up

# Hoặc chạy trực tiếp
go run ./cmd/migrate -command=up
```

### 5. Chạy Server

```bash
make run

# Hoặc
go run ./cmd/server
```

## Migration Commands

```bash
# Chạy tất cả migrations
make migrate-up

# Rollback tất cả migrations
make migrate-down

# Xem version hiện tại
make migrate-version

# Chạy N migrations
make migrate-up-steps STEPS=1

# Rollback N migrations
make migrate-down-steps STEPS=1

# Force version (khi bị dirty)
make migrate-force VERSION=12
```

## Database Schema

Hệ thống bao gồm 12 bảng:

1. **users** - Quản lý người dùng
2. **social_auths** - Đăng nhập qua mạng xã hội
3. **categories** - Danh mục sản phẩm
4. **products** - Sản phẩm (food/drink)
5. **product_images** - Hình ảnh sản phẩm
6. **carts** - Giỏ hàng
7. **cart_items** - Sản phẩm trong giỏ hàng
8. **orders** - Đơn hàng
9. **order_items** - Chi tiết đơn hàng
10. **ratings** - Đánh giá sản phẩm
11. **suggestions** - Đề xuất sản phẩm mới
12. **order_notifications** - Thông báo đơn hàng

## License

MIT
