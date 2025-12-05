ALTER TABLE products ADD COLUMN user_id INTEGER NOT NULL default 0;

CREATE INDEX idx_products_user_id ON products(user_id);