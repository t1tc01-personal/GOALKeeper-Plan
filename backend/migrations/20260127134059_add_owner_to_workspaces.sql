-- Add owner_id to workspaces table to associate workspaces with users
-- This migration adds a foreign key relationship to the users table

ALTER TABLE workspaces 
ADD COLUMN owner_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_workspaces_owner_id ON workspaces(owner_id);
