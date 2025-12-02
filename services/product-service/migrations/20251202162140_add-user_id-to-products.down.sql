DROP INDEX IF EXISTS idx_products_user_id;

ALTER TABLE products DROP COLUMN IF EXISTS user_id;