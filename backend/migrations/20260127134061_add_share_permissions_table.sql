-- Create share_permissions table
-- This table stores sharing permissions for pages (collaboration feature)
CREATE TABLE IF NOT EXISTS share_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('owner', 'editor', 'viewer')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    
    -- Ensure a user can only have one permission per page
    UNIQUE(page_id, user_id)
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_share_permissions_page_id ON share_permissions(page_id);
CREATE INDEX IF NOT EXISTS idx_share_permissions_user_id ON share_permissions(user_id);

