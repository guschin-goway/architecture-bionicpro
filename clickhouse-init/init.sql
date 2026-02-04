-- Создаём базу и таблицу
CREATE DATABASE IF NOT EXISTS default;

CREATE TABLE IF NOT EXISTS user_reports (
                user_id String,
                sensor_value Int32,
                ts DateTime,
                plan String
            ) ENGINE = MergeTree()
            ORDER BY ts;