-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
-- Add up migration script here
CREATE TABLE IF NOT EXISTS users (
    "id" SERIAL PRIMARY KEY,
    "name" TEXT NOT NULL,
    "pass" TEXT NOT NULL,
    "coins" INT CHECK (coins >= 0)
);

CREATE TABLE IF NOT EXISTS transactions (
    "id" SERIAL PRIMARY KEY,
    "sender_id" INT REFERENCES users(id) ON DELETE SET NULL,
    "receiver_id" INT REFERENCES users(id) ON DELETE SET NULL,
    "amount" INT NOT NULL CHECK (amount > 0),
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE user_merch (
    "id" SERIAL PRIMARY KEY,
    "user_id" INT NOT NULL,
    "item" TEXT NOT NULL,
    "quantity" INT NOT NULL,
    UNIQUE (user_id, item),
    FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS merch (
    "id" SERIAL PRIMARY KEY,
    "name" TEXT UNIQUE NOT NULL,
    "price" INT NOT NULL CHECK (price > 0)
);

INSERT INTO merch (name, price) VALUES
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500);

INSERT INTO users (name, pass, coins) VALUES
('Иван Иванов', 'pass1', 1000),
('Мария Петрова', 'pass2', 1200),
('Алексей Смирнов', 'pass3', 850),
('Елена Кузнецова', 'pass4', 1500),
('Дмитрий Попов', 'pass5', 900),
('Ольга Васильева', 'pass6', 2000),
('Сергей Петров', 'pass7', 750),
('Анна Соколова', 'pass8', 1800),
('Николай Михайлов', 'pass9', 950),
('Татьяна Новикова', 'pass10', 1300),
('Андрей Федоров', 'pass11', 1100),
('Екатерина Морозова', 'pass12', 1600),
('Павел Волков', 'pass13', 700),
('Юлия Алексеева', 'pass14', 1400),
('Владимир Лебедев', 'pass15', 1250),
('Надежда Козлова', 'pass16', 1150),
('Артем Егоров', 'pass17', 1900),
('Людмила Павлова', 'pass18', 800),
('Григорий Семенов', 'pass19', 1700),
('Виктория Голубева', 'pass20', 2200);

INSERT INTO transactions (sender_id, receiver_id, amount) VALUES
(1, 2, 100), (2, 3, 50), (4, 5, 200), (6, 7, 150),
(3, 8, 80), (5, 10, 300), (7, 12, 120), (9, 15, 90),
(10, 1, 250), (12, 4, 180), (14, 6, 220), (16, 18, 70),
(18, 20, 400), (19, 17, 150), (15, 11, 60), (13, 9, 85),
(17, 14, 110), (8, 19, 95), (20, 16, 130), (11, 13, 45);

INSERT INTO user_merch (user_id, item, quantity) VALUES
(1, 't-shirt', 2), (1, 'cup', 1), (2, 'book', 3), (3, 'pen', 5),
(4, 'powerbank', 1), (5, 'hoody', 2), (6, 'umbrella', 1),
(7, 'socks', 4), (8, 'wallet', 2), (9, 'pink-hoody', 1),
(10, 't-shirt', 3), (11, 'cup', 2), (12, 'book', 1),
(13, 'pen', 4), (14, 'powerbank', 2), (15, 'hoody', 1),
(16, 'umbrella', 3), (17, 'socks', 2), (18, 'wallet', 1),
(19, 'pink-hoody', 2), (20, 't-shirt', 1);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
-- Add down migration script here
DROP TABLE IF EXISTS merch;
DROP TABLE IF EXISTS user_merch;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS users;