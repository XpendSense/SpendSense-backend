-- +goose Up

-- Drop old budget-scoped tables in FK-safe order
DROP TABLE IF EXISTS transaction;
DROP TABLE IF EXISTS income_to_budget_mapping;
DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS budget_to_user_mapping;
DROP TABLE IF EXISTS budget;

-- Stable, user-configured budget. People, payment methods, and income sources
-- live here and carry forward across every period.
CREATE TABLE budget_profile (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    cycle      VARCHAR(20)  NOT NULL DEFAULT 'monthly',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- People on a profile (shared across all periods of this profile)
CREATE TABLE budget_to_profile_mapping (
    id                SERIAL       PRIMARY KEY,
    budget_profile_id UUID         NOT NULL REFERENCES budget_profile(id) ON DELETE CASCADE,
    user_name         VARCHAR(100),
    user_id           UUID         REFERENCES users(id),
    is_active         BOOLEAN      NOT NULL DEFAULT TRUE
);

-- Payment methods (person-attributed, profile-scoped via budget_to_profile_mapping)
CREATE TABLE payment_methods (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name             VARCHAR(100) NOT NULL,
    payment_type_id  INTEGER      REFERENCES payment_type(id),
    user_id          UUID         REFERENCES users(id),
    is_active        BOOLEAN      NOT NULL DEFAULT TRUE,
    budget_person_id INTEGER      REFERENCES budget_to_profile_mapping(id)
);

-- Income source templates (profile-level; carry forward each period)
CREATE TABLE income_source (
    id                SERIAL         PRIMARY KEY,
    budget_profile_id UUID           NOT NULL REFERENCES budget_profile(id) ON DELETE CASCADE,
    budget_person_id  INTEGER        REFERENCES budget_to_profile_mapping(id),
    name              VARCHAR(100)   NOT NULL,
    income_type       VARCHAR(20)    NOT NULL DEFAULT 'other',
    default_amount    NUMERIC(15, 4) NOT NULL DEFAULT 0,
    recurring         BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- One cycle window of a profile. First is created on profile setup; rest by the cycling job.
CREATE TABLE budget_period (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_profile_id UUID        NOT NULL REFERENCES budget_profile(id) ON DELETE CASCADE,
    start_date        DATE        NOT NULL,
    end_date          DATE        NOT NULL,
    is_archived       BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Actual income for one period, pre-filled from income_source; user edits each period.
CREATE TABLE income_entry (
    id               SERIAL         PRIMARY KEY,
    budget_period_id UUID           NOT NULL REFERENCES budget_period(id) ON DELETE CASCADE,
    income_source_id INTEGER        REFERENCES income_source(id),
    budget_person_id INTEGER        REFERENCES budget_to_profile_mapping(id),
    name             VARCHAR(100),
    amount           NUMERIC(15, 4) NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- Transactions are period-scoped
CREATE TABLE transaction (
    id                       UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    name                     VARCHAR(100),
    amount                   NUMERIC(15, 4) NOT NULL DEFAULT 0,
    planned_amount           NUMERIC(15, 4) NOT NULL DEFAULT 0,
    date                     DATE,
    renewal_date             DATE,
    recurring                BOOLEAN,
    budget_period_id         UUID           REFERENCES budget_period(id) ON DELETE CASCADE,
    category_id              INTEGER        REFERENCES category(id),
    payment_method_id        UUID           REFERENCES payment_methods(id),
    transaction_frequency_id INTEGER        REFERENCES transaction_frequency(id),
    transaction_type_id      INTEGER        REFERENCES transaction_type(id)
);

-- +goose Down

DROP TABLE IF EXISTS transaction;
DROP TABLE IF EXISTS income_entry;
DROP TABLE IF EXISTS budget_period;
DROP TABLE IF EXISTS income_source;
DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS budget_to_profile_mapping;
DROP TABLE IF EXISTS budget_profile;
