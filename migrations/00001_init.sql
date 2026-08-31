-- +goose Up

-- ============================================================================
-- НАЧАЛЬНАЯ СХЕМА СЕРВИСА УЧЁТА РАСХОДОВ
--
-- Мягкого удаления нет нигде: удаление настоящее.
-- Все идентификаторы — INTEGER PRIMARY KEY AUTOINCREMENT, чтобы id удалённых
-- записей не переиспользовались (иначе клиентские кеши видят «новую» запись
-- под старым id).
-- ============================================================================

CREATE TABLE users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    login           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    name            TEXT,
    -- Счётчик неудачных попыток входа и время до которого вход заблокирован.
    -- Rate limit сделан по логину, а не по IP.
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until    TEXT,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);

CREATE TABLE categories (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    name        TEXT NOT NULL,
    description TEXT,
    color       TEXT,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    UNIQUE (user_id, name)
);
CREATE INDEX idx_categories_user ON categories(user_id);

CREATE TABLE expenses (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    -- RESTRICT: категорию с тратами удалить нельзя. Сервис проверяет это явным
    -- запросом и отдаёт 409, а FK остаётся последним рубежом на случай бага.
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    -- REAL с CHECK на два знака: база сама не даст записать 123.456.
    amount      REAL NOT NULL CHECK (amount > 0 AND amount = ROUND(amount, 2)),
    description TEXT,
    -- Календарная дата YYYY-MM-DD, а не момент времени: учёт ведётся по дням,
    -- это снимает вопросы с таймзонами. ISO-формат корректно сортируется
    -- и сравнивается лексикографически.
    spent_at    TEXT NOT NULL CHECK (spent_at GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]'),
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

-- Индексы составные и начинаются с user_id, потому что user_id есть в WHERE
-- каждого запроса — отдельный индекс только по spent_at был бы бесполезен.
CREATE INDEX idx_expenses_user_date     ON expenses(user_id, spent_at);
CREATE INDEX idx_expenses_user_cat_date ON expenses(user_id, category_id, spent_at);

-- +goose Down
DROP TABLE expenses;
DROP TABLE categories;
DROP TABLE users;
