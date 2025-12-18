ALTER TABLE users
    ADD COLUMN role TEXT NOT NULL DEFAULT 'buyer';

ALTER TABLE users
    ADD CONSTRAINT users_role_check
        CHECK (role IN ('buyer', 'seller'));
