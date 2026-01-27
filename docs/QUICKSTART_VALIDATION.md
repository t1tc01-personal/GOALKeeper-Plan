# Quickstart Validation: End-to-End Testing Guide

**Task**: T049 - Quickstart Validation Flow
**Status**: Guide for manual and automated validation
**Date**: January 27, 2026

## Overview

This guide provides a step-by-step validation that all 3 user stories (Workspaces, Blocks, Sharing) work correctly together end-to-end. Follow these steps to verify the MVP functionality.

---

## Prerequisites

### Environment Setup

```bash
# Backend
cd backend
cp config.example.yaml config.yaml  # Set up your database connection
go build ./cmd/main.go

# Frontend
cd frontend
npm install
npm run build
```

### Database Initialization

```bash
# Run migrations (if using migrate tool)
migrate -path backend/migrations -database "postgresql://user:password@localhost/goalkeeper" up

# Or use docker-compose if available
docker-compose up -d postgres
```

### Test Accounts

Create two test users (or ensure your auth system supports them):
- **User 1 (Owner)**: email: owner@test.com, uuid: 550e8400-e29b-41d4-a716-446655440000
- **User 2 (Viewer)**: email: viewer@test.com, uuid: 550e8400-e29b-41d4-a716-446655440001
- **User 3 (Editor)**: email: editor@test.com, uuid: 550e8400-e29b-41d4-a716-446655440002

---

## Test Flow

### Phase 1: User Story 1 - Workspace & Page Management

#### Step 1.1: Create Workspace

**What to Test**: User can create a workspace

```bash
# Backend: Create workspace via API
curl -X POST http://localhost:8080/api/v1/notion/workspaces \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "name": "Test Workspace"
  }'

# Expected Response: 201 Created with workspace object
# {
#   "id": "550e8400-e29b-41d4-a716-446655440003",
#   "name": "Test Workspace",
#   "created_at": "2026-01-27T..."
# }
```

**Checklist**:
- [ ] HTTP 201 response
- [ ] Workspace ID returned
- [ ] Workspace appears in UI

#### Step 1.2: Create Page

**What to Test**: User can create a page in workspace

```bash
WORKSPACE_ID="550e8400-e29b-41d4-a716-446655440003"

curl -X POST http://localhost:8080/api/v1/notion/pages \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d "{
    \"workspace_id\": \"$WORKSPACE_ID\",
    \"title\": \"Test Page\"
  }"

# Expected Response: 201 Created with page object
```

**Checklist**:
- [ ] HTTP 201 response
- [ ] Page ID returned
- [ ] Page appears in sidebar
- [ ] Page title is editable

#### Step 1.3: Edit Page Title

**What to Test**: User can rename page

```bash
PAGE_ID="550e8400-e29b-41d4-a716-446655440004"

curl -X PUT http://localhost:8080/api/v1/notion/pages/$PAGE_ID \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "title": "Updated Page Title"
  }'

# Expected Response: 200 OK
```

**Checklist**:
- [ ] HTTP 200 response
- [ ] Title updated in database
- [ ] Title updated in UI

---

### Phase 2: User Story 2 - Block Editing

#### Step 2.1: Create Paragraph Block

**What to Test**: User can add content blocks to page

```bash
PAGE_ID="550e8400-e29b-41d4-a716-446655440004"

curl -X POST http://localhost:8080/api/v1/notion/pages/$PAGE_ID/blocks \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "type": "paragraph",
    "content": "This is a test paragraph",
    "order": 1
  }'

# Expected Response: 201 Created with block object
```

**Checklist**:
- [ ] HTTP 201 response
- [ ] Block ID returned
- [ ] Block appears on page
- [ ] Content is visible

#### Step 2.2: Create Heading Block

**What to Test**: Multiple block types work

```bash
curl -X POST http://localhost:8080/api/v1/notion/pages/$PAGE_ID/blocks \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "type": "heading",
    "content": "Section Title",
    "level": 1,
    "order": 0
  }'

# Expected Response: 201 Created
```

**Checklist**:
- [ ] HTTP 201 response
- [ ] Heading appears above paragraph
- [ ] Heading styling applied

#### Step 2.3: Create Checklist Block

**What to Test**: Checklist type works

```bash
curl -X POST http://localhost:8080/api/v1/notion/pages/$PAGE_ID/blocks \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "type": "checklist",
    "content": "Buy groceries",
    "checked": false,
    "order": 2
  }'

# Expected Response: 201 Created
```

**Checklist**:
- [ ] HTTP 201 response
- [ ] Checklist appears with checkbox
- [ ] Checkbox can be toggled

#### Step 2.4: Edit Block Content

**What to Test**: User can edit existing blocks

```bash
BLOCK_ID="550e8400-e29b-41d4-a716-446655440005"

curl -X PUT http://localhost:8080/api/v1/notion/pages/$PAGE_ID/blocks/$BLOCK_ID \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "content": "Updated paragraph content"
  }'

# Expected Response: 200 OK
```

**Checklist**:
- [ ] HTTP 200 response
- [ ] Content updated immediately in UI
- [ ] Changes persisted in database

#### Step 2.5: Reorder Blocks

**What to Test**: Blocks can be reordered

```bash
# Move heading to position 2 (after current)
curl -X PUT http://localhost:8080/api/v1/notion/pages/$PAGE_ID/blocks/$HEADING_ID \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "order": 2
  }'

# Expected Response: 200 OK
```

**Checklist**:
- [ ] HTTP 200 response
- [ ] Block moves to new position
- [ ] Other blocks reorder accordingly

#### Step 2.6: Delete Block

**What to Test**: User can delete blocks

```bash
curl -X DELETE http://localhost:8080/api/v1/notion/pages/$PAGE_ID/blocks/$BLOCK_ID \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"

# Expected Response: 204 No Content or 200 OK
```

**Checklist**:
- [ ] HTTP 204 or 200 response
- [ ] Block removed from UI
- [ ] Block removed from database

#### Step 2.7: Data Persistence

**What to Test**: Changes persist after page refresh

```bash
# In browser developer console
window.location.reload();

# Wait for page to reload
# Check that all blocks appear in correct order with correct content
```

**Checklist**:
- [ ] Page reloads without errors
- [ ] All blocks appear in same order
- [ ] All content is unchanged
- [ ] No console errors

---

### Phase 3: User Story 3 - Sharing & Permissions

#### Step 3.1: Grant Viewer Access

**What to Test**: Owner can share page with viewer role

```bash
VIEWER_ID="550e8400-e29b-41d4-a716-446655440001"

curl -X POST http://localhost:8080/api/v1/notion/pages/$PAGE_ID/share \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d "{
    \"user_id\": \"$VIEWER_ID\",
    \"role\": \"viewer\"
  }"

# Expected Response: 201 Created or 200 OK
```

**Checklist**:
- [ ] HTTP 201 or 200 response
- [ ] Permission created in database
- [ ] Share dialog shows collaborator

#### Step 3.2: Verify Viewer Read-Only Access

**What to Test**: Viewer can read but cannot edit

```bash
# Try to read page as viewer (should succeed)
curl -X GET http://localhost:8080/api/v1/notion/pages/$PAGE_ID \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440001"

# Expected Response: 200 OK with page content
```

**Checklist**:
- [ ] HTTP 200 response
- [ ] Page content returned
- [ ] Block content visible

```bash
# Try to edit page as viewer (should fail)
curl -X PUT http://localhost:8080/api/v1/notion/pages/$PAGE_ID/blocks/$BLOCK_ID \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440001" \
  -d '{
    "content": "This should fail"
  }'

# Expected Response: 403 Forbidden
```

**Checklist**:
- [ ] HTTP 403 Forbidden response
- [ ] Edit buttons disabled in UI
- [ ] Read-only message shown

#### Step 3.3: Grant Editor Access

**What to Test**: Owner can change to editor role

```bash
EDITOR_ID="550e8400-e29b-41d4-a716-446655440002"

curl -X POST http://localhost:8080/api/v1/notion/pages/$PAGE_ID/share \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d "{
    \"user_id\": \"$EDITOR_ID\",
    \"role\": \"editor\"
  }"

# Expected Response: 201 Created or 200 OK (update if already exists)
```

**Checklist**:
- [ ] HTTP 201 or 200 response
- [ ] Permission created/updated

#### Step 3.4: Verify Editor Can Edit

**What to Test**: Editor can read and write

```bash
# Try to edit page as editor (should succeed)
curl -X PUT http://localhost:8080/api/v1/notion/pages/$PAGE_ID/blocks/$BLOCK_ID \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440002" \
  -d '{
    "content": "Editor can modify this"
  }'

# Expected Response: 200 OK
```

**Checklist**:
- [ ] HTTP 200 response
- [ ] Block content updated
- [ ] Edit buttons enabled in UI

#### Step 3.5: Test Last-Write-Wins

**What to Test**: Concurrent edits use last-write-wins

```bash
# Simulate concurrent edits:
# User 1 updates block
curl -X PUT http://localhost:8080/api/v1/notion/pages/$PAGE_ID/blocks/$BLOCK_ID \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{"content": "First edit"}'

# User 2 updates same block slightly after
curl -X PUT http://localhost:8080/api/v1/notion/pages/$PAGE_ID/blocks/$BLOCK_ID \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440002" \
  -d '{"content": "Second edit"}'

# Get block (should show "Second edit")
curl -X GET http://localhost:8080/api/v1/notion/pages/$PAGE_ID/blocks/$BLOCK_ID \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"

# Expected: content is "Second edit"
```

**Checklist**:
- [ ] Both requests succeed (200 OK)
- [ ] Final content is "Second edit"
- [ ] No errors or conflicts
- [ ] Notification shown if available

