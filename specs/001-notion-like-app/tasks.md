# Tasks: Notion-like workspace

**Input**: Design documents from `/specs/001-notion-like-app/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are REQUIRED for P1 and high-risk user stories in line with the constitution, and SHOULD be included whenever the feature spec or risk profile indicates they are needed.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Web app**: `backend/src/`, `backend/tests/`, `frontend/src/`, `frontend/tests/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create backend and frontend project structure per implementation plan (`backend/`, `frontend/`)
- [x] T002 Initialize backend project with chosen Go web framework and dependencies in `backend/`
- [x] T003 Initialize frontend project with chosen TypeScript framework and tooling in `frontend/`
- [x] T004 [P] Configure backend linting and formatting (e.g., golangci-lint) in `backend/`
- [x] T005 [P] Configure frontend linting and formatting (ESLint + Prettier) in `frontend/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T006 Setup database connection and migration framework for PostgreSQL in `backend/src/`
- [x] T007 [P] Define core domain models (Workspace, Page, Block, SharePermission) in `backend/src/models/`
- [x] T008 [P] Implement basic error handling and logging middleware in `backend/src/api/`
- [x] T009 [P] Setup API routing structure for workspace, pages, blocks, and sharing in `backend/src/api/`
- [x] T010 Configure environment-based settings (database URL, ports, feature flags) in `backend/src/`
- [x] T011 Setup frontend routing and base layout (workspace shell + page view) in `frontend/src/pages/`
- [x] T012 [P] Implement shared UI components scaffold (sidebar, page shell, block list) in `frontend/src/components/`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Create and organize personal workspace (Priority: P1) 🎯 MVP

**Goal**: Logged-in users can create, edit, and organize pages in a personal workspace with persisted hierarchy.

**Independent Test**: A single user can sign in, create a workspace page, add content blocks, move pages in the sidebar, and see all changes persisted across refresh and new sessions.

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T013 [P] [US1] Backend integration test for workspace and page CRUD in `backend/tests/integration/workspace_page_crud_test.go`
- [x] T014 [P] [US1] Backend contract test for workspace/page API in `backend/tests/contract/workspace_page_api_test.go`
- [x] T015 [P] [US1] Frontend integration test for creating and organizing pages in `frontend/tests/integration/workspace_page_flow.test.tsx`

### Implementation for User Story 1

- [x] T016 [P] [US1] Implement workspace and page models/repositories in `backend/src/models/workspace_page.go`
- [x] T017 [US1] Implement workspace and page services (create, rename, delete, reorder, nesting) in `backend/src/services/workspace_page_service.go`
- [x] T018 [US1] Implement workspace and page HTTP handlers in `backend/src/api/workspace_page_handlers.go`
- [x] T019 [P] [US1] Implement sidebar workspace navigation (list, create, reorder, nesting) in `frontend/src/components/WorkspaceSidebar.tsx`
- [x] T020 [P] [US1] Implement page shell view (title editing, basic content placeholder) in `frontend/src/pages/WorkspacePageView.tsx`
- [x] T021 [US1] Wire frontend workspace/page views to backend APIs using a client in `frontend/src/services/workspaceApi.ts`
- [x] T022 [US1] Add validation and UX feedback for page create/rename/delete errors in `frontend/src/components/WorkspaceSidebar.tsx`
- [x] T023 [US1] Add logging for workspace/page operations in `backend/src/services/workspace_page_service.go`

**Checkpoint**: User Story 1 fully functional and testable independently

---

## Phase 4: User Story 2 - Rich content editing with blocks (Priority: P2)

**Goal**: Users can compose pages from rich content blocks (text, headings, checklists, etc.) and manage them easily.

**Independent Test**: For an existing page, a user can add, edit, reorder, and delete blocks of different types, see the correct formatting, and have all changes persist.

### Tests for User Story 2 ✅

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T024 [P] [US2] Backend contract test for block CRUD API in `backend/tests/contract/block_api_test.go`
- [X] T025 [P] [US2] Backend integration test for block ordering and persistence, including last-write-wins behavior, in `backend/tests/integration/block_flow_test.go`
- [X] T026 [P] [US2] Frontend integration test for block editing and reordering in `frontend/tests/integration/block_editor_flow.test.tsx`
- [X] T050 [P] [US2] Frontend integration test verifying concurrent edits apply last-write-wins at block level with user feedback in `frontend/tests/integration/concurrent_edits_ui.test.tsx`

### Implementation for User Story 2 ✅

- [X] T027 [P] [US2] Implement Block model/repository methods (CRUD, ordering) in `backend/src/models/block.go`
- [X] T028 [US2] Implement block services (create, update, delete, reorder, type-specific behavior) in `backend/src/services/block_service.go`
- [X] T029 [US2] Implement block HTTP handlers in `backend/src/api/block_handlers.go`
- [X] T030 [P] [US2] Implement block editor UI components (paragraph, heading, checklist) in `frontend/src/components/BlockEditor.tsx`
- [X] T031 [P] [US2] Implement drag-and-drop or other reordering UX for blocks in `frontend/src/components/BlockList.tsx`
- [X] T032 [US2] Wire block editor to backend APIs via `frontend/src/services/blockApi.ts`
- [X] T033 [US2] Add UX feedback for block-level errors (failed save, validation) in `frontend/src/components/BlockEditor.tsx`
- [X] T051 [US2] Implement last-write-wins conflict handling and overwrite notification for blocks in `frontend/src/components/BlockEditor.tsx`

**Checkpoint**: User Stories 1 AND 2 both work independently

---

## Phase 5: User Story 3 - Collaborative sharing and permissions (Priority: P3)

**Goal**: Workspace owners can share pages with other users (view or edit) and permissions are correctly enforced.

**Independent Test**: One user can grant another user access to a page, and the invited user can see the page in their list and interact according to their permission level.

### Tests for User Story 3 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T034 [P] [US3] Backend contract test for sharing and permissions API in `backend/tests/contract/sharing_api_test.go`
- [X] T035 [P] [US3] Backend integration test for permission enforcement (view vs edit) in `backend/tests/integration/sharing_permissions_flow_test.go`
- [X] T036 [P] [US3] Frontend integration test for share UI and permission behavior in `frontend/tests/integration/sharing_flow.test.tsx`

### Implementation for User Story 3

- [X] T037 [P] [US3] Implement SharePermission model/repository methods in `backend/src/models/share_permission.go`
- [X] T038 [US3] Implement sharing/permissions services (grant/revoke, role checks) in `backend/src/services/sharing_service.go`
- [X] T039 [US3] Implement sharing HTTP handlers (create/revoke shares, list collaborators) in `backend/src/api/sharing_handlers.go`
- [X] T040 [P] [US3] Implement share dialog UI (select user, choose role) in `frontend/src/components/ShareDialog.tsx`
- [X] T041 [P] [US3] Integrate share entry points into page UI (e.g., share button) in `frontend/src/pages/WorkspacePageView.tsx`
- [X] T042 [US3] Enforce view vs edit permissions in frontend (disable editing for viewers) in `frontend/src/shared/PermissionGuard.tsx`
- [X] T043 [US3] Enforce permissions in backend services/handlers and return appropriate errors in `backend/internal/workspace/middleware/permission_middleware.go`

**Checkpoint**: All user stories are independently functional with correct permission enforcement

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T044 [P] Documentation updates for Notion-like workspace feature in `specs/001-notion-like-app/`
- [x] T045 Code cleanup and refactoring across backend and frontend to improve readability and reuse
- [x] T046 Performance optimization for large pages and high-traffic scenarios to meet SC-002 and SC-004 in `backend/src/` and `frontend/src/`
- [x] T047 [P] Additional unit tests for critical services (workspace, blocks, sharing) in `backend/tests/unit/`
- [x] T048 Security hardening for authorization checks and data exposure in `backend/src/` to support SC-004 and overall reliability targets
- [x] T049 Run quickstart validation flow and fix any observed UX or reliability issues to validate SC-001 and SC-003 as described in `specs/001-notion-like-app/quickstart.md`
- [x] T052 [P] Implement accessibility improvements (keyboard navigation, ARIA labels, contrast) for workspace sidebar and page views in `frontend/src/components/` and `frontend/src/pages/`
- [ ] T053 [P] Implement accessibility improvements and basic a11y checks for block editor interactions in `frontend/src/components/BlockEditor.tsx`
- [ ] T054 [P] Add backend metrics collection for page load latency, error rates, and permission failures in `backend/src/` and expose them to monitoring

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Reuses workspace/page data but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Integrates with pages but should be independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Models before services
- Services before endpoints
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently using quickstart steps and tests
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently


