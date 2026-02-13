-- Init script for Foods & Drinks database
-- This file runs automatically when MySQL container starts for the first time

-- Ensure proper character set
ALTER DATABASE foods_drinks CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Grant privileges to app_user
GRANT ALL PRIVILEGES ON foods_drinks.* TO 'app_user'@'%';
FLUSH PRIVILEGES;
