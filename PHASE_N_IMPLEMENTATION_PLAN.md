# Phase N: Polish & Cross-Cutting Concerns - Implementation Strategy

**Execution Model**: Module-Based (Organized by Package/Folder Structure)
**Date**: January 27, 2026

## Phase N Tasks Overview (T044-T054)

Total Tasks: 10 unique tasks (some with [P] parallelizable)

### Critical Path (Must Do First)
1. **T049 - Quickstart Validation** (Validates everything works end-to-end)
2. **T044 - Documentation** (Unblocks understanding for others)
3. **T045 - Code Cleanup** (Improves maintainability)

### Supporting Tasks
4. **T047 - Unit Tests** (Reliability)
5. **T048 - Security Hardening** (Production-ready)
6. **T046 - Performance** (Scale)
7. **T052, T053, T054** (Accessibility & Metrics) - Can parallelize

---

## Execution Plan: Organized by Module

### PHASE N.1: Validation & Documentation (Critical Path)

#### T049: Quickstart Validation Flow (CRITICAL FIRST)

**Purpose**: End-to-end validation that all 3 user stories work together

**What to Test**:
- [ ] User signup/signin workflow
- [ ] Create workspace → Create page
- [ ] Add blocks (paragraph, heading, checklist)
- [ ] Edit, reorder, delete blocks
- [ ] Browser refresh → Data persists
- [ ] Share page with user (viewer role)
- [ ] Verify viewer: read-only, cannot edit
- [ ] Share same page with different user (editor role)
- [ ] Verify editor: can read and edit
- [ ] Test concurrent edits → last-write-wins
- [ ] Verify performance: pages load < 2s

**Test Flow File**: `quickstart_validation.md` (Manual test guide + automated checks)
**Entry Point**: Backend API must have /health endpoint
**Expected Time**: 2-3 hours to identify gaps

---

#### T044: Documentation Updates [P] (Parallel possible)

**Module Structure**:

```
specs/001-notion-like-app/
├── README.md (NEW: Project overview)
├── API_REFERENCE.md (NEW: OpenAPI-style endpoint docs)
├── ARCHITECTURE.md (NEW: System design & patterns)
├── SETUP_GUIDE.md (NEW: Dev environment setup)
└── DEPLOYMENT.md (NEW: Production deployment)
```

**T044.1: API Reference** (`API_REFERENCE.md`)
- Document all endpoints (workspace, page, block, sharing)
- Request/response examples
- Error codes and meanings
- Auth headers required

**T044.2: Architecture Guide** (`ARCHITECTURE.md`)
- Repository → Service → Controller pattern
- Dependency injection approach
- Frontend component hierarchy
- Data flow diagrams

**T044.3: Setup Guide** (`SETUP_GUIDE.md`)
- Database setup (PostgreSQL)
- Environment variables (.env template)
- Backend build & run
- Frontend build & run
- Running tests

**Deliverable**: All documentation in specs folder, linked from main README

---

### PHASE N.2: Code Quality & Reliability

#### T045: Code Cleanup & Refactoring (Backend Priority)

**Backend Module Structure** (following internal/ pattern):

```
backend/internal/
├── api/
│   ├── helpers.go (REVIEW: Add missing helpers)
│   └── response.go (REVIEW: Ensure consistency)
├── workspace/
│   ├── repository/ (REVIEW: Interfaces, consistency)
│   ├── service/ (REVIEW: Error handling, logging)
│   ├── controller/ (REVIEW: Response formats)
│   └── middleware/ (REVIEW: Auth logic)
├── block/
│   ├── repository/ (REVIEW: Query optimization)
│   ├── service/ (REVIEW: Reorder logic)
│   └── controller/ (REVIEW: Error handling)
├── page/
│   ├── repository/ (REVIEW: Nested queries)
│   ├── service/ (REVIEW: Consistency)
│   └── controller/ (REVIEW: REST compliance)
└── errors/ (REVIEW: Error hierarchy)
```

**T045.1: Backend Service Layer**
- [ ] Remove duplicate logic in services
- [ ] Extract common patterns (Create, Update, Delete)
- [ ] Standardize error handling
- [ ] Add input validation helpers
- [ ] Review logging consistency

**T045.2: Backend Controller Layer**
- [ ] Standardize response format across all endpoints
- [ ] Consistent error responses
- [ ] Request validation helpers
- [ ] Add request ID for tracing

