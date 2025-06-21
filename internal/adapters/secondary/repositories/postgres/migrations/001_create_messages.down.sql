-- Drop trigger first
DROP TRIGGER IF EXISTS update_messages_updated_at ON messages;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_messages_status_created_at;
DROP INDEX IF EXISTS idx_messages_phone_number;
DROP INDEX IF EXISTS idx_messages_external_id;
DROP INDEX IF EXISTS idx_messages_status;

-- Drop table
DROP TABLE IF EXISTS messages; 