# Task Breakdown

Mỗi task tối đa 8 giờ cho fresher Golang.

**Priority**: P0 → P1 → P2 → P3  
**Status**: TODO → IN_PROGRESS → REVIEW → DONE

---

## Phase 1: Setup & Foundation (P0)

### Task 1.1: Setup Project Structure & Database Migration
**Thời gian**: 6-8 giờ  
**Priority**: P0

**Mô tả**: Setup project structure và database migration.

**Subtasks**:
1. Setup project structure (cmd, internal, pkg, migrations)
2. Setup Go modules và dependencies (GORM, MySQL driver, viper)
3. Config file cho database
4. Migration scripts tạo 12 bảng
5. Seeder dữ liệu mẫu
6. Test migration up/down

**Acceptance**:
- Project structure rõ ràng
- 12 bảng được tạo thành công
- Migration up/down hoạt động
- Seeder tạo được dữ liệu mẫu

**Dependencies**: None

### Task 1.2: Setup HTTP Server & Basic Middleware
**Thời gian**: 4-6 giờ  
**Priority**: P0

**Mô tả**: Setup HTTP server và middleware cơ bản.

**Subtasks**:
1. Setup HTTP framework (Gin/Echo)
2. Middleware: CORS, Logger, Recovery
3. Routing structure (v1, admin, public)
4. Health check endpoint

**Acceptance**:
- Server chạy được, port configurable
- CORS, Logger, Recovery middleware hoạt động
- Health check endpoint OK

**Dependencies**: Task 1.1

## Phase 2: Authentication & Authorization (P0)

### Task 2.1: User Registration & Login (Email/Password)
**Thời gian**: 6-8 giờ  
**Priority**: P0

**Mô tả**: Register và Login với email/password.

**Subtasks**:
1. Models: User, SocialAuth
2. Repository layer cho User
3. Password hashing (bcrypt)
4. API Register/Login
5. Generate JWT token
6. Validation email/password

**Acceptance**:
- Register/Login hoạt động
- Password hash bằng bcrypt
- Validation email/password
- JWT token được trả về

**Dependencies**: Task 1.1, Task 1.2

---

### Task 2.2: JWT Authentication Middleware
**Thời gian**: 4-6 giờ  
**Priority**: P0

**Mô tả**: JWT authentication middleware và role-based authorization.

**Subtasks**:
1. Setup JWT library
2. JWT service: Generate, Validate, Extract claims
3. Auth middleware: Verify token, extract user info
4. Role middleware: Check admin/user
5. Helper: Get current user from context

**Acceptance**:
- JWT middleware verify token và extract user info
- Role middleware check admin/user
- Handle token expiration

**Dependencies**: Task 2.1

### Task 2.3: Social Authentication (OAuth)
**Thời gian**: 6-8 giờ  
**Priority**: P1

**Mô tả**: OAuth login với Facebook, Google, Twitter.

**Subtasks**:
1. Setup OAuth2 client (Facebook, Google, Twitter)
2. OAuth config
3. API initiate OAuth flow
4. OAuth callback handler
5. Auto create user nếu chưa tồn tại
6. Lưu social auth info, generate JWT

**Acceptance**:
- OAuth flow hoạt động cho Facebook, Google, Twitter
- Social auth info lưu vào DB
- Auto create user nếu chưa tồn tại
- Trả về JWT sau OAuth

**Dependencies**: Task 2.1, Task 2.2

## Phase 3: User Profile Management (P1)

### Task 3.1: User Profile APIs
**Thời gian**: 4-6 giờ  
**Priority**: P1

**Mô tả**: User profile APIs.

**Subtasks**:
1. API Get/Update Profile
2. API Upload Avatar
3. Validation phone/address
4. File upload handling

**Acceptance**:
- Get/Update profile hoạt động
- Upload avatar được
- Validation phone/address

**Dependencies**: Task 2.2

## Phase 4: Categories Management (P0 - Admin)

### Task 4.1: Categories CRUD APIs (Admin)
**Thời gian**: 6-8 giờ  
**Priority**: P0

**Mô tả**: Admin CRUD categories.

**Subtasks**:
1. Model Category, repository layer
2. CRUD APIs
3. Slug auto-generate
4. Pagination, filter by status
5. Soft delete

**Acceptance**:
- CRUD categories hoạt động
- Pagination và filter
- Slug auto-generate và unique

**Dependencies**: Task 2.2

## Phase 5: Products Management (P0)