**T045.3: Frontend Components**
- [ ] Review BlockEditor for reusable sub-components
- [ ] Extract form helpers (input, dropdown, etc.)
- [ ] Consolidate API error handling
- [ ] Remove code duplication in hooks

**T045.4: Shared Code**
- [ ] Create utils/ package for common functions
- [ ] DRY up test helpers
- [ ] Standardize naming conventions

---

#### T047: Unit Tests [P] (Parallel execution)

**Backend Module Structure**:

```
backend/tests/unit/
├── workspace/
│   ├── service_test.go (Service business logic)
│   └── repository_test.go (Data access mocking)
├── block/
│   ├── service_test.go (Block operations)
│   └── repository_test.go (Ordering logic)
└── sharing/
    ├── service_test.go (Permission checks)
    └── repository_test.go (Query logic)
```

**T047.1: Workspace Service Unit Tests** (workspace/service_test.go)
- [ ] Test create page validation
- [ ] Test rename with conflict handling
- [ ] Test delete with cascades
- [ ] Test reorder logic
- [ ] Test permission on create

**T047.2: Block Service Unit Tests** (block/service_test.go)
- [ ] Test create block validation
- [ ] Test reorder with boundary checks
- [ ] Test delete with history
- [ ] Test type-specific behavior
- [ ] Test concurrent update handling

**T047.3: Sharing Service Unit Tests** (sharing/service_test.go)
- [ ] Test grant access validation
- [ ] Test role hierarchy (viewer < editor < owner)
- [ ] Test revoke logic
- [ ] Test permission checks (HasAccess, CanEdit)
- [ ] Test permission queries

**Target**: 70%+ coverage on critical services

---

#### T048: Security Hardening

**Backend Module Structure**:

```
backend/internal/
├── auth/
│   ├── middleware/ (REVIEW: Token validation)
│   └── helpers.go (ADD: Authorization helpers)
├── workspace/middleware/ (REVIEW: Permission checks)
└── errors/codes.go (ADD: Security error codes)
```

**T048.1: Authorization Audit** (workspace/middleware/)
- [ ] Review AuthorizePageRead logic
- [ ] Review AuthorizePageEdit logic
- [ ] Verify owner-only operations (delete, share)
- [ ] Test permission edge cases
- [ ] Add rate limiting prep

**T048.2: Input Validation** (api/helpers.go)
- [ ] Add UUID validation helper
- [ ] Add string length validation
- [ ] Add role validation (enum check)
- [ ] Review all request DTOs

**T048.3: Data Exposure** (controller layers)
- [ ] Ensure queries are scoped to user
- [ ] Verify no data leakage on errors
- [ ] Check permission checks before returns
- [ ] Audit logging for sensitive ops

**T048.4: HTTPS/CORS Ready** (config/)
- [ ] Document HTTPS requirements
- [ ] CORS configuration template
- [ ] Security header recommendations

---

### PHASE N.3: Performance & Scale

#### T046: Performance Optimization

**Backend Module Structure**:

```
backend/internal/
├── workspace/repository/ (ADD: Query optimization)
├── block/repository/ (ADD: Bulk operations)
├── redis/ (ADD: Cache layer - optional)
└── metrics/ (ADD: Timing instrumentation)
```

**T046.1: Database Query Optimization** (repositories)
- [ ] Add indexes to queries (page_id, user_id)
- [ ] Implement pagination for large lists
- [ ] Use SELECT only needed columns
- [ ] Profile slow queries (N+1 problems)
- [ ] Cache frequently accessed data

**T046.2: Frontend Performance** (src/components/)
- [ ] Implement virtualization for long block lists
- [ ] Memoize expensive components
- [ ] Add lazy loading for images
- [ ] Optimize re-renders (useCallback, useMemo)

**T046.3: API Response Optimization**
- [ ] Compress responses (gzip)
- [ ] Add ETag support for caching
- [ ] Implement incremental loading
- [ ] Add response time headers

---

### PHASE N.4: Polish & Accessibility (Parallelizable)

#### T052: Accessibility - Sidebar [P]

**Frontend: WorkspaceSidebar.tsx**
- [ ] Add keyboard navigation (arrow keys, Tab)
- [ ] Add ARIA labels for all interactive elements
- [ ] Add focus indicators (visible outline)
- [ ] Check color contrast (WCAG AA)
- [ ] Test with screen reader

#### T053: Accessibility - Block Editor [P]

