-- Create block_types table
-- Stores metadata and configuration for all available block types
CREATE TABLE IF NOT EXISTS block_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL, -- text, heading_1, task_item, etc.
    category VARCHAR(50) NOT NULL, -- content, framework, media, smart
    display_name VARCHAR(100) NOT NULL, -- Human-readable name
    description TEXT,
    icon VARCHAR(50), -- Icon identifier for frontend
    default_metadata JSONB DEFAULT '{}', -- Default metadata schema
    metadata_schema JSONB DEFAULT '{}', -- JSON Schema for validation
    is_system BOOLEAN DEFAULT TRUE, -- System types cannot be deleted
    is_framework_block BOOLEAN DEFAULT FALSE, -- True for framework_container, task_item, etc.
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_block_types_category ON block_types(category);
CREATE INDEX IF NOT EXISTS idx_block_types_is_framework ON block_types(is_framework_block);

-- Seed default block types

-- Content Blocks
INSERT INTO block_types (name, category, display_name, description, icon, default_metadata, is_system) VALUES
    ('text', 'content', 'Text', 'Plain text block with Markdown support', 'text', '{"format": "markdown"}', true),
    ('heading_1', 'content', 'Heading 1', 'Main heading', 'heading-1', '{"level": 1}', true),
    ('heading_2', 'content', 'Heading 2', 'Sub heading', 'heading-2', '{"level": 2}', true),
    ('heading_3', 'content', 'Heading 3', 'Sub-sub heading', 'heading-3', '{"level": 3}', true),
    ('bulleted_list', 'content', 'Bulleted List', 'Unordered list', 'list-bulleted', '{"items": []}', true),
    ('numbered_list', 'content', 'Numbered List', 'Ordered list', 'list-numbered', '{"items": []}', true),
    ('todo_list', 'content', 'Todo List', 'Basic checklist (different from advanced Task)', 'checkbox', '{"checked": false}', true),
    ('toggle', 'content', 'Toggle', 'Collapsible block to hide/show content (important for Learning Roadmap)', 'toggle', '{"collapsed": false}', true),
    ('callout', 'content', 'Callout', 'Highlighted block with icon for notices', 'callout', '{"icon": "info", "color": "blue"}', true),
    ('quote', 'content', 'Quote', 'Quote block', 'quote', '{}', true),
    ('divider', 'content', 'Divider', 'Horizontal divider line', 'divider', '{}', true)
ON CONFLICT (name) DO NOTHING;

-- Framework Blocks
INSERT INTO block_types (name, category, display_name, description, icon, default_metadata, is_system, is_framework_block) VALUES
    ('framework_container', 'framework', 'Framework Container', 'Parent block containing framework configuration (Kanban, Gantt, Habit, etc.)', 'container', '{"framework_type": null, "columns": [], "settings": {}}', true, true),
    ('task_item', 'framework', 'Task Item', 'Task block used for Kanban, Gantt, Action Plan. Metadata determines where it is displayed', 'task', '{"status": "todo", "priority": "medium", "due_date": null}', true, true),
    ('habit_item', 'framework', 'Habit Item', 'Habit block with history and streak data', 'habit', '{"history": [], "streak": 0, "frequency": "daily"}', true, true),
    ('wbs_node', 'framework', 'WBS Node', 'Work breakdown structure node with parent-child calculation logic', 'wbs', '{"progress": 0}', true, true),
    ('learning_node', 'framework', 'Learning Node', 'Learning roadmap node with understanding status and attached documents', 'learning', '{"understood": false}', true, true),
    ('milestone', 'framework', 'Milestone', 'Important milestone point (often used in Gantt and Roadmap)', 'milestone', '{"date": null, "completed": false}', true, true)
ON CONFLICT (name) DO NOTHING;

-- Media & Embed Blocks
INSERT INTO block_types (name, category, display_name, description, icon, default_metadata, is_system) VALUES
    ('image', 'media', 'Image', 'Display image with zoom and caption support', 'image', '{"url": "", "caption": ""}', true),
    ('file_attachment', 'media', 'File Attachment', 'PDF, Docx, Zip files with icon and size', 'file', '{"url": "", "filename": "", "size": 0, "mime_type": ""}', true),
    ('video_embed', 'media', 'Video Embed', 'Embed video from Youtube, Vimeo or direct link', 'video', '{"url": "", "provider": null}', true),
    ('web_bookmark', 'media', 'Web Bookmark', 'Display link as card with preview image, title, description', 'bookmark', '{"url": "", "title": "", "description": ""}', true),
    ('code_snippet', 'media', 'Code Snippet', 'Code block with syntax highlighting', 'code', '{"language": "plaintext", "code": ""}', true)
ON CONFLICT (name) DO NOTHING;

-- Smart / AI Blocks
INSERT INTO block_types (name, category, display_name, description, icon, default_metadata, is_system) VALUES
    ('ai_prompt', 'smart', 'AI Prompt', 'Block allowing user to write AI requests directly in page', 'ai-prompt', '{"prompt": "", "status": "pending"}', true),
    ('ai_response', 'smart', 'AI Response', 'Block containing AI response, usually with distinctive background', 'ai-response', '{"response": "", "model": null}', true),
    ('sync_block', 'smart', 'Sync Block', 'Synchronized block; content changes propagate to linked blocks', 'sync', '{"sync_id": ""}', true),
    ('button', 'smart', 'Button', 'Button block to trigger automations (e.g., create daily habit tasks)', 'button', '{"label": "", "action": "", "automation_id": null}', true)
ON CONFLICT (name) DO NOTHING;


-- Create blocks table
-- Blocks are atomic data units
CREATE TABLE IF NOT EXISTS blocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    parent_block_id UUID REFERENCES blocks(id) ON DELETE CASCADE,
    type_id UUID NOT NULL REFERENCES block_types(id) ON DELETE CASCADE,
    content TEXT,
    metadata JSONB DEFAULT '{}',
    rank BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_blocks_page_id ON blocks(page_id);
CREATE INDEX IF NOT EXISTS idx_blocks_parent_block_id ON blocks(parent_block_id);
CREATE INDEX IF NOT EXISTS idx_blocks_type_id ON blocks(type_id);
CREATE INDEX IF NOT EXISTS idx_blocks_rank ON blocks(page_id, rank); -- Composite index for ordering

