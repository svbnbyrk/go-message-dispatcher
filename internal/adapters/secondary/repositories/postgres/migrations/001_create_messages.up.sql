-- Create messages table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    external_id VARCHAR(255),
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_messages_status_created_at ON messages(status, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_phone_number ON messages(phone_number);
CREATE INDEX IF NOT EXISTS idx_messages_external_id ON messages(external_id) WHERE external_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);

-- Add check constraints for business rules
ALTER TABLE messages ADD CONSTRAINT chk_messages_status 
    CHECK (status IN ('PENDING', 'SENT', 'FAILED'));

ALTER TABLE messages ADD CONSTRAINT chk_messages_retry_count 
    CHECK (retry_count >= 0 AND retry_count <= 3);

ALTER TABLE messages ADD CONSTRAINT chk_messages_phone_number_not_empty 
    CHECK (LENGTH(TRIM(phone_number)) > 0);

ALTER TABLE messages ADD CONSTRAINT chk_messages_content_not_empty 
    CHECK (LENGTH(TRIM(content)) > 0);

-- Create function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at field
CREATE TRIGGER update_messages_updated_at 
    BEFORE UPDATE ON messages 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 