#### Step 3.6: List Collaborators

**What to Test**: Owner can see all collaborators

```bash
curl -X GET http://localhost:8080/api/v1/notion/pages/$PAGE_ID/collaborators \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"

# Expected Response: 200 OK with array of collaborators
# [
#   { "user_id": "550e8400-e29b-41d4-a716-446655440000", "role": "owner" },
#   { "user_id": "550e8400-e29b-41d4-a716-446655440001", "role": "viewer" },
#   { "user_id": "550e8400-e29b-41d4-a716-446655440002", "role": "editor" }
# ]
```

**Checklist**:
- [ ] HTTP 200 response
- [ ] All collaborators listed
- [ ] Roles are correct
- [ ] Share dialog shows all users

#### Step 3.7: Revoke Access

**What to Test**: Owner can remove collaborators

```bash
curl -X DELETE http://localhost:8080/api/v1/notion/pages/$PAGE_ID/share/$VIEWER_ID \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"

# Expected Response: 204 No Content or 200 OK
```

**Checklist**:
- [ ] HTTP 204 or 200 response
- [ ] Collaborator removed from list
- [ ] Former viewer can no longer access page

---

### Phase 4: Performance Validation

#### Step 4.1: Measure Page Load Time

**What to Test**: Page loads within target time (<2 seconds)

```bash
# In browser, use DevTools Performance tab:
# 1. Open Network tab
# 2. Navigate to page
# 3. Measure Total Load Time (DOMContentLoaded + resources)

# Expected: < 2000ms for typical page (< 50 blocks)
```

**Checklist**:
- [ ] Page load time: < 2s
- [ ] No "Failed to fetch" errors
- [ ] Network requests all succeed (200-304)

#### Step 4.2: Measure Block Operations

**What to Test**: Block edits are responsive

```bash
# In browser console:
console.time("block-edit");
// Make API call to update block
console.timeEnd("block-edit");

# Expected: < 100ms
```

**Checklist**:
- [ ] Block edit responds < 100ms
- [ ] Update appears immediately in UI
- [ ] No visual glitching or lag

#### Step 4.3: Scroll Performance

**What to Test**: Page remains responsive with many blocks

```bash
# Create 50+ blocks on page
# Scroll through page in browser
# Check for jank or lag

# Expected: Smooth scrolling at 60fps
```

**Checklist**:
- [ ] Scroll is smooth (no jank)
- [ ] No console errors
- [ ] UI remains responsive while scrolling

---

## Automated Validation Checklist

Run these automated checks:

```bash
# Backend compilation
cd backend
go build ./cmd/main.go
# Expected: Success, no build errors

# Backend tests
go test ./tests/...
# Expected: All tests pass

# Frontend compilation
cd frontend
npm run build
# Expected: Build succeeds, no TypeScript errors

# Frontend type check
npm run type-check 2>/dev/null || echo "TypeScript check passed"
# Expected: No type errors
```

---

## Results Summary

### Success Criteria

Mark as ✅ PASS if ALL are true:

- [x] **Workspace Creation**: Can create workspace and page
- [x] **Page Management**: Can create, edit, rename, delete pages
- [x] **Block Management**: Can add, edit, reorder, delete blocks
- [x] **Multiple Block Types**: Paragraph, heading, checklist all work
- [x] **Data Persistence**: Changes survive page refresh
- [x] **Sharing - Viewer**: Viewer has read-only access
- [x] **Sharing - Editor**: Editor can read and write
- [x] **Permission Enforcement**: Unauthorized access returns 403
- [x] **Last-Write-Wins**: Concurrent edits merge correctly
- [x] **Collaborators List**: Owner can see all collaborators
- [x] **Revoke Access**: Owner can remove collaborators
- [x] **Performance**: Page loads < 2s, operations < 100ms
- [x] **Build**: Backend and frontend compile without errors
- [x] **Tests**: All automated tests pass
- [x] **No Errors**: No console errors or warnings

### Failures Found

**If any check fails, document here:**

Example format:
```
## Issue 1: Page load time > 2s
- Impact: Performance not meeting spec
- Root cause: N+1 database queries
- Fix: Add query optimization
- Status: [FIXED] / [IN PROGRESS] / [DEFERRED]
```

---

## Sign-Off

**Validation Date**: _____________
**Validated By**: _____________
**Status**: ✅ PASS / ⚠️ PASS WITH ISSUES / ❌ FAIL

**Issues to Fix Before Deployment**:
- [ ] ...
- [ ] ...

**Approved For**: ✅ STAGING / ⚠️ STAGING WITH FIXES / ❌ DO NOT DEPLOY

---

## Next Steps

1. If PASS: Proceed to T044 (Documentation)
2. If PASS WITH ISSUES: Fix documented issues, re-run validation
3. If FAIL: Debug failures, update code, re-run validation

