-- Create order_items table
CREATE TABLE `order_items` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `order_id` BIGINT UNSIGNED NOT NULL,
  `product_id` BIGINT UNSIGNED NOT NULL,
  `product_name` VARCHAR(255) NOT NULL COMMENT 'Lưu tên sản phẩm tại thời điểm đặt hàng',
  `product_price` DECIMAL(10, 2) NOT NULL COMMENT 'Lưu giá sản phẩm tại thời điểm đặt hàng',
  `quantity` INT NOT NULL,
  `subtotal` DECIMAL(10, 2) NOT NULL COMMENT 'quantity * product_price',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX `idx_order_id` (`order_id`),
  INDEX `idx_product_id` (`product_id`),
  FOREIGN KEY (`order_id`) REFERENCES `orders`(`id`) ON DELETE CASCADE,
  FOREIGN KEY (`product_id`) REFERENCES `products`(`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
