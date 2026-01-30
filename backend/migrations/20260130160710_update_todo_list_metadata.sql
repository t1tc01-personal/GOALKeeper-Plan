-- Update todo_list block type to use single-item-per-block pattern
-- Change from {"items": [], "checked": []} to {"checked": false}

UPDATE block_types
SET default_metadata = '{"checked": false}'::jsonb
WHERE name = 'todo_list';
