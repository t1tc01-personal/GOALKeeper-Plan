# Data Model: Notion-like workspace

## Entities

### Workspace
- **Fields**: id, ownerUserId, name, createdAt, updatedAt
- **Relationships**: has many Pages

### Page
- **Fields**: id, workspaceId, parentPageId (nullable), title, position, createdAt, updatedAt
- **Relationships**: belongs to Workspace, optional parent Page, has many Blocks

### Block
- **Fields**: id, pageId, type, content, isCompleted (for checklists), position, createdAt, updatedAt
- **Relationships**: belongs to Page

### SharePermission
- **Fields**: id, pageId, userId, role (owner, editor, viewer), createdAt, updatedAt
- **Relationships**: belongs to Page, belongs to User

## Notes

- Positions for Pages and Blocks are numeric (e.g., floats or integers) to support reordering and
  basic nesting.
- Timestamps support auditing changes for history features in future iterations.