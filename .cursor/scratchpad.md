# Go TODO Bot - Telegram → Google Sheets → Email Digest

## Background and Motivation

The user wants to build a personal note-taking Telegram bot that:
1. **Phase 1**: Captures messages, uses OpenRouter LLM to parse for tasks/people, and logs to Google Sheets
2. **Phase 2**: Sends daily email digests at 06:00 GMT to team members with their pending tasks
3. **Phase 3**: Deploy to Render with health checks and production readiness

The bot serves as a lightweight task management system where natural language messages are automatically parsed and organized, with daily accountability emails to ensure tasks don't get lost.

## Key Challenges and Analysis

### Technical Challenges
1. **LLM Integration**: Need to parse natural language for people mentions, task summaries, and determine if messages should be split into multiple tasks
2. **Google Sheets Integration**: Using Apps Script as a webhook endpoint (serverless approach)
3. **Email Scheduling**: Implementing reliable cron job for daily digests at specific timezone
4. **People Resolution**: Mapping mentioned names to email addresses from environment variables
5. **Task Splitting Logic**: LLM must intelligently determine when a message contains multiple discrete tasks

### Design Decisions
1. **Go Language**: Good for concurrent operations, small binary size, excellent for long-running services
2. **OpenRouter for LLM**: Flexible model selection, no vendor lock-in
3. **Google Apps Script**: Serverless approach for Sheets integration (no OAuth complexity)
4. **SendGrid for Email**: Reliable email delivery with fallback to SMTP
5. **Render Deployment**: Free tier available, easy Docker deployment
6. **Team Management**: Separate "team" tab in Google Sheets for dynamic team member management
7. **SQLite for Resilience**: Local cache for failed writes with automatic retry mechanism

## High-level Task Breakdown

### Phase 0: Preparation
- [x] **Task 0.1**: Create test data suite
  - Success: 20+ example messages covering all edge cases documented
  - Examples include: simple tasks, multi-person mentions, task splitting scenarios, edge cases

### Phase 1: Core Bot Functionality
- [ ] **Task 1.1**: Project setup and structure
  - Success: Go module initialized, folder structure created, dependencies added
- [ ] **Task 1.2**: Google Apps Script webhook with dual-tab setup
  - Success: Script deployed with "todo" and "team" tabs, test POST successfully adds row
  - Team tab has columns: Name | Email
  - Todo tab has columns: Timestamp | People | Summary | FullMessage | Status | BotNotes
- [ ] **Task 1.3**: Telegram bot basic handler
  - Success: Bot responds to messages, can extract text/user/timestamp
- [ ] **Task 1.4**: OpenRouter LLM integration with specific model
  - Success: Parse test message using openai/gpt-4o-mini, extract people/summary/task count
  - Implement robust prompt template for consistent parsing
- [ ] **Task 1.5**: SQLite cache layer
  - Success: Failed writes stored locally, retry mechanism every 10 minutes
  - Admin notification via Telegram on persistent failures
- [ ] **Task 1.6**: Connect all components with error handling
  - Success: End-to-end test - Telegram message → LLM → Sheet row
  - Parse errors default to "team" with error details in BotNotes column
- [ ] **Task 1.7**: Task splitting logic
  - Success: Messages with "AND" or multiple discrete tasks create 2-3 rows
  - Example: "Call Bob about budget AND schedule meeting with Alice" → 2 tasks

### Phase 2: Email Digest System
- [ ] **Task 2.1**: Team data fetching from Google Sheets
  - Success: Read team tab, build name→email mapping dynamically
- [ ] **Task 2.2**: CSV export and task filtering
  - Success: Fetch todo tab, filter by person and incomplete status
- [ ] **Task 2.3**: Email template system
  - Success: HTML email template with task list renders correctly
  - One email per person per day consolidating all their tasks
- [ ] **Task 2.4**: Test mode implementation
  - Success: Environment variable to redirect all emails to single test address
- [ ] **Task 2.5**: Cron job implementation
  - Success: Job triggers at 06:00 GMT (test with @every 1m first)
