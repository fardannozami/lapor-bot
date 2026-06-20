# Plan: Personal Dashboard — Update/Add Data (Username, Goal with Start Day+Hour, Job)

## Context

- "Lapor Bot" is a WhatsApp activity tracker with a web dashboard (frontend/apps/web) and Go backend.
- After login (via phone), users land in a **Personal Dashboard** (`PersonalPage.tsx`).
- Currently `PersonalPage` is **read-only** (displays stats, streak map, active goal, side quests, job, etc.).
- Onboarding flow (`ProfileSetup.tsx`) already implements name → job → goal using existing APIs, but it is only shown for first-time users (phone-like names).
- User wants **in-dashboard** update/add capabilities:
  1. Change username (update name).
  2. Set goal, with ability to **choose start day + start hour**; the goal runs for 7 days / 1 week from that start.
  3. Choose job from existing list.
- Backend already exposes commands/APIs that can be reused:
  - `POST /api/user/name` → `updateName(phone, name)`
  - `POST /api/user/job` → `selectJob(phone, jobId)`
  - `POST /api/user/goal` → `setGoal(phone, targetDays, activity)` (simulates `#goal set`)
  - `GET /api/jobs` → `listJobs()`
- Current goal implementation always uses "now" as `StartAt` (see `goal_usecase.go:77`, `HandleSetGoal`). Custom start day/hour will require **backend extension**.

## Approach (Recommended)

- Add **inline edit / action sections or lightweight modals** inside `PersonalPage` (after login) for the three operations.
- Reuse existing repository interfaces and HTTP client calls (do not duplicate logic).
- For username and job: call existing endpoints directly (no backend change needed).
- For goal with start day+hour:
  - Extend the goal API surface (backend handler + usecase/repo path + contract) to accept optional start info.
  - Keep backward compatibility for existing callers (ProfileSetup, WA commands).
  - Frontend will allow picking day (today or future within a small window) + hour (e.g., 00:00–23:00), compute a start timestamp, and send it.
- Prefer consistency with existing patterns:
  - Use `useRepositories()` + `IReportRepository` methods.
  - After successful mutation, refetch the personal user (via `fetchUserByPhone`) and update local state so UI reflects changes (like ProfileSetup does).
- Keep the visual style (glass, orbitron, system colors, font-mono labels).
- Do not break the public leaderboard or existing WA flows.

## Files to Modify

**Frontend (web app + shared packages):**

- `frontend/apps/web/src/components/PersonalPage.tsx` — add UI sections/buttons for edit name, select job, set/update goal (with start controls). Show current values and allow change.
- `frontend/apps/web/src/components/PersonalPage.tsx` (or new small components) — possibly extract small forms or use inline controlled inputs + save buttons.
- `frontend/apps/web/src/App.tsx` — ensure refreshed user after updates can be passed back (already does some of this via `setPersonalUser`).
- `frontend/packages/shared/src/domain/repositories.ts` — extend `setGoal` signature if we add start params (or add a new method `setGoalWithStart`).
- `frontend/packages/contract/src/repositories/HttpReportRepository.ts` — implement the (extended) call to `/api/user/goal`.
- `frontend/packages/shared/src/types.ts` — if needed, add optional start fields or a richer goal request type.
- (Optional) New small component under `components/` like `GoalSetter.tsx` or `InlineEdit.tsx` if we want to keep PersonalPage clean.

**Backend:**

- `internal/infra/http/handler.go` — extend `HandleSetGoal` to accept optional start fields (e.g., `start_date`, `start_hour` or a full `start_at` ISO). Update body struct and call site.
- `internal/app/usecase/goal_usecase.go` — add or expose a path that accepts a custom start time instead of always `now`. Keep the WA command path using current time.
- `internal/domain/report.go` — (no change expected to structs).
- `internal/infra/sqlite/report_repository.go` — no change (already stores `start_at`/`end_at`).
- Ensure `HandleGetUser` / `HandleGetUserByPhone` continue to return the enriched goal with computed `start_at`/`end_at` (already does via `buildPersonalGoal`).

**Contract / Shared (to keep types aligned):**

- Possibly add a small request type or extend the existing `setGoal` to accept optional start info without breaking current callers.

## Reuse (Existing Code to Leverage)

- Repository interface + implementations:
  - `IReportRepository.updateName`, `selectJob`, `setGoal`, `listJobs`, `fetchUserByPhone` (frontend/packages/shared/src/domain/repositories.ts)
  - `HttpReportRepository` methods (frontend/packages/contract/src/repositories/HttpReportRepository.ts)
- Existing handlers:
  - `HandleUpdateName`, `HandleSelectJob`, `HandleSetGoal`, `HandleListJobs` (internal/infra/http/handler.go)
