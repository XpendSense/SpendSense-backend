CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    hashed_password TEXT,
    first_name      VARCHAR(100),
    last_name       VARCHAR(100),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    is_superuser    BOOLEAN NOT NULL DEFAULT FALSE,
    is_verified     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE oauth_account (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    oauth_name    VARCHAR(50) NOT NULL,
    account_id    VARCHAR(255) NOT NULL,
    account_email VARCHAR(255) NOT NULL,
    UNIQUE (oauth_name, account_id)
);

CREATE TABLE budget (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    start_date DATE,
    end_date   DATE,
    active     BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE budget_to_user_mapping (
    id        SERIAL PRIMARY KEY,
    budget_id UUID NOT NULL REFERENCES budget(id) ON DELETE CASCADE,
    user_name VARCHAR(100),
    user_id   UUID REFERENCES users(id)
);

CREATE TABLE income_to_budget_mapping (
    id        SERIAL PRIMARY KEY,
    budget_id UUID NOT NULL REFERENCES budget(id) ON DELETE CASCADE,
    user_id   UUID REFERENCES users(id),
    name      VARCHAR(100),
    amount    NUMERIC(15, 4) NOT NULL DEFAULT 0,
    recurring BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE transaction_type (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

CREATE TABLE transaction_frequency (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

CREATE TABLE category_type (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

CREATE TABLE payment_type (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

CREATE TABLE category (
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(100) NOT NULL,
    type_id INTEGER REFERENCES category_type(id),
    user_id UUID REFERENCES users(id)
);

CREATE TABLE payment_methods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    payment_type_id INTEGER REFERENCES payment_type(id),
    user_id         UUID REFERENCES users(id)
);

CREATE TABLE transaction (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                     VARCHAR(100),
    amount                   NUMERIC(15, 4) NOT NULL DEFAULT 0,
    planned_amount           NUMERIC(15, 4) NOT NULL DEFAULT 0,
    date                     DATE,
    renewal_date             DATE,
    recurring                BOOLEAN,
    budget_id                UUID REFERENCES budget(id) ON DELETE CASCADE,
    category_id              INTEGER REFERENCES category(id),
    payment_method_id        UUID REFERENCES payment_methods(id),
    transaction_frequency_id INTEGER REFERENCES transaction_frequency(id),
    transaction_type_id      INTEGER REFERENCES transaction_type(id)
);

INSERT INTO transaction_type (name) VALUES ('Fixed'), ('Variable');

INSERT INTO transaction_frequency (name) VALUES
    ('One-off'), ('Weekly'), ('Bi-weekly'), ('Monthly'), ('Yearly');

INSERT INTO category_type (name) VALUES ('Expense'), ('Saving'), ('Income');

INSERT INTO payment_type (name) VALUES
    ('Cash'), ('Credit'), ('Debit'), ('Digital Wallet'),
    ('Bank Transfer'), ('Crypto'), ('Investment'), ('Other');