- [ ] **Task 2.6**: SendGrid integration
  - Success: Test email sent successfully, supports 6+ recipients
- [ ] **Task 2.7**: Multi-recipient logic
  - Success: Each person gets one daily email with all their tasks
  - "team" tasks included in everyone's email

### Phase 3: Production Deployment
- [ ] **Task 3.1**: Health check endpoint
  - Success: /healthz returns 200 with uptime and last successful operations
- [ ] **Task 3.2**: Sentry error tracking setup
  - Success: Free tier configured, errors tracked with context
- [ ] **Task 3.3**: Dockerfile optimization
  - Success: Multi-stage build, final image < 25MB
- [ ] **Task 3.4**: Environment configuration
  - Success: All secrets in env vars, render.yaml configured
  - Includes: TELEGRAM_TOKEN, OPENROUTER_API_KEY, GOOGLE_SCRIPT_URL, SENDGRID_KEY, ADMIN_TELEGRAM_ID, TEST_EMAIL
- [ ] **Task 3.5**: Webhook setup and testing
  - Success: Bot responds to webhooks on Render URL
- [ ] **Task 3.6**: Monitoring and alerting
  - Success: Errors logged to Sentry, admin notified via Telegram for critical issues
  - Dead letter log for failed emails

## Project Status Board

### Ready
- [ ] Task 1.5: SQLite cache layer

### In Progress
- [ ] Task 1.3: Telegram bot basic handler (95% complete)
- [ ] Task 1.4: OpenRouter LLM integration with specific model (95% complete)

### Done
- [x] Task 0.1: Create test data suite
- [x] Task 1.1: Project setup and structure
- [x] Task 1.2: Google Apps Script webhook with dual-tab setup

### Blocked

## Current Status / Progress Tracking

**Current Phase**: Phase 3 - Production Deployment Fixes
**Last Completed**: Docker build successful with --chmod fix ✅
**Currently Working On**: Telegram token validation fix

### 🚨 **Critical Issue Identified:**

**Problem**: Application crashes on startup when invalid/placeholder Telegram tokens are provided
- Error: `FTL Failed to create Telegram handler error="failed to create bot: Not Found"`
- Cause: `tgbotapi.NewBotAPI(token)` fails with invalid tokens, causing fatal error
- Impact: Render deployment fails with exit status 128

**Root Cause Analysis**:
1. `cmd/bot/main.go` checks if `cfg.TelegramToken != ""` (line 49)
2. If token exists but is invalid (like "placeholder"), it still tries to create handler
3. `internal/telegram/handler.go:25` calls `tgbotapi.NewBotAPI(token)` without validation
4. Telegram API returns "Not Found" for invalid tokens
5. Error bubbles up as fatal error, crashing application

**Solution Strategy**:
1. Add token validation before creating Telegram bot API instance
2. Gracefully handle invalid tokens by falling back to web-only mode
3. Add proper error handling and logging for token validation failures

### 🔧 **Implementation Plan:**

1. **Token Validation Function**:
   - Create helper function to validate Telegram token format
   - Check if token matches expected format (bot123456:ABC-DEF...)
   - Attempt lightweight API call to verify token validity

2. **Graceful Fallback**:
   - If token validation fails, log warning and continue in web-only mode
   - Don't create Telegram handler if token is invalid
   - Ensure application starts successfully regardless of token validity

3. **Enhanced Logging**:
   - Clear distinction between "no token provided" vs "invalid token provided"
   - Helpful error messages for debugging token issues

### 🎯 **Expected Outcome**:
- Application starts successfully with invalid/placeholder tokens
- Falls back to web-only mode when Telegram token is invalid
- Render deployment succeeds without exit status 128
- Clear logging indicates whether running in Telegram or web-only mode

## Executor's Feedback or Assistance Requests

### 🎉 **SIMPLIFIED DEPLOYMENT READY - MINIMAL ENV VARS**