**Frontend: BlockEditor.tsx**
- [ ] Add keyboard shortcuts (Cmd+B for bold, etc.)
- [ ] Add ARIA descriptions for block types
- [ ] Add focus management
- [ ] Add proper alt text for images
- [ ] Test with screen reader

#### T054: Backend Metrics [P]

**Backend Module Structure**:

```
backend/internal/
└── metrics/
    ├── metrics.go (Prometheus metrics)
    ├── middleware.go (HTTP timing)
    └── logger.go (Integration with logging)
```

**Metrics to Collect**:
- [ ] HTTP request latency (histogram)
- [ ] Request count by endpoint (counter)
- [ ] Error rate by type (counter)
- [ ] Permission check latency (histogram)
- [ ] Database query time (histogram)
- [ ] Active users/connections (gauge)

---

## Implementation Sequence

### Week 1: Foundation (Critical Path)
1. **Day 1-2**: T049 - Quickstart validation (identify gaps)
2. **Day 3**: T044 - Documentation (unblock team)
3. **Day 4-5**: T045 - Code cleanup (improve quality)

### Week 2: Reliability & Security
4. **Day 1-2**: T047 - Unit tests [P] (parallelize)
5. **Day 3-4**: T048 - Security hardening
6. **Day 5**: T046 - Performance optimization

### Week 3: Polish (Can parallelize)
7. **Day 1-2**: T052 - Accessibility sidebar [P]
8. **Day 1-2**: T053 - Accessibility editor [P]
9. **Day 1-2**: T054 - Metrics [P]

---

## Success Criteria for Each Task

### T049 ✅
- [ ] All quickstart steps can be completed without errors
- [ ] Data persists across browser refresh
- [ ] Permissions enforce correctly (viewer read-only)
- [ ] Last-write-wins works for concurrent edits
- [ ] Page loads < 2s for typical content
- [ ] No console errors or warnings

### T044 ✅
- [ ] All endpoints documented with examples
- [ ] Setup guide tested by someone new
- [ ] Architecture document reviewed
- [ ] Deployment guide complete
- [ ] All files linked from main README

### T045 ✅
- [ ] No duplicate code patterns
- [ ] Consistent error handling
- [ ] All tests still pass
- [ ] Code review: 0 comments on style
- [ ] Improved cyclomatic complexity (tools report improvement)

### T047 ✅
- [ ] 70%+ coverage on critical services
- [ ] All tests pass
- [ ] No flaky tests
- [ ] Tests document expected behavior
- [ ] Edge cases covered

### T048 ✅
- [ ] Permission checks enforced consistently
- [ ] No unauthorized data access possible
- [ ] All inputs validated
- [ ] Security team review passed
- [ ] OWASP Top 10 checklist complete

### T046 ✅
- [ ] 50% reduction in query time (baseline → target)
- [ ] Page load < 200ms for 1000 blocks
- [ ] Block operations < 100ms
- [ ] Permission checks < 50ms
- [ ] Frontend virtualizes lists > 100 items

### T052, T053, T054 ✅
- [ ] Keyboard navigation works
- [ ] Screen reader passes audit
- [ ] WCAG 2.1 AA compliance
- [ ] Metrics exposed on /metrics endpoint
- [ ] Dashboards configured

---

## File Organization After Phase N

```
GOALKeeper-Plan/
├── backend/
│   ├── internal/
│   │   ├── metrics/          ← NEW (T054)
│   │   ├── workspace/middleware/ ← ENHANCED (T048)
│   │   └── */repository/     ← OPTIMIZED (T046)
│   └── tests/
│       └── unit/             ← NEW (T047)
├── frontend/
│   ├── src/components/       ← ENHANCED (T052, T053)
│   └── tests/unit/           ← NEW (if needed)
├── specs/
│   └── 001-notion-like-app/
│       ├── README.md         ← NEW (T044)
│       ├── API_REFERENCE.md  ← NEW (T044)
│       ├── ARCHITECTURE.md   ← NEW (T044)
│       ├── SETUP_GUIDE.md    ← NEW (T044)
│       └── DEPLOYMENT.md     ← NEW (T044)
└── docs/
    └── QUICKSTART_RESULTS.md ← NEW (T049)
```

---

## Next Steps

1. Start with T049 (Quickstart validation)
2. Document findings in quickstart_validation.md
3. Fix any blocking issues before proceeding
4. Then proceed with T044 (Documentation)
5. Proceed in sequence as outlined above

Status: Ready to begin Phase N implementation
