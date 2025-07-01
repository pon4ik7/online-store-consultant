DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS popular_products;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS anonymous_sessions;

CREATE TABLE user_sessions
(
    session_id          UUID PRIMARY KEY,
    context             TEXT,
    last_active         TIMESTAMP DEFAULT NOW(),
    was_context_updated BOOLEAN   DEFAULT FALSE,
    user_id             UUID
);

CREATE TABLE anonymous_sessions
(
    session_id  UUID PRIMARY KEY,
    last_active TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE popular_products
(
    product_id  SERIAL PRIMARY KEY,
    name        VARCHAR(255),
    description TEXT,
    price       DECIMAL(10, 2),
    rating      DECIMAL(3, 2),
    category    VARCHAR(255),
    product_url VARCHAR(255),
    image_url   VARCHAR(255)
);

CREATE TABLE users
(
    user_id     UUID PRIMARY KEY,
    credentials TEXT,
    session_id  UUID,
    FOREIGN KEY (session_id) REFERENCES user_sessions (session_id) ON DELETE CASCADE
);