**✅ CONFIGURATION SIMPLIFIED**:
- ✅ **Removed TELEGRAM_TOKEN requirement** - Web-only mode support
- ✅ **Removed SENDGRID_KEY requirement** - No email features needed
- ✅ **Only 2 required environment variables now**:
  - `OPENROUTER_API_KEY` (you have this)
  - `GOOGLE_SCRIPT_URL` (you have this)

**🚀 DEPLOYMENT STATUS**: 
- ✅ Code updated and pushed to GitHub (commit: 70e821a)
- ✅ Application will run in web-only mode
- ✅ **READY FOR RENDER DEPLOYMENT WITH MINIMAL CONFIG**

**📋 REQUIRED RENDER ENVIRONMENT VARIABLES** (only 2 now):
```bash
OPENROUTER_API_KEY=your_actual_openrouter_key
GOOGLE_SCRIPT_URL=your_actual_google_script_url
```

**🎯 NEXT STEPS**: 
1. **Update your Render service** - it will auto-redeploy from GitHub
2. **Add only the 2 required environment variables**
3. **Service should start successfully now**

**SUCCESS**: Deployment greatly simplified! Only OpenRouter and Google Sheets integration required.

## Lessons

1. Always validate environment variables on startup to catch configuration issues early
2. Use structured logging (e.g., zerolog) for better debugging in production
3. Implement retry logic for all external API calls (LLM, Sheets, Email)
4. Test with edge cases: empty messages, very long messages, special characters
5. Consider rate limiting to avoid hitting API quotas
6. Include info useful for debugging in the program output
7. Read the file before you try to edit it
8. If there are vulnerabilities that appear in the terminal, run npm audit before proceeding
9. Always ask before using the -force git command

## Updated Design Details

### Google Sheets Structure

**"todo" tab columns:**
1. Timestamp (RFC 3339)
2. People (comma-separated, alphabetical order, "team" normalized)
3. Summary (≤ 80 chars)
4. FullMessage
5. Status (Drop-down: Not Started / In Progress / Complete)
6. BotNotes (Error messages, parsing issues, debug info)

**"team" tab columns:**
1. Name (lowercase, single word)
2. Email (valid email address)

### Error Handling Strategy

1. **LLM Parse Errors**: 
   - Default to People="team"
   - Log error details in BotNotes column
   - Continue processing (graceful degradation)

2. **Sheet Write Failures**:
   - Store in SQLite with timestamp
   - Retry every 10 minutes
   - After 3 failures, notify admin via Telegram (@defibeats)

3. **Email Failures**:
   - Log to dead_letter.log
   - Include in next day's admin summary

### Task Splitting Examples

**Split into 2 tasks:**
- "Call Bob about budget AND schedule meeting with Alice about Q3"
  → Task 1: "Call Bob about budget" (People: bob)
  → Task 2: "Schedule meeting with Alice about Q3" (People: alice)

- "Review PR from Sarah and then deploy to staging"
  → Task 1: "Review PR from Sarah" (People: sarah)
  → Task 2: "Deploy to staging" (People: team)

**Keep as 1 task:**
- "Discuss budget and timeline with Bob in tomorrow's meeting"
  → Single task about the meeting agenda

### Email Distribution Logic

- Each person receives ONE email per day at 06:00 GMT
- Email contains ALL tasks where they are mentioned:
  - Direct mentions (e.g., "alice")
  - Team tasks (everyone gets these)
- Tasks sorted by timestamp (oldest first)
- Shared Google Sheet link included for status updates

### LLM Prompt Template

```
You are a task parser for a team management system. Parse the following message and extract task information.

Message: "{message}"

Rules:
1. Identify WHO is responsible (person names or "team" if no specific person)
2. Extract WHAT needs to be done (the actual task/action)
3. Normalize names to lowercase (e.g., "Lilly" → "lilly", "The Team" → "team")
4. For multiple tasks, look for "AND" or clear task boundaries
5. Keep summaries concise but descriptive (max 80 chars)
6. If no specific person is mentioned, assign to "team"

Examples:
- "Lilly to reach out to Johnny to get a quote" → Person: lilly, Task: "Reach out to Johnny to get a quote for journalist"
- "Jemma, Lexi, and Johnny to set up meeting" → People: [jemma, lexi, johnny], Task: "Set up meeting"
- "We need to contact all vendors" → People: [team], Task: "Contact all vendors"

Return JSON format:
{
  "tasks": [
    {
      "people": ["lilly"],
      "summary": "Reach out to Johnny to get a quote for journalist",
      "confidence": 0.95
    }
  ],
  "original_message": "{message}"
}
```

