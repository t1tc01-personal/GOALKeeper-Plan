-- Create block_history table
-- Stores snapshots/versions of blocks
CREATE TABLE IF NOT EXISTS block_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    block_id UUID NOT NULL REFERENCES blocks(id) ON DELETE CASCADE,
    
    -- Snapshot data
    content TEXT, -- Content at the time of save
    metadata JSONB DEFAULT '{}', -- Complete logical state (Kanban, Gantt, Habit...)
    
    -- Version information
    version_number INTEGER NOT NULL, -- Incremental version number for each block
    snapshot_reason VARCHAR(50), -- Reason: manual_save, ai_transform, auto_backup
    
    -- AI context (AI Reasoning)
    ai_context JSONB DEFAULT '{}', -- If AI changed it, save why it did so
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_block_history_block_id ON block_history(block_id);
CREATE INDEX IF NOT EXISTS idx_block_history_created_at ON block_history(created_at);
CREATE INDEX IF NOT EXISTS idx_block_history_version ON block_history(block_id, version_number);

-- Unique constraint: Ensure no duplicate version for one block
CREATE UNIQUE INDEX IF NOT EXISTS idx_block_history_block_version_unique ON block_history(block_id, version_number);

