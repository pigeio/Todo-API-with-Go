-- migrations/000001_create_tables.down.sql

-- Drop trigger first
DROP TRIGGER IF EXISTS update_todos_updated_at ON todos;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_todos_user_id;
DROP INDEX IF EXISTS idx_todos_completed;
DROP INDEX IF EXISTS idx_todos_created_at;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS todos;
DROP TABLE IF EXISTS users;