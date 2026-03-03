Foods & Drinks

## Architecture

- **User Site**: REST API (JSON) — frontend tách biệt, giao tiếp qua API.
- **Admin Site**: SSR (Server-Side Rendering) — Go render HTML trực tiếp, không cần frontend framework riêng.

---

[Normal User]
- Sign up/Sign in/Sign out
- Authenticate via Facebook, Twitter, Google
- Can see Food and Drink (information)
- Can filter Food, Drink via alphabet, classify (drink/food), price, category, rating …
- Can see his summary (history order, his cart)
- Can see profile
- Update profile
- Can order Food or Drink
- Can see information/image/price/number of products
- Can see products in the cart when he choice

- Can rating below each products
- Can share social-network below each products
- Can add more food or drink to cart
- Can remove food or drink from cart
- Can suggest more food or drink to admin

> Tất cả tính năng User được expose qua **REST API (JSON)**.

[Admin]
- Can manage all users
- Can manage all categories
- Can manage all product (with images)
- Can manage all list order

> Admin site được xây dựng theo kiểu **SSR**: Go render HTML template (html/template hoặc tương đương) và trả về trực tiếp — không dùng API JSON riêng cho admin.

[System]
- Send message to chatwork room with order by user
- Send email to admin with order
- Send statistic of order to admin at end of month