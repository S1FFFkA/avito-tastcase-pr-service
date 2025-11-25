-- Скрипт для очистки базы данных перед нагрузочным тестированием
-- Выполните этот скрипт в вашей БД перед запуском теста в Postman

-- Очистка всех таблиц (в правильном порядке из-за внешних ключей)
TRUNCATE TABLE reviewers CASCADE;
TRUNCATE TABLE pull_requests CASCADE;
TRUNCATE TABLE users CASCADE;
TRUNCATE TABLE teams CASCADE;

-- Проверка, что таблицы очищены
SELECT 
    (SELECT COUNT(*) FROM teams) as teams_count,
    (SELECT COUNT(*) FROM users) as users_count,
    (SELECT COUNT(*) FROM pull_requests) as prs_count,
    (SELECT COUNT(*) FROM reviewers) as reviewers_count;

