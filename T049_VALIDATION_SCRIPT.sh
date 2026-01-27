#!/bin/bash

##############################################################################
# T049: Quickstart Validation Script
# Tests end-to-end user flows for Notion-like workspace MVP
##############################################################################

set -e

API_BASE="http://localhost:8000/api/v1"
FRONTEND_BASE="http://localhost:3000"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test utilities
test_start() {
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "${BLUE}[TEST $TESTS_RUN]${NC} $1"
}

test_pass() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo -e "${GREEN}✓ PASS:${NC} $1"
}

test_fail() {
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo -e "${RED}✗ FAIL:${NC} $1"
}

test_section() {
    echo ""
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}$1${NC}"
    echo -e "${YELLOW}========================================${NC}"
}

# Step 1: Health Check
test_section "STEP 0: System Health Check"

test_start "Backend is running and healthy"
if curl -s http://localhost:8000/health > /dev/null 2>&1; then
    test_pass "Backend responding at http://localhost:8000"
else
    test_fail "Backend not responding"
    exit 1
fi

test_start "Frontend is running"
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    test_pass "Frontend responding at http://localhost:3000"
else
    test_fail "Frontend not responding"
    exit 1
fi

# Step 1: Sign Up / Login
test_section "STEP 1: User Authentication (Sign In)"

TEST_EMAIL="testuser_$(date +%s)@example.com"
TEST_PASSWORD="TestPass123!@"

test_start "Sign up with test user"
SIGNUP_RESPONSE=$(curl -s -X POST "$API_BASE/auth/signup" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\",
        \"name\": \"Test User $(date +%s)\"
    }")

USER_ID=$(echo $SIGNUP_RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$USER_ID" ]; then
    test_fail "Sign up failed: $SIGNUP_RESPONSE"
else
    test_pass "User created with ID: $USER_ID"
fi

test_start "Login with test user"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\"
    }")

ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$ACCESS_TOKEN" ]; then
    test_fail "Login failed: $LOGIN_RESPONSE"
else
    test_pass "Login successful, token obtained"
fi

# Step 2: Create Workspace and Page
test_section "STEP 2: Create Page in Workspace"

test_start "Create workspace"
WS_RESPONSE=$(curl -s -X POST "$API_BASE/notion/workspaces" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"name\": \"Test Workspace\",
        \"description\": \"Workspace for quickstart validation\"
    }")