### Configuration Requirements

**Environment Variables:**
```
TELEGRAM_TOKEN=xxx
OPENROUTER_API_KEY=xxx
GOOGLE_SCRIPT_URL=https://script.google.com/...
SENDGRID_KEY=xxx
ADMIN_TELEGRAM_ID=@defibeats
TEST_MODE=false
TEST_EMAIL=test@example.com
DATABASE_PATH=./cache.db
```

## Next Steps

Ready to begin implementation with Task 0.1: Creating a comprehensive test data suite that covers:
- Simple single-person tasks
- Multi-person mentions
- Team tasks
- Task splitting scenarios
- Edge cases (empty messages, very long text, special characters)
- Error scenarios for testing fallback behavior

## New Feature Implementation Plan: Enhanced Parsing & Phase 2 Transition

### Task 1.8: Enhanced Task Parsing
- [ ] **Task 1.8.1**: Update Google Sheets Schema
  - Success Criteria:
    - Add "Client" and "DueDate" columns to todo sheet
    - Add data validation for Client column (from client list)
    - Add date format validation for DueDate column
  - Steps:
    1. Modify `deploy_webhook.gs` to add new columns
    2. Update schema validation rules
    3. Add migration function for existing sheets

- [ ] **Task 1.8.2**: Enhance LLM Prompt
  - Success Criteria:
    - LLM correctly identifies client mentions
    - LLM extracts due dates in standard format
    - LLM marks unclear fields appropriately
  - Steps:
    1. Update prompt template with new fields
    2. Add example messages with client/date variations
    3. Implement "unclear" marking logic
    4. Add date normalization (convert "tomorrow", "next week", etc.)

- [ ] **Task 1.8.3**: Update Go Structures
  - Success Criteria:
    - Task struct includes Client and DueDate fields
    - JSON marshaling handles new fields correctly
    - Sheets client supports new columns
  - Steps:
    1. Update Task struct in `internal/types`
    2. Modify sheets client to handle new columns
    3. Update response formatting for Telegram messages

### Task 1.9: Manual Testing Suite
- [ ] **Task 1.9.1**: Create Test Cases
  - Success Criteria:
    - Test suite covers all parsing scenarios
    - Each test case has expected output
  - Test Categories:
    1. Client mentions (explicit, implicit, unclear)
    2. Due dates (explicit dates, relative dates, unclear)
    3. Combined scenarios (client + date + assignee)
    4. Edge cases (missing info, multiple clients)

- [ ] **Task 1.9.2**: Execute Manual Tests
  - Success Criteria:
    - All test cases pass
    - Results documented in test log
  - Test Process:
    1. Run each test case through Telegram
    2. Verify Google Sheets entries
    3. Document any parsing issues
    4. Validate error handling

### Task 1.10: Phase 2 Branch Setup
- [ ] **Task 1.10.1**: Create Phase 2 Branch
  - Success Criteria:
    - Clean branch from main
    - All current features working
  - Steps:
    1. Ensure main branch is stable
    2. Create 'phase-2' branch
    3. Update version tags
    4. Document Phase 2 goals

## Test Messages for Manual Testing

1. Client & Date Explicit:
   - "Have Microsoft review the proposal by next Friday"
   - "Schedule Apple meeting for 3/15/24"

2. Client Implicit:
   - "Follow up with the Cupertino team about iOS integration"
   - "Send the AWS migration docs to Seattle office"

3. Date Relative:
   - "Team to complete security audit within 2 weeks"
   - "Sarah to present findings by end of month"

