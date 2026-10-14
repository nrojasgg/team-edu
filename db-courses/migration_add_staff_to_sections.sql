-- Migration: Add staff assignment fields to sections table
-- Date: 2026-01-21

ALTER TABLE sections 
ADD COLUMN IF NOT EXISTS professor_id bigint,
ADD COLUMN IF NOT EXISTS assistant_id bigint;

-- Optional: Add indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_sections_professor_id ON sections(professor_id);
CREATE INDEX IF NOT EXISTS idx_sections_assistant_id ON sections(assistant_id);
