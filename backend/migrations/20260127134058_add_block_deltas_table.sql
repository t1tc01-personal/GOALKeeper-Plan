-- Create block_deltas table
-- Stores detailed changes (Delta/Event Sourcing) - Optional for extremely detailed undo
CREATE TABLE IF NOT EXISTS block_deltas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    block_id UUID NOT NULL REFERENCES blocks(id) ON DELETE CASCADE,
    
    delta JSONB NOT NULL DEFAULT '{}', -- Only store differences from previous version (JSON Patch)
    actor_type VARCHAR(20) NOT NULL, -- 'user' or 'ai'
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_block_deltas_block_id ON block_deltas(block_id);
CREATE INDEX IF NOT EXISTS idx_block_deltas_created_at ON block_deltas(created_at);
CREATE INDEX IF NOT EXISTS idx_block_deltas_actor_type ON block_deltas(actor_type);