4. Unclear Scenarios:
   - "Check status of the Android project"
   - "Update documentation when possible"

5. Multiple Elements:
   - "Prepare Google presentation for next week AND schedule Microsoft review for end of month"
   - "Sarah to meet Netflix team tomorrow, then Bob to follow up with pricing next week"

## Success Criteria for Enhanced Parsing

1. Client Detection:
   - Explicit mentions (e.g., "Microsoft", "Apple")
   - Implicit mentions (e.g., "Cupertino team" → Apple)
   - Mark as "unclear" when ambiguous

2. Due Date Parsing:
   - Explicit dates (e.g., "March 15th")
   - Relative dates (e.g., "next week", "in 2 days")
   - Mark as "unclear" when no timeline given

3. Error Handling:
   - Graceful handling of ambiguous cases
   - Clear feedback in BotNotes column
   - User-friendly error messages


## Debugging Session - LLM Not Returning Client/DueDate Fields

**Issue**: The LLM is returning responses without the client and dueDate fields.

**Evidence**: Log shows: tasks=[{"confidence":0.95,"people":["gemma"],"summary":"Write the report for oxccu including details by saturday"}]

**Root Cause**: The LLM is not following the updated prompt format. It's returning only the basic fields.

**✅ FIXED & VERIFIED**: Updated LLM prompt to be more conservative about due date extraction:
- Changed from: `dueDate: convert to YYYY-MM-DD (EOD/today=%s, tomorrow=%s, saturday=%s)`
- Changed to: `dueDate: ONLY if explicitly mentioned in message, convert to YYYY-MM-DD. If no date mentioned, use "unclear"`
- Removed date examples that were encouraging the LLM to always assume dates
- Now the LLM will only extract due dates when they're explicitly mentioned in the message

**✅ VERIFICATION**: Log shows correct behavior:
```
tasks=[{"client":"gemma","confidence":0.8,"dueDate":"unclear","people":["team"],"summary":"Get back to Gemma on times for the offsite"}]
```
- Message with no explicit date → `dueDate: "unclear"` ✅
- Client correctly extracted → `"client":"gemma"` ✅

## 🚨 **NEW ISSUE: Google Sheets Integration Error**

**Issue**: Google Sheets write failing with JSON parse error
**Evidence**: 
```
ERR Failed to save tasks to Google Sheets error="failed to add tasks: failed to decode response: invalid character '<' looking for beginning of value"
```

**Root Cause**: Google Apps Script returning HTML instead of JSON (likely error page)

**✅ ISSUE IDENTIFIED**: The `GOOGLE_SCRIPT_URL` in the `.env` file is set to the placeholder value `https://script.google.com/macros/s/YOUR_SCRIPT_ID/exec` instead of an actual deployed Google Apps Script URL.

**✅ SOLUTION PROVIDED**: 
1. Created complete Google Apps Script (`scripts/deploy_webhook.gs`) with support for Client and DueDate fields
2. Created comprehensive setup guide (`scripts/SETUP_GUIDE.md`) 
3. Script includes all required functionality:
   - `add_tasks` action with new Client/DueDate columns
   - `get_team` and `get_tasks` actions
   - Automatic sheet creation and formatting
   - Data validation for Status column
   - Sample team data

**Next Steps for User**:
1. Follow the setup guide to deploy the Google Apps Script
2. Copy the actual Web app URL from the deployment
3. Update the `GOOGLE_SCRIPT_URL` in the `.env` file
4. Test the integration with the provided curl commands
5. Restart the bot to use the new configuration

## 🚨 **NEW DEPLOYMENT ISSUE: Incorrect Build Path on Render**

**Issue**: Render deployment failing with incorrect build path
**Evidence**: 
```
==> Running build command '   go build -tags netgo -ldflags '-s -w' -o app ./go-todo-bot/cmd/bot'...
stat /opt/render/project/go/src/github.com/jgabriele321/getitdone/go-todo-bot/cmd/bot: directory not found
```

