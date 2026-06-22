-- Add thumbnail_id column to documents table
ALTER TABLE documents ADD COLUMN thumbnail_id UUID;

-- Create index for faster thumbnail lookups
CREATE INDEX idx_documents_thumbnail_id ON documents(thumbnail_id);