WORKSPACE_ID=$(echo $WS_RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$WORKSPACE_ID" ]; then
    test_fail "Workspace creation failed: $WS_RESPONSE"
else
    test_pass "Workspace created with ID: $WORKSPACE_ID"
fi

test_start "Create page in workspace"
PAGE_RESPONSE=$(curl -s -X POST "$API_BASE/notion/pages" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"title\": \"Test Page\",
        \"workspace_id\": \"$WORKSPACE_ID\"
    }")

PAGE_ID=$(echo $PAGE_RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$PAGE_ID" ]; then
    test_fail "Page creation failed: $PAGE_RESPONSE"
else
    test_pass "Page created with ID: $PAGE_ID"
fi

# Step 3: Add Content Blocks
test_section "STEP 3: Add and Manage Content Blocks"

test_start "Add paragraph block"
PARA_RESPONSE=$(curl -s -X POST "$API_BASE/notion/blocks" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"page_id\": \"$PAGE_ID\",
        \"type\": \"paragraph\",
        \"content\": \"This is a test paragraph block.\"
    }")

PARA_BLOCK_ID=$(echo $PARA_RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$PARA_BLOCK_ID" ]; then
    test_fail "Paragraph block creation failed: $PARA_RESPONSE"
else
    test_pass "Paragraph block created with ID: $PARA_BLOCK_ID"
fi

test_start "Add heading block"
HEADING_RESPONSE=$(curl -s -X POST "$API_BASE/notion/blocks" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"page_id\": \"$PAGE_ID\",
        \"type\": \"heading\",
        \"content\": \"Test Heading\"
    }")

HEADING_BLOCK_ID=$(echo $HEADING_RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$HEADING_BLOCK_ID" ]; then
    test_fail "Heading block creation failed: $HEADING_RESPONSE"
else
    test_pass "Heading block created with ID: $HEADING_BLOCK_ID"
fi

test_start "Add checklist block"
CHECKLIST_RESPONSE=$(curl -s -X POST "$API_BASE/notion/blocks" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"page_id\": \"$PAGE_ID\",
        \"type\": \"checklist\",
        \"content\": \"Task 1\",
        \"checked\": false
    }")

CHECKLIST_BLOCK_ID=$(echo $CHECKLIST_RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$CHECKLIST_BLOCK_ID" ]; then
    test_fail "Checklist block creation failed: $CHECKLIST_RESPONSE"
else
    test_pass "Checklist block created with ID: $CHECKLIST_BLOCK_ID"
fi

test_start "Edit paragraph block"
EDIT_RESPONSE=$(curl -s -X PUT "$API_BASE/notion/blocks/$PARA_BLOCK_ID" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"content\": \"Updated paragraph content.\"
    }")

if echo $EDIT_RESPONSE | grep -q '"id"'; then
    test_pass "Paragraph block updated"
else
    test_fail "Paragraph block update failed: $EDIT_RESPONSE"
fi

# Step 4: Verify Persistence
test_section "STEP 4: Verify Data Persistence"

test_start "Fetch page and verify blocks persist"
GET_PAGE_RESPONSE=$(curl -s -X GET "$API_BASE/notion/pages/$PAGE_ID" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

if echo $GET_PAGE_RESPONSE | grep -q '"id":"'$PAGE_ID'"'; then
    test_pass "Page fetched successfully"
else
    test_fail "Page fetch failed: $GET_PAGE_RESPONSE"
fi

test_start "List blocks for page"
LIST_BLOCKS_RESPONSE=$(curl -s -X GET "$API_BASE/notion/blocks?page_id=$PAGE_ID" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

BLOCK_COUNT=$(echo $LIST_BLOCKS_RESPONSE | grep -o '"id":"[^"]*"' | wc -l)
if [ $BLOCK_COUNT -ge 3 ]; then
    test_pass "All 3 blocks persist ($BLOCK_COUNT found)"
else
    test_fail "Expected at least 3 blocks, found $BLOCK_COUNT"
fi

# Step 5: Share Page (Requires second user)
test_section "STEP 5: Share Page with Another User"

TEST_EMAIL_2="testuser2_$(date +%s)@example.com"

test_start "Create second test user"
SIGNUP2_RESPONSE=$(curl -s -X POST "$API_BASE/auth/signup" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_EMAIL_2\",
        \"password\": \"$TEST_PASSWORD\",
        \"name\": \"Test User 2 $(date +%s)\"
    }")

USER_ID_2=$(echo $SIGNUP2_RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$USER_ID_2" ]; then
    test_fail "Second user creation failed: $SIGNUP2_RESPONSE"
else
    test_pass "Second user created with ID: $USER_ID_2"
fi

test_start "Grant viewer permission to second user"
SHARE_RESPONSE=$(curl -s -X POST "$API_BASE/notion/pages/$PAGE_ID/share" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
        \"user_id\": \"$USER_ID_2\",
        \"role\": \"viewer\"
    }")

if echo $SHARE_RESPONSE | grep -q '"id"'; then
    test_pass "Viewer permission granted"
else
    test_fail "Viewer permission grant failed: $SHARE_RESPONSE"
fi

# Step 6: Verify Permissions
test_section "STEP 6: Verify Permission Enforcement"

test_start "Second user can read page"
USER2_LOGIN=$(curl -s -X POST "$API_BASE/auth/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_EMAIL_2\",
        \"password\": \"$TEST_PASSWORD\"
    }")

USER2_TOKEN=$(echo $USER2_LOGIN | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$USER2_TOKEN" ]; then
    test_fail "Second user login failed"
else
    test_pass "Second user logged in"
fi

test_start "Verify viewer cannot edit blocks"
VIEWER_EDIT=$(curl -s -X PUT "$API_BASE/notion/blocks/$PARA_BLOCK_ID" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $USER2_TOKEN" \
    -d "{
        \"content\": \"Viewer trying to edit\"
    }")

if echo $VIEWER_EDIT | grep -q '"error"'; then
    test_pass "Viewer correctly denied edit access"
else
    test_fail "Viewer should not be able to edit: $VIEWER_EDIT"
fi

# Performance check
test_section "STEP 7: Performance Check"

test_start "Page load time is reasonable"
START_TIME=$(date +%s%N)
curl -s -X GET "$API_BASE/notion/pages/$PAGE_ID" \
    -H "Authorization: Bearer $ACCESS_TOKEN" > /dev/null
END_TIME=$(date +%s%N)
LOAD_TIME=$((($END_TIME - $START_TIME) / 1000000))

if [ $LOAD_TIME -lt 2000 ]; then
    test_pass "Page loaded in ${LOAD_TIME}ms (expected < 2000ms)"
else
    test_fail "Page loaded in ${LOAD_TIME}ms (slow, expected < 2000ms)"
fi

# Summary
test_section "VALIDATION SUMMARY"
echo "Total Tests Run:    $TESTS_RUN"
echo -e "Tests Passed:       ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed:       ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