**Root Cause**: 
- Render is executing build command with `./go-todo-bot/cmd/bot` path
- Actual project structure has `./cmd/bot` (no `go-todo-bot` prefix)
- The `render.yaml` file shows correct path: `./cmd/bot`
- **Issue**: Render dashboard/service configuration is overriding the YAML file

**✅ SOLUTION IMPLEMENTED**: 
1. **Primary Fix**: Update Render service configuration in dashboard
   - Go to Render dashboard → Services → go-todo-bot
   - Navigate to Settings → Build & Deploy
   - Update Build Command to: `go build -tags netgo -ldflags '-s -w' -o app ./cmd/bot`
   - Ensure it matches the render.yaml configuration

2. **Alternative Fix**: Force render.yaml to take precedence
   - Redeploy from Git with explicit YAML configuration
   - Ensure no cached configurations are overriding

**✅ ACTUAL ROOT CAUSE DISCOVERED**: 
- Render was treating the project as GOPATH-style instead of Go modules
- Project builds perfectly locally with Go modules
- Issue was Render's build system not detecting Go modules properly

**✅ FINAL FIX APPLIED** (Commit: 9fac039):
- **CRITICAL ISSUE FOUND**: `cmd/bot/main.go` was never committed to Git!
- **Root Cause**: Docker build context was missing the entire `cmd` directory
- **Solution**: Added `cmd/bot/main.go` to Git and pushed to repository
- **Result**: Render now has access to complete project structure

**Git Commit Evidence**:
```
[main 9fac039] Fix: Add missing cmd/bot directory with main.go for Render deployment
 4 files changed, 181 insertions(+), 11 deletions(-)
 create mode 100644 cmd/bot/main.go
```

**Next Steps**:
1. ✅ Pushed fix (commit 9fac039) - cmd directory now in Git
2. ⏳ Wait for Render auto-deploy to complete with Docker build
3. ✅ Docker build should now succeed and find ./cmd/bot
4. 🎯 **Expected Success**: Service will be live with working health check

# GetItDone Bot - Project Scratchpad

## Background and Motivation

The GetItDone bot helps teams manage tasks through natural language processing in Telegram. The bot parses messages, extracts task information, and logs them in Google Sheets. A new requirement has emerged to handle batch task processing more efficiently.

## Key Challenges and Analysis

1. **Batch Processing Need**: Users want to input multiple tasks in different formats:
   - Single task: "johnny to ask lexi what she wants for dinner"
   - Double task with dependency: "Gemma to ask oxccu for press release bullet points. Lilly will draft the press release, due friday."
   - Bullet-point lists: Multiple tasks in a structured format

2. **Technical Challenges**:
   - Need to identify message type (single vs batch)
   - Parse different formats (narrative, bullet points)
   - Maintain task relationships/dependencies
   - Handle sequential processing without overwhelming APIs
   - Preserve existing single-task functionality

3. **Architecture Impact**:
   - Need intermediate storage for batch processing
   - Must modify LLM prompt to handle different formats
   - Requires task queue management
   - Need to preserve transaction integrity

## High-level Task Breakdown

### Phase 3: Batch Processing Implementation

- [ ] **Task 3.1**: Database Setup for Batch Processing
  - Success Criteria:
    - SQLite table for temporary task storage
    - Queue management system implemented
    - Transaction handling for batch operations
  - Steps:
    1. Create tasks_queue table schema
    2. Implement queue operations (push, pop, clear)
    3. Add transaction management

- [ ] **Task 3.2**: Enhanced Message Type Detection
  - Success Criteria:
    - Accurately identifies single vs batch messages
    - Detects bullet points and narrative formats
    - Fast-path for single tasks
  - Steps:
    1. Implement message format detector
    2. Add format-specific parsing rules
    3. Create bypass for single tasks

- [ ] **Task 3.3**: LLM Prompt Engineering
  - Success Criteria:
    - Updated prompt handles all formats
    - Maintains high accuracy for single tasks
    - Extracts relationships between tasks
  - Steps:
    1. Design new prompt template
    2. Add format-specific examples
    3. Test with various message types