### Task 5.1: Products CRUD APIs (Admin)
**Thời gian**: 6-8 giờ  
**Priority**: P0

**Mô tả**: Admin CRUD products.

**Subtasks**:
1. Models: Product, ProductImage
2. Repository layer
3. CRUD APIs với multiple images
4. Slug generation
5. Image upload và management
6. Soft delete

**Acceptance**:
- CRUD products với multiple images
- Upload và quản lý images
- Primary image được set đúng

**Dependencies**: Task 4.1, Task 2.2

---

### Task 5.2: Products Listing & Filtering APIs (Public)
**Thời gian**: 6-8 giờ  
**Priority**: P0

**Mô tả**: Public API list products với filter.

**Subtasks**:
1. Public API list products
2. Filter: classify, category, price, rating
3. Sort: price, rating, name, created_at
4. Full-text search
5. Pagination
6. Optimize query performance

**Acceptance**:
- Public API list products
- Filter: classify, category, price, rating
- Search full-text
- Sort và pagination

**Dependencies**: Task 5.1

### Task 5.3: Product Detail API (Public)
**Thời gian**: 3-4 giờ  
**Priority**: P0

**Mô tả**: Public API product detail.

**Subtasks**:
1. Public API product detail
2. Include images và category info
3. Error handling

**Acceptance**:
- Public API product detail
- Include images và category info

**Dependencies**: Task 5.1

## Phase 6: Cart Management (P0)

### Task 6.1: Cart APIs
**Thời gian**: 6-8 giờ  
**Priority**: P0

**Mô tả**: Cart management APIs.

**Subtasks**:
1. Models: Cart, CartItem
2. Repository layer
3. Auto-create cart khi user đăng ký
4. APIs: Get, Add, Update, Remove, Clear
5. Validate stock và quantity
6. Calculate total

**Acceptance**:
- Auto create cart khi user đăng ký
- Get/Add/Update/Remove cart items
- Clear cart
- Validate stock và quantity

**Dependencies**: Task 2.2, Task 5.1

## Phase 7: Order Management (P0)

### Task 7.1: Create Order API
**Thời gian**: 6-8 giờ  
**Priority**: P0

**Mô tả**: Create order từ cart.

**Subtasks**:
1. Models: Order, OrderItem
2. Repository layer
3. API Create Order từ cart
4. Generate order_number
5. Validate cart, stock
6. Create order_items với snapshot price/name
7. Clear cart, update stock
8. Transaction handling

**Acceptance**:
- Create order từ cart
- Order number auto-generate
- Order items lưu snapshot price/name
- Clear cart và update stock
- Transaction đảm bảo consistency

**Dependencies**: Task 6.1

### Task 7.2: Order History APIs
**Thời gian**: 4-6 giờ  
**Priority**: P0

**Mô tả**: User order history APIs.

**Subtasks**:
1. API list orders của user
2. API order detail với items
3. Filter status, date range
4. Pagination

**Acceptance**:
- List orders của user
- Order detail với items
- Filter status, date range
- Pagination

**Dependencies**: Task 7.1

### Task 7.3: Order Management APIs (Admin)
**Thời gian**: 4-6 giờ  
**Priority**: P0

**Mô tả**: Admin order management APIs.

**Subtasks**:
1. Admin APIs: List all orders, order detail
2. Update order status
3. Filter, sort, pagination
4. Validate status transition

**Acceptance**:
- Admin list all orders
- Update order status
- Filter, sort, pagination

**Dependencies**: Task 7.1, Task 2.2

## Phase 8: Rating System (P1)

### Task 8.1: Product Rating APIs
**Thời gian**: 6-8 giờ  
**Priority**: P1

**Mô tả**: Product rating APIs.

**Subtasks**:
1. Model Rating, repository layer
2. APIs: Create, Update, Get ratings
3. Validate: rating 1-5, user đã mua
4. Update rating_average và rating_count
5. One rating per user per product

**Acceptance**:
- Create/Update rating (1-5 sao)
- Chỉ đánh giá sản phẩm đã mua
- 1 rating per user per product
- Auto update rating_average và rating_count
- List ratings với user info

**Dependencies**: Task 7.1

## Phase 9: Suggestions System (P2)

### Task 9.1: Product Suggestions APIs
**Thời gian**: 4-6 giờ  
**Priority**: P2

**Mô tả**: Product suggestions APIs.

**Subtasks**:
1. Model Suggestion, repository layer
2. User API: Create suggestion
3. Admin APIs: List, Approve/Reject
4. Filter, pagination

