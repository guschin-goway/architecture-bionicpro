-- Создаём базу и таблицу
CREATE DATABASE IF NOT EXISTS default;

CREATE TABLE IF NOT EXISTS default.report_user_daily
(
    report_date Date,
    prosthesis_id UInt32,
    movements_count UInt32,
    avg_reaction_time_ms Float32,
    battery_avg_level Float32,
    errors_count UInt32,
    user_id String
)
    ENGINE = MergeTree()
ORDER BY (user_id, report_date);

-- Вставка данных (ClickHouse требует отдельного VALUES на каждую строку)
INSERT INTO default.report_user_daily
(report_date, prosthesis_id, movements_count, avg_reaction_time_ms, battery_avg_level, errors_count, user_id)
VALUES ('2026-02-01', 1, 120, 250.5, 90.0, 2, 'user1');

INSERT INTO default.report_user_daily
(report_date, prosthesis_id, movements_count, avg_reaction_time_ms, battery_avg_level, errors_count, user_id)
VALUES ('2026-02-02', 1, 130, 240.2, 88.5, 1, 'user1');

INSERT INTO default.report_user_daily
(report_date, prosthesis_id, movements_count, avg_reaction_time_ms, battery_avg_level, errors_count, user_id)
VALUES ('2026-02-03', 2, 100, 260.0, 85.0, 3, 'user2');

INSERT INTO default.report_user_daily
(report_date, prosthesis_id, movements_count, avg_reaction_time_ms, battery_avg_level, errors_count, user_id)
VALUES ('2026-02-04', 2, 110, 255.5, 87.0, 0, 'user2');

INSERT INTO default.report_user_daily
(report_date, prosthesis_id, movements_count, avg_reaction_time_ms, battery_avg_level, errors_count, user_id)
VALUES ('2026-02-01', 3, 95, 270.0, 92.0, 1, 'user3');