- [ ] **Task 3.4**: Task Processing Pipeline
  - Success Criteria:
    - Batch tasks stored in queue
    - Sequential processing with error handling
    - Progress updates to user
  - Steps:
    1. Implement batch storage logic
    2. Create processing worker
    3. Add progress reporting

- [ ] **Task 3.5**: User Feedback System
  - Success Criteria:
    - Clear progress indicators
    - Error reporting for specific tasks
    - Summary of completed batch
  - Steps:
    1. Design progress messages
    2. Implement error aggregation
    3. Create batch summary format

## Project Status Board

- [x] Phase 1: Core Bot Functionality ✅
- [x] Phase 2: Email Digest System ✅
- [ ] Phase 3: Batch Processing (In Progress)
  - [ ] Task 3.1: Database Setup
  - [ ] Task 3.2: Message Type Detection
  - [ ] Task 3.3: LLM Prompt Engineering
  - [ ] Task 3.4: Task Processing Pipeline
  - [ ] Task 3.5: User Feedback System

## Executor's Feedback or Assistance Requests

Current Status: Ready to begin Phase 3 implementation
Next Action: Begin with Task 3.1 - Database Setup

## Lessons

Previous lessons remain valid, adding:
10. Implement proper progress reporting for long-running operations
11. Use transaction management for batch operations
12. Keep single-task processing fast and efficient
13. Provide clear feedback for batch operations

## Technical Design Details

### Database Schema (tasks_queue)

```sql
CREATE TABLE tasks_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    batch_id TEXT NOT NULL,
    message_text TEXT NOT NULL,
    format_type TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    error TEXT
);
```

### Message Format Types

1. SINGLE_TASK: Direct task assignment
2. NARRATIVE_MULTI: Multiple tasks in narrative form
3. BULLET_LIST: Bullet-pointed task list
4. MIXED: Combination of formats

### Processing Pipeline

1. Message Reception
   - Format detection
   - Single task fast-path
   - Batch task queueing

2. Batch Processing
   - Sequential task processing
   - Progress updates
   - Error handling and retry

3. User Feedback
   - Initial acknowledgment
   - Progress updates
   - Final summary

### Implementation Approach

1. **Fast Path** (Single Tasks):
   - Direct processing without queueing
   - Immediate feedback
   - Existing pipeline used

2. **Batch Processing** (Multiple Tasks):
   - Store in queue
   - Process sequentially
   - Aggregate results
   - Provide summary

This enhancement will maintain the simplicity of single-task processing while adding robust batch capabilities for more complex scenarios.

# AI Agent Integration Plan

## Background and Motivation
We are transitioning from a Telegram-based interface to a React frontend while maintaining the core AI agent functionality. The goal is to minimize code rewriting while providing a seamless user experience.

## Key Challenges and Analysis
1. **Current State**:
   - React frontend with chat interface already implemented
   - Message handling and UI components in place
   - Currently using mock responses
   - Backend logic exists and works, just needs to be "hooked" to frontend

2. **Integration Strategy**:
   - Create Next.js API routes as bridge between frontend and existing backend logic
   - Reuse existing backend processing without major changes
   - Maintain current user experience
   - Keep Telegram integration for testing purposes

## High-level Task Breakdown

### Phase 1: Backend API Setup
1. Create API routes in Next.js
   - [ ] Set up `/api/chat` endpoint
   - [ ] Implement message handling logic
   - [ ] Add error handling and rate limiting

2. Environment Configuration
   - [ ] Set up environment variables for AI agent
   - [ ] Configure API keys and secrets
   - [ ] Document required environment variables

### Phase 2: Frontend Integration
1. API Client Setup
   - [ ] Create API client utilities
   - [ ] Implement error handling
   - [ ] Add request/response types

2. State Management
   - [ ] Update message handling logic
   - [ ] Implement proper loading states
   - [ ] Add error state handling

### Phase 3: Testing & Optimization
1. Testing
   - [ ] Add unit tests for API integration
   - [ ] Implement end-to-end tests
   - [ ] Test error scenarios

