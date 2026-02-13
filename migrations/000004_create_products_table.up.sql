-- Create products table
CREATE TABLE `products` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `category_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `slug` VARCHAR(255) NOT NULL UNIQUE,
  `description` TEXT NULL,
  `classify` VARCHAR(50) NOT NULL COMMENT 'Các giá trị: food, drink',
  `price` DECIMAL(10, 2) NOT NULL,
  `stock` INT NOT NULL DEFAULT 0 COMMENT 'Số lượng tồn kho',
  `rating_average` DECIMAL(3, 2) NOT NULL DEFAULT 0.00 COMMENT 'Điểm đánh giá trung bình',
  `rating_count` INT NOT NULL DEFAULT 0 COMMENT 'Số lượng đánh giá',
  `status` VARCHAR(50) NOT NULL DEFAULT 'active' COMMENT 'Các giá trị: active, inactive, out_of_stock',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` TIMESTAMP NULL,

  INDEX `idx_category_id` (`category_id`),
  INDEX `idx_slug` (`slug`),
  INDEX `idx_classify` (`classify`),
  INDEX `idx_price` (`price`),
  INDEX `idx_rating_average` (`rating_average`),
  INDEX `idx_status` (`status`),
  INDEX `idx_deleted_at` (`deleted_at`),
  FULLTEXT INDEX `ft_name_description` (`name`, `description`),
  FOREIGN KEY (`category_id`) REFERENCES `categories`(`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
