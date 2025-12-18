CREATE TABLE IF NOT EXISTS cart_items
(
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL,
    product_id INT NOT NULL,

    quantity   INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_cart_user
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE,

    CONSTRAINT fk_cart_product
    FOREIGN KEY (product_id)
    REFERENCES products(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT,

    UNIQUE(user_id, product_id)
    );

CREATE INDEX IF NOT EXISTS idx_cart_user ON cart_items(user_id);
CREATE INDEX IF NOT EXISTS idx_cart_product ON cart_items(product_id);

CREATE TRIGGER trg_cart_set_updated_at
    BEFORE UPDATE ON cart_items
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
