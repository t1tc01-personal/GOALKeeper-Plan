-- Create pages table
-- Pages can be either Doc or Planner
CREATE TABLE IF NOT EXISTS pages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    parent_page_id UUID REFERENCES pages(id) ON DELETE SET NULL, -- Recursive: page within page
    title TEXT NOT NULL,
    
    -- Planner coordination: If this page is a Planner, it links to framework_library
    is_planner BOOLEAN DEFAULT FALSE,
    framework_id UUID REFERENCES framework_library(id) ON DELETE SET NULL,
    
    view_config JSONB DEFAULT '{}', -- Store view config: filter, sort, hide/show columns
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_pages_workspace_id ON pages(workspace_id);
CREATE INDEX IF NOT EXISTS idx_pages_parent_page_id ON pages(parent_page_id);
CREATE INDEX IF NOT EXISTS idx_pages_framework_id ON pages(framework_id);
CREATE INDEX IF NOT EXISTS idx_pages_is_planner ON pages(is_planner);

