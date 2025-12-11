-- Удаляем триггер
DROP TRIGGER IF EXISTS trg_cart_set_updated_at ON cart_items;

-- Удаляем функцию
DROP FUNCTION IF EXISTS set_updated_at();

-- Удаляем индексы (если нужны, но DROP TABLE удалит их автоматически)
DROP INDEX IF EXISTS idx_cart_user;
DROP INDEX IF EXISTS idx_cart_product;

-- Удаляем таблицу
DROP TABLE IF EXISTS cart_items;
