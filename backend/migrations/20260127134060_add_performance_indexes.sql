-- T046: Performance optimization indexes
-- Add indexes to improve query performance for high-traffic scenarios

-- Blocks: Optimize ListByPageID and ListByParentBlockID queries
CREATE INDEX IF NOT EXISTS idx_blocks_page_id_rank ON blocks(page_id, rank ASC) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_blocks_parent_block_id_rank ON blocks(parent_block_id, rank ASC) 
WHERE deleted_at IS NULL;

-- Pages: Optimize hierarchy and list queries
CREATE INDEX IF NOT EXISTS idx_pages_workspace_parent ON pages(workspace_id, parent_page_id) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_pages_workspace_created ON pages(workspace_id, created_at DESC) 
WHERE deleted_at IS NULL;

-- Workspaces: Optimize owner lookup
CREATE INDEX IF NOT EXISTS idx_workspaces_owner_id_created ON workspaces(owner_id, created_at DESC) 
WHERE deleted_at IS NULL;

-- Share permissions: Optimize permission lookups
-- Note: These indexes are created only if the share_permissions table exists
-- The table is created in a separate migration (20260127134059_add_share_permissions_table.sql)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'share_permissions') THEN
        CREATE INDEX IF NOT EXISTS idx_share_permissions_page_user ON share_permissions(page_id, user_id) 
        WHERE deleted_at IS NULL;

        CREATE INDEX IF NOT EXISTS idx_share_permissions_user_created ON share_permissions(user_id, created_at DESC) 
        WHERE deleted_at IS NULL;
    END IF;
END $$;

-- Block type lookups
CREATE INDEX IF NOT EXISTS idx_block_types_name ON block_types(name);

-- User role lookups (for RBAC)
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
