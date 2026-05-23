-- Remove thumbnail_id column from documents table
DROP INDEX IF EXISTS idx_documents_thumbnail_id;
ALTER TABLE documents DROP COLUMN IF EXISTS thumbnail_id;