2. Performance
   - [ ] Optimize message handling
   - [ ] Implement proper caching
   - [ ] Add performance monitoring

## Required Information from User ✓ COMPLETED
1. Documentation needed:
   - [x] Current AI agent API endpoints and their specifications - No external APIs, reuse existing backend logic
   - [x] Authentication requirements - None required
   - [x] Rate limiting rules - No rate limiting needed
   - [x] Error handling patterns - Standard Next.js error handling

2. Environment variables needed:
   - [x] OPENROUTER_API_KEY - for AI calls
   - [x] GOOGLE_SCRIPT_URL - for spreadsheet integration  
   - [x] SENDGRID_KEY - for email functionality
   - [x] TELEGRAM_TOKEN, ADMIN_TELEGRAM_ID - keep for testing
   - [x] DATABASE_PATH - for local database
   - [x] PORT - for health check
   - [x] TEST_MODE, TEST_EMAIL - for testing

## Project Status Board
- [x] Phase 1: Backend API Setup ✓ COMPLETED
  - [x] Set up `/api/chat` endpoint
  - [x] Implement basic message handling logic
  - [x] Add error handling
- [x] Phase 2: Frontend Integration ✓ COMPLETED  
  - [x] Update frontend to call API instead of mock response
  - [x] Implement proper error handling
  - [x] Maintain existing UI/UX
- [ ] Phase 3: Replace mock logic with actual AI agent processing

## Current Status / Progress Tracking
✅ **PWA IMPLEMENTATION COMPLETED**: 
- Created PWA manifest.json with app metadata
- Updated layout.tsx with comprehensive SEO and PWA metadata
- Configured Next.js for PWA features and security headers
- App is now installable on mobile/desktop devices
- Optimized for `shift6.dwings.app` domain

✅ **Environment Variables**: Configured in `app/my-app/.env.local`
✅ **AI Integration**: Full TypeScript port working locally
✅ **Git Repository**: All changes committed and pushed to `ui-ux` branch
✅ **BRANCH MERGE COMPLETED**: ui-ux branch successfully merged into main
✅ **RENDER DEPLOYMENT COMPLETED**: Live at https://shift6-buildout.onrender.com

## Project Status Board
- [x] **Phase 1**: PWA Implementation (App-like Properties) ✅ COMPLETED
  - [x] Create `manifest.json` with app metadata
  - [x] Update Next.js layout with PWA metadata
  - [x] Configure app name, theme colors, display mode
  - [x] Add security and performance headers
- [x] **Git Integration**: ✅ COMPLETED
  - [x] All changes committed to ui-ux branch
  - [x] ui-ux branch merged into main
  - [x] All changes pushed to GitHub main branch
- [x] **Phase 2**: Render Deployment ✅ COMPLETED
  - [x] Create new Render service for Next.js app
  - [x] Configure build settings and environment variables
  - [x] Deploy and get Render URL: https://shift6-buildout.onrender.com
- [ ] **Phase 3**: Domain Configuration (IN PROGRESS)
  - [ ] Add CNAME record in Squarespace DNS
  - [ ] Configure custom domain on Render
  - [ ] Test https://shift6.dwings.app

## Executor's Feedback or Assistance Requests
🎯 **RENDER DEPLOYMENT SUCCESSFUL**: App is live at https://shift6-buildout.onrender.com

**Next Task**: Configure DNS to point shift6.dwings.app to the Render deployment.

**DNS Configuration for Squarespace**:
1. Login to Squarespace → Settings → Domains → dwings.app → DNS Settings
2. Add this CNAME record:
   ```
   Type: CNAME
   Name: shift6
   Value: shift6-buildout.onrender.com
   TTL: 300
   ```

**After DNS Configuration**:
3. In Render service settings, add custom domain: shift6.dwings.app
4. Render will automatically provision SSL certificate
5. Test final URL: https://shift6.dwings.app

**The app should be fully functional at the Render URL now. Ready to configure the custom domain?**
