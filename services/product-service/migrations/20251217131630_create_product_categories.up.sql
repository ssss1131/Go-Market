CREATE TABLE IF NOT EXISTS product_categories (
                                                  product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                                                  category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
                                                  PRIMARY KEY (product_id, category_id)
);

CREATE INDEX IF NOT EXISTS idx_product_categories_product_id ON product_categories(product_id);
CREATE INDEX IF NOT EXISTS idx_product_categories_category_id ON product_categories(category_id);