- Existing usecases:
  - `UpdateNameUsecase.Execute`, `JobUsecase.Select`, `GoalUsecase.Execute` + `set` (internal/app/usecase/*)
- Job list source: `domain.AllJobClasses` exposed via `GET /api/jobs`
- Personal data enrichment + goal building: `enrichReportWithMasking`, `buildPersonalGoal`, `HandleGetUser`/`HandleGetUserByPhone`
- UI patterns: `ProfileSetup.tsx` (stepper, loading, error display, job cards, select, inputs) as visual/UX reference.
- Auth flow: `useAuth`, `LoginPage` → `App` state for `personalUser`.
- After mutation pattern: `repo.fetchUserByPhone(phone).then(setPersonalUser)` (ProfileSetup already does this).

## Steps (Implementation Checklist)

- [ ] **Backend goal start support**
  - [ ] Extend request body in `HandleSetGoal` to accept optional start fields (e.g., `start_date?: string` (YYYY-MM-DD), `start_hour?: number` (0-23), or `start_at?: string` ISO). Keep `target_days` + `activity`.
  - [ ] In handler, if start provided, parse and pass a concrete start time to a new or extended goal usecase method; otherwise fall back to "now".
  - [ ] Update `GoalUsecase` to support a `SetWithStart(ctx, userID, targetDays, activity, start time.Time)` path (or overload). Compute `EndAt = start + 7 days`. Keep existing `set` for WA commands.
  - [ ] Return success message (can reuse or adapt existing formatting).
  - [ ] Ensure validation: start not too far in past/future, targetDays 1-7.
- [ ] **Update shared contract**
  - [ ] Update `IReportRepository.setGoal` signature (or add `setGoalWithStart(phone, targetDays, activity, start?: Date)`) in `frontend/packages/shared/src/domain/repositories.ts`.
  - [ ] Implement in `HttpReportRepository` (send the extra fields if present).
  - [ ] Update types if a dedicated request shape is introduced.
- [ ] **Frontend UI in Personal Dashboard**
  - [ ] In `PersonalPage`, add an "Edit Profile" or per-section actions:
    - Username: show current `name`, an "Ubah Nama" button that reveals an input + save (or a small modal). Call `repo.updateName(phone, newName)`, then `fetchUserByPhone` and update parent state.
    - Job: show current job; a "Pilih / Ganti Job" section or button that lists jobs (call `repo.listJobs()`), lets user pick one, calls `repo.selectJob(phone, jobId)`. Gate or show the same 50-point rule message if backend returns it.
    - Goal: show current active goal (already rendered). Add "Atur / Ubah Goal" that allows:
      - `targetDays` (1-7) select
      - `activity` text
      - Start day picker (e.g., date input or day-of-week relative to today)
      - Start hour select (00:00 … 23:00)
      - On save, compute a start timestamp (Asia/Jakarta aware if possible) and call the (extended) `setGoal`.
      - If an active goal exists, either:
        - Call reset first (if we expose a reset), or
        - Let backend error and surface "reset dulu" message (match current WA behavior), or
        - Offer a "Ganti Goal" that resets then sets.
  - [ ] Handle loading, success toast or inline confirmation, and error states consistently with ProfileSetup/Login.
  - [ ] After any successful update, refresh the `personalUser` so the whole dashboard reflects new data (streak map, goal card, job badge, header name).
  - [ ] Ensure "Kembali" and "Keluar" remain available; avoid accidental navigation away during edits.
- [ ] **Polish & consistency**
  - Match existing design tokens (glass, borders, font-orbitron titles, font-mono labels, system-* colors).
  - Make sure mobile responsiveness is preserved.
  - If an active goal exists, surface "Sisa X hari" and perhaps a "Reset Goal" action that calls a reset path (add a thin `resetGoal(phone)` if not present, or use WA-style messaging).
- [ ] **Backward compatibility**
  - Existing `ProfileSetup` calls to `setGoal(phone, targetDays, activity)` should continue to work (start = now).
  - WA commands (`#goal set`) continue to use server "now".
- [ ] **Testing / verification paths**
  - Manual: login → personal → change name → verify header and public profile (if unmasked in other views) updates.
  - Manual: pick a different job → verify job badge + description updates; attempt job change below threshold and see error.
  - Manual: set goal with specific start day/hour → verify in the goal card the `start_at`/`end_at` reflect 7-day window; progress updates when reporting.
  - Ensure public leaderboard and other users are unaffected.

## Verification

- Run web dev server and backend; login with a phone that has reports.
- Exercise the three flows end-to-end and confirm:
  1. Name change persists and is visible immediately in PersonalPage header.
  2. Job selection updates the job card and (after refresh) side quests if applicable.
  3. Goal creation with chosen start day + hour results in a 7-day window shown in the Weekly Goal section; `start_at` and `end_at` in the enriched response match.
- Check that existing flows still work:
  - Public dashboard loads.
  - ProfileSetup (first login with phone-like name) still completes.
  - WhatsApp `#goal set`, `#job`, `#setname` continue to function.
- No regressions in `HandleGetUser`/`HandleGetUserByPhone` responses (they must still enrich the goal correctly).
- TypeScript build + lint clean; Go build clean.

## Open Questions / Notes for User

- Should the start day/hour be required, or optional (default "now" like today)?
- What is the allowed range for start (e.g., today ± N days)?
- Do we want a dedicated "Reset Goal" button in the UI, or rely on the set flow to fail with a helpful message?
- Should job change be allowed without the 50-point gate in the personal dashboard, or keep the same rule as WA?
- Any preference on UI pattern: fully inline sections vs. slide-in panels vs. modals?
- Is there a mobile (React Native) personal screen that should receive the same treatment, or is this web-only for now?

## Notes

- Current goal duration is hard-coded to 7 days in multiple places (`goalWindowDays`, `AddDate(0,0,7)`). The plan preserves that.
- The personal user object returned by `fetchUserByPhone` already includes `active_goal` with `start_at`/`end_at`, so once backend accepts custom start, the UI can display it without additional backend work.