**Acceptance**:
- User tạo suggestion
- Admin list và approve/reject
- Filter, pagination

**Dependencies**: Task 2.2, Task 4.1

## Phase 10: Notification System (P1)

### Task 10.1: Email Notification Service
**Thời gian**: 6-8 giờ  
**Priority**: P1

**Mô tả**: Email notification service cho orders.

**Subtasks**:
1. Setup email service (SMTP/SendGrid)
2. Email templates
3. Model OrderNotification
4. Service gửi email khi có order
5. Background job, retry mechanism
6. Log status vào DB

**Acceptance**:
- Gửi email khi có order mới
- Email template với order info
- Log status vào DB
- Background job, retry mechanism

**Dependencies**: Task 7.1

### Task 10.2: Chatwork Notification Service
**Thời gian**: 4-6 giờ  
**Priority**: P1

**Mô tả**: Chatwork notification service.

**Subtasks**:
1. Setup Chatwork API client
2. Service gửi message khi có order
3. Format message
4. Background job, retry
5. Log status vào DB

**Acceptance**:
- Gửi message đến Chatwork khi có order
- Log status vào DB
- Background job, retry

**Dependencies**: Task 7.1

## Phase 11: Statistics & Reports (P1)

### Task 11.1: Order Statistics API (Admin)
**Thời gian**: 6-8 giờ  
**Priority**: P1

**Mô tả**: Order statistics API cho admin.

**Subtasks**:
1. Statistics API: orders, revenue, avg order value
2. Filter date range, status
3. Group by month/week/day
4. Format cho chart
5. Optimize query

**Acceptance**:
- Statistics API: orders, revenue, avg order value
- Filter date range
- Format phù hợp cho chart

**Dependencies**: Task 7.1

### Task 11.2: Monthly Statistics Report (Scheduled Job)
**Thời gian**: 4-6 giờ  
**Priority**: P1

**Mô tả**: Monthly statistics report job.

**Subtasks**:
1. Setup cron job
2. Generate monthly report
3. Format HTML email/PDF
4. Send to admin email
5. Schedule cuối tháng
6. Log execution

**Acceptance**:
- Cron job chạy cuối tháng
- Generate và gửi report email
- Log execution

**Dependencies**: Task 11.1, Task 10.1

## Phase 12: User Management (Admin) (P1)

### Task 12.1: User Management APIs (Admin)
**Thời gian**: 6-8 giờ  
**Priority**: P1

**Mô tả**: Admin user management APIs.

**Subtasks**:
1. Admin APIs: List users, user detail
2. Update user status và role
3. Filter, sort, pagination
4. Validation: không ban admin, không tự ban mình

**Acceptance**:
- Admin list users
- Update user status và role
- Filter, sort, pagination
- Validation: không ban admin, không tự ban mình

**Dependencies**: Task 2.2

## Phase 13: Additional Features (P2-P3)

### Task 13.1: Social Share Feature
**Thời gian**: 3-4 giờ  
**Priority**: P2

**Mô tả**: Social share links API.

**Subtasks**:
1. Generate share URLs (Facebook, Twitter, Google+)
2. Format share content
3. Return trong product detail response

**Acceptance**:
- Generate share URLs (Facebook, Twitter, Google+)
- Format share content

**Dependencies**: Task 5.3

### Task 13.2: API Documentation (Swagger)
**Thời gian**: 4-6 giờ  
**Priority**: P2

**Mô tả**: Swagger/OpenAPI documentation.

**Subtasks**:
1. Setup Swagger/OpenAPI
2. Add annotations cho APIs
3. Generate Swagger UI
4. Document examples và errors

**Acceptance**:
- Swagger UI hoạt động
- Document tất cả APIs
- Examples và error responses

**Dependencies**: All API tasks

### Task 13.3: Unit Tests & Integration Tests
**Thời gian**: 8 giờ (có thể chia nhỏ)  
**Priority**: P2

**Mô tả**: Unit tests và integration tests.

**Subtasks**:
1. Setup testing framework
2. Unit tests: repository, service
3. Integration tests: APIs
4. Mock external services
5. Coverage > 70%

**Acceptance**:
- Unit tests cho repository và service
- Integration tests cho APIs
- Coverage > 70%

**Dependencies**: All tasks

## Tổng kết

**25 tasks** chia thành 13 phases:
- P0: 13 tasks (~90-110h)
- P1: 7 tasks (~40-50h)
- P2-P3: 5 tasks (~20-25h)

**Total**: ~150-185h (~19-23 ngày)
