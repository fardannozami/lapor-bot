package usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type cancelReportRepoStub struct {
	domain.ReportRepository
	report           *domain.Report
	dates            []time.Time
	datesByKind      map[string][]time.Time
	deletedLog       bool
	deletedLogDate   time.Time
	deletedLogKind   string
	latestDeleted    bool
	dailyCount       int
	dailyCountByKind map[string]int
	remainingCount   int
	deletedReport    bool
	upsertedReport   *domain.Report
	questText        string
	questJSON        string
	savedQuestJSON   string
}

func (r *cancelReportRepoStub) GetReport(ctx context.Context, userID string) (*domain.Report, error) {
	return r.report, nil
}

func (r *cancelReportRepoStub) UpsertReport(ctx context.Context, report *domain.Report) error {
	r.upsertedReport = report
	return nil
}

func (r *cancelReportRepoStub) UpsertReportWithActivity(ctx context.Context, report *domain.Report, activityDate time.Time) error {
	return nil
}

func (r *cancelReportRepoStub) GetAllReports(ctx context.Context) ([]*domain.Report, error) {
	return nil, nil
}

func (r *cancelReportRepoStub) LogActivity(ctx context.Context, userID string, activityDate time.Time) error {
	return nil
}

func (r *cancelReportRepoStub) GetActivityCountsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.ActivityLeaderboardEntry, error) {
	return nil, nil
}

func (r *cancelReportRepoStub) GetInactiveUsers(ctx context.Context, days int) ([]*domain.Report, error) {
	return nil, nil
}

func (r *cancelReportRepoStub) ResetAllReports(ctx context.Context) error {
	return nil
}

func (r *cancelReportRepoStub) InitTable(ctx context.Context) error {
	return nil
}

func (r *cancelReportRepoStub) ResolveLIDToPhone(ctx context.Context, lid string) string {
	return lid
}

func (r *cancelReportRepoStub) GetUserActivityDates(ctx context.Context, userID string) ([]time.Time, error) {
	return r.dates, nil
}

func (r *cancelReportRepoStub) GetUserActivityDatesByKind(ctx context.Context, userID string, kind string) ([]time.Time, error) {
	if r.datesByKind != nil {
		return r.datesByKind[kind], nil
	}
	return r.dates, nil
}

func (r *cancelReportRepoStub) DeleteActivityLog(ctx context.Context, userID string, activityDate time.Time) error {
	r.deletedLog = true
	r.deletedLogDate = activityDate
	r.dailyCount = 0
	return nil
}

func (r *cancelReportRepoStub) DeleteActivityLogByKind(ctx context.Context, userID string, activityDate time.Time, kind string) error {
	r.deletedLog = true
	r.deletedLogDate = activityDate
	r.deletedLogKind = kind
	if r.dailyCountByKind != nil {
		r.dailyCountByKind[kind] = 0
	} else {
		r.dailyCount = 0
	}
	return nil
}

func (r *cancelReportRepoStub) DeleteLatestActivityLog(ctx context.Context, userID string, activityDate time.Time) (int, error) {
	r.latestDeleted = true
	r.deletedLogDate = activityDate
	if r.dailyCount > 0 {
		r.dailyCount--
	}
	r.remainingCount = r.dailyCount
	return r.remainingCount, nil
}

func (r *cancelReportRepoStub) DeleteLatestActivityLogByKind(ctx context.Context, userID string, activityDate time.Time, kind string) (int, error) {
	r.latestDeleted = true
	r.deletedLogDate = activityDate
	r.deletedLogKind = kind
	if r.dailyCountByKind != nil {
		if r.dailyCountByKind[kind] > 0 {
			r.dailyCountByKind[kind]--
		}
		r.remainingCount = r.dailyCountByKind[kind]
	} else {
		if r.dailyCount > 0 {
			r.dailyCount--
		}
		r.remainingCount = r.dailyCount
	}
	return r.remainingCount, nil
}

func (r *cancelReportRepoStub) DeleteReport(ctx context.Context, userID string) error {
	r.deletedReport = true
	return nil
}

func (r *cancelReportRepoStub) UpsertStravaAccount(ctx context.Context, account *domain.StravaAccount) error {
	return nil
}

func (r *cancelReportRepoStub) GetStravaAccountByAthleteID(ctx context.Context, athleteID int64) (*domain.StravaAccount, error) {
	return nil, nil
}

func (r *cancelReportRepoStub) GetStravaAccountByUserID(ctx context.Context, userID string) (*domain.StravaAccount, error) {
	return nil, nil
}

func (r *cancelReportRepoStub) GetDailyActivityCount(ctx context.Context, userID string, date time.Time) (int, error) {
	return r.dailyCount, nil
}

func (r *cancelReportRepoStub) GetDailyActivityCountByKind(ctx context.Context, userID string, date time.Time, kind string) (int, error) {
	if r.dailyCountByKind != nil {
		return r.dailyCountByKind[kind], nil
	}
	return r.dailyCount, nil
}

func (r *cancelReportRepoStub) GetLatestActivityText(ctx context.Context, userID string, activityDate time.Time, kind string) (string, error) {
	return r.questText, nil
}

func (r *cancelReportRepoStub) GetDailyQuest(ctx context.Context, userID, questDate string) (string, error) {
	return r.questJSON, nil
}

func (r *cancelReportRepoStub) SaveDailyQuest(ctx context.Context, userID, questDate, tasksJSON string) error {
	r.savedQuestJSON = tasksJSON
	return nil
}

func (r *cancelReportRepoStub) SetGoal(ctx context.Context, goal *domain.WeeklyGoal) error {
	return nil
}

func (r *cancelReportRepoStub) GetActiveGoal(ctx context.Context, userID string, now time.Time) (*domain.WeeklyGoal, error) {
	return nil, nil
}

func (r *cancelReportRepoStub) DeleteActiveGoal(ctx context.Context, userID string, now time.Time) error {
	return nil
}

func (r *cancelReportRepoStub) DeleteExpiredGoals(ctx context.Context, now time.Time) (int64, error) {
	return 0, nil
}

func (r *cancelReportRepoStub) GetGoalActivities(ctx context.Context, userID string, startDate, endDate time.Time) ([]domain.GoalActivity, error) {
	return nil, nil
}

func (r *cancelReportRepoStub) RecordGoalActivity(ctx context.Context, userID string, activityDate time.Time, activityText string) (bool, error) {
	return false, nil
}

func TestCancelReport_NoReport(t *testing.T) {
	repo := &cancelReportRepoStub{report: nil}
	uc := NewCancelReportUsecase(repo)

	msg, err := uc.Execute(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "Halo Budi, kamu belum pernah laporan. Belum ada yang bisa dibatalkan." {
		t.Fatalf("unexpected message: %s", msg)
	}
	if repo.deletedLog {
		t.Fatal("DeleteActivityLog should not be called")
	}
	if repo.deletedReport {
		t.Fatal("DeleteReport should not be called")
	}
}

func TestCancelReport_NotToday(t *testing.T) {
	yesterday := domain.GetToday(time.Now().AddDate(0, 0, -1))
	repo := &cancelReportRepoStub{
		report: &domain.Report{
			UserID:         "user1",
			Name:           "Budi",
			LastReportDate: yesterday,
		},
	}
	uc := NewCancelReportUsecase(repo)

	msg, err := uc.Execute(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "Halo Budi, tidak menemukan laporan untuk hari ini." {
		t.Fatalf("unexpected message: %s", msg)
	}
	if repo.deletedLog {
		t.Fatal("DeleteActivityLog should not be called")
	}
}

func TestCancelReport_NoActivityDates(t *testing.T) {
	repo := &cancelReportRepoStub{
		report: &domain.Report{
			UserID:         "user1",
			Name:           "Budi",
			LastReportDate: time.Now(),
		},
		dates:      []time.Time{},
		dailyCount: 1,
	}
	uc := NewCancelReportUsecase(repo)

	msg, err := uc.Execute(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "Halo Budi, tidak ada aktivitas yang tercatat." {
		t.Fatalf("unexpected message: %s", msg)
	}
	if repo.deletedLog {
		t.Fatal("DeleteActivityLog should not be called")
	}
}

func TestCancelReport_TodayNotInDates(t *testing.T) {
	yesterday := domain.GetToday(time.Now()).AddDate(0, 0, -1)
	repo := &cancelReportRepoStub{
		report: &domain.Report{
			UserID:         "user1",
			Name:           "Budi",
			LastReportDate: time.Now(),
		},
		dates:      []time.Time{yesterday},
		dailyCount: 1,
	}
	uc := NewCancelReportUsecase(repo)

	msg, err := uc.Execute(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "Halo Budi, tidak menemukan laporan untuk hari ini." {
		t.Fatalf("unexpected message: %s", msg)
	}
	if repo.deletedLog {
		t.Fatal("DeleteActivityLog should not be called")
	}
}

func TestCancelReport_SingleDay_CancelsWithoutDeletingAll(t *testing.T) {
	today := domain.GetToday(time.Now())
	repo := &cancelReportRepoStub{
		report: &domain.Report{
			UserID:         "user1",
			Name:           "Budi",
			LastReportDate: time.Now(),
			ActivityCount:  1,
			Streak:         1,
			TotalPoints:    10,
		},
		dates:      []time.Time{today},
		dailyCount: 1,
	}
	uc := NewCancelReportUsecase(repo)

	msg, err := uc.Execute(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.deletedLog {
		t.Fatal("DeleteActivityLog should be called")
	}
	if !repo.deletedLogDate.Equal(today) {
		t.Fatalf("DeleteActivityLog should be called with today, got %v", repo.deletedLogDate)
	}
	if repo.deletedReport {
		t.Fatal("DeleteReport should NOT be called - report hari ini saja yang dihapus, bukan semua report")
	}
	if repo.upsertedReport == nil {
		t.Fatal("UpsertReport should be called")
	}

	// Pastikan report direset ke nilai 0 tapi tetap ada
	if repo.upsertedReport.ActivityCount != 0 {
		t.Fatalf("expected ActivityCount=0, got %d", repo.upsertedReport.ActivityCount)
	}
	if repo.upsertedReport.Streak != 0 {
		t.Fatalf("expected Streak=0, got %d", repo.upsertedReport.Streak)
	}
	if repo.upsertedReport.TotalPoints != 0 {
		t.Fatalf("expected TotalPoints=0, got %d", repo.upsertedReport.TotalPoints)
	}
	if repo.upsertedReport.UserID != "user1" {
		t.Fatalf("expected UserID=user1, got %s", repo.upsertedReport.UserID)
	}
	if repo.upsertedReport.Name != "Budi" {
		t.Fatalf("expected Name=Budi, got %s", repo.upsertedReport.Name)
	}

	// Pesan harus mengandung konfirmasi cancel
	if msg == "" {
		t.Fatal("message should not be empty")
	}
}

func TestCancelReport_MultipleDays_Recalculates(t *testing.T) {
	today := domain.GetToday(time.Now())
	lastWeek := today.AddDate(0, 0, -7)
	repo := &cancelReportRepoStub{
		report: &domain.Report{
			UserID:         "user1",
			Name:           "Budi",
			LastReportDate: time.Now(),
			ActivityCount:  2,
			Streak:         2,
			TotalPoints:    20,
		},
		dates:      []time.Time{lastWeek, today},
		dailyCount: 1,
	}
	uc := NewCancelReportUsecase(repo)

	msg, err := uc.Execute(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.deletedLog {
		t.Fatal("DeleteActivityLog should be called")
	}
	if repo.deletedReport {
		t.Fatal("DeleteReport should NOT be called")
	}
	if repo.upsertedReport == nil {
		t.Fatal("UpsertReport should be called")
	}

	// Setelah cancel hari ini, sisa 1 tanggal minggu lalu
	if repo.upsertedReport.ActivityCount != 1 {
		t.Fatalf("expected ActivityCount=1, got %d", repo.upsertedReport.ActivityCount)
	}
	if repo.upsertedReport.UserID != "user1" {
		t.Fatalf("expected UserID=user1, got %s", repo.upsertedReport.UserID)
	}

	if msg == "" {
		t.Fatal("message should not be empty")
	}
}

func TestCancelReport_DeletesOnlyRegularReportKind(t *testing.T) {
	today := domain.GetToday(time.Now())
	repo := &cancelReportRepoStub{
		report: &domain.Report{
			UserID:             "user1",
			Name:               "Budi",
			JobClass:           "fighter",
			LastReportDate:     time.Now(),
			ActivityCount:      1,
			TotalPoints:        10,
			TotalSideQuests:    2,
			SeasonalSideQuests: 2,
			Str:                7,
			End:                4,
			Luk:                2,
		},
		datesByKind: map[string][]time.Time{
			domain.ActivityKindRegularReport: {today},
			domain.ActivityKindSideQuest:     {today},
		},
		dailyCountByKind: map[string]int{
			domain.ActivityKindRegularReport: 1,
			domain.ActivityKindSideQuest:     2,
		},
	}
	uc := NewCancelReportUsecase(repo)

	_, err := uc.Execute(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.deletedLogKind != domain.ActivityKindRegularReport {
		t.Fatalf("expected regular report delete, got %q", repo.deletedLogKind)
	}
	if repo.dailyCountByKind[domain.ActivityKindSideQuest] != 2 {
		t.Fatalf("sidequest count should not change, got %d", repo.dailyCountByKind[domain.ActivityKindSideQuest])
	}
	if repo.upsertedReport.TotalSideQuests != 2 {
		t.Fatalf("sidequest total should be preserved, got %d", repo.upsertedReport.TotalSideQuests)
	}
	if repo.upsertedReport.JobClass != "fighter" || repo.upsertedReport.Str != 7 {
		t.Fatalf("non-report fields should be preserved, got job=%q str=%d", repo.upsertedReport.JobClass, repo.upsertedReport.Str)
	}
	if repo.upsertedReport.End != 4 || repo.upsertedReport.Luk != 2 {
		t.Fatalf("END/LUK should be preserved, got end=%d luk=%d", repo.upsertedReport.End, repo.upsertedReport.Luk)
	}
}

func TestCancelSideQuest_DeletesOnlyLatestSideQuestKind(t *testing.T) {
	today := domain.GetToday(time.Now())
	repo := &cancelReportRepoStub{
		report: &domain.Report{
			UserID:             "user1",
			Name:               "Budi",
			LastReportDate:     time.Now(),
			ActivityCount:      3,
			TotalSideQuests:    2,
			SeasonalSideQuests: 2,
		},
		dailyCountByKind: map[string]int{
			domain.ActivityKindRegularReport: 1,
			domain.ActivityKindSideQuest:     2,
		},
	}
	uc := NewCancelReportUsecase(repo)

	msg, err := uc.ExecuteSideQuest(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.latestDeleted {
		t.Fatal("DeleteLatestActivityLogByKind should be called")
	}
	if !repo.deletedLogDate.Equal(today) {
		t.Fatalf("expected delete date %v, got %v", today, repo.deletedLogDate)
	}
	if repo.deletedLogKind != domain.ActivityKindSideQuest {
		t.Fatalf("expected sidequest delete, got %q", repo.deletedLogKind)
	}
	if repo.dailyCountByKind[domain.ActivityKindRegularReport] != 1 {
		t.Fatalf("regular report count should not change, got %d", repo.dailyCountByKind[domain.ActivityKindRegularReport])
	}
	if repo.upsertedReport.TotalSideQuests != 1 {
		t.Fatalf("expected TotalSideQuests=1, got %d", repo.upsertedReport.TotalSideQuests)
	}
	if repo.upsertedReport.ActivityCount != 3 {
		t.Fatalf("regular ActivityCount should not change, got %d", repo.upsertedReport.ActivityCount)
	}
	if msg == "" {
		t.Fatal("message should not be empty")
	}
}

// cancelLedgerRepoStub adds the ledger capabilities (event query + atomic
// cancel persistence + quest storage + exact walk subtraction) on top of the
// base stub, so tests can exercise the exact-reversal paths.
type cancelLedgerRepoStub struct {
	*cancelReportRepoStub
	events          []domain.ReportActivityEvent
	cancelEvent     *domain.ReportActivityEvent
	cancelDeleteAll bool
	cancelReport    *domain.Report
	cancelCalled    bool
}

func (s *cancelLedgerRepoStub) GetReportEvents(ctx context.Context, userID string, activityDate time.Time, kind string) ([]domain.ReportActivityEvent, error) {
	return s.events, nil
}

func (s *cancelLedgerRepoStub) CancelActivityEvents(ctx context.Context, report *domain.Report, cancelEvent domain.ReportActivityEvent, kind string, deleteAll bool) (int, error) {
	s.cancelCalled = true
	s.cancelDeleteAll = deleteAll
	s.cancelReport = report
	ev := cancelEvent
	s.cancelEvent = &ev
	// Mirror the repo contract: decrement or clear the kind counter.
	if s.dailyCountByKind == nil {
		if deleteAll {
			s.dailyCount = 0
			return 0, nil
		}
		if s.dailyCount > 0 {
			s.dailyCount--
		}
		return s.dailyCount, nil
	}
	if deleteAll {
		s.dailyCountByKind[kind] = 0
		return 0, nil
	}
	if s.dailyCountByKind[kind] > 0 {
		s.dailyCountByKind[kind]--
	}
	return s.dailyCountByKind[kind], nil
}

// TestCancelReport_MultiDay_PreservesSeasonProgress is the F1 regression test:
// cancelling today's report must remove ONLY today's value. Points and
// seasonal progress from earlier days survive; today's exact grant (from the
// ledger, including badge points bundled into the event) is reversed.
func TestCancelReport_MultiDay_PreservesSeasonProgress(t *testing.T) {
	today := domain.GetToday(time.Now())
	lastWeek := today.AddDate(0, 0, -7)
	base := &cancelReportRepoStub{
		report: &domain.Report{
			UserID: "user1", Name: "Budi",
			LastReportDate:        time.Now(),
			ActivityCount:         2,
			Streak:                2,
			MaxStreak:             2,
			SeasonalMaxStreak:     2,
			TotalPoints:           100,
			SeasonalPoints:        80,
			SeasonalActivityCount: 5,
			Achievements:          "first_report,streak_1",
			SeasonalAchievements:  "first_report,streak_1,streak_2",
			WalkStepsTotal:        9000,
			SeasonalWalkSteps:     9000,
		},
		dates:      []time.Time{lastWeek, today},
		dailyCount: 1,
	}
	repo := &cancelLedgerRepoStub{
		cancelReportRepoStub: base,
		events: []domain.ReportActivityEvent{{
			EventID: "ev-today", UserID: "user1", SeasonNumber: 1,
			Kind: domain.ActivityKindRegularReport, ActivityDate: today,
			OccurredAt: time.Now(), PointsDelta: 17, RegularCountDelta: 1,
			Source: "whatsapp", ActivityText: "gym",
			MetadataJSON: `{"walk_steps": 4000}`,
		}},
	}
	uc := NewCancelReportUsecase(repo)

	msg, err := uc.Execute(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.cancelCalled {
		t.Fatal("expected atomic CancelActivityEvents persistence")
	}
	if !repo.cancelDeleteAll {
		t.Fatal("single-report day must cancel with deleteAll=true")
	}
	ce := repo.cancelEvent
	if ce == nil || ce.Source != "cancel" || ce.PointsDelta != -17 || ce.RegularCountDelta != -1 {
		t.Fatalf("compensating event wrong: %+v", ce)
	}
	if !strings.Contains(ce.MetadataJSON, `"walk_steps":-4000`) {
		t.Fatalf("compensating event must carry negative walk metadata, got %s", ce.MetadataJSON)
	}

	r := repo.cancelReport
	if r == nil {
		t.Fatal("cancel must persist the recalculated report")
	}
	if r.ActivityCount != 1 || r.Streak != 1 {
		t.Fatalf("counters should be recalculated to last week only, got count=%d streak=%d", r.ActivityCount, r.Streak)
	}
	if r.TotalPoints != 83 || r.SeasonalPoints != 63 {
		t.Fatalf("points must be preserved minus today's 17, got total=%d season=%d", r.TotalPoints, r.SeasonalPoints)
	}
	if r.SeasonalActivityCount != 4 {
		t.Fatalf("one season active day removed: got %d, want 4", r.SeasonalActivityCount)
	}
	if r.SeasonalAchievements != "first_report,streak_1,streak_2" || r.Achievements != "first_report,streak_1" {
		t.Fatalf("earned badges must be preserved, got season=%q all=%q", r.SeasonalAchievements, r.Achievements)
	}
	if r.WalkStepsTotal != 5000 || r.SeasonalWalkSteps != 5000 {
		t.Fatalf("walk totals must keep earlier days' steps: got %d/%d, want 5000/5000", r.WalkStepsTotal, r.SeasonalWalkSteps)
	}
	if !strings.Contains(msg, "-17") {
		t.Fatalf("response should announce the deducted points, got: %s", msg)
	}
}

// TestCancelSideQuest_ReversesPointsAndQuestState is the F2 regression test:
// cancelling a side quest must reverse its difficulty points, decrement the
// counters, and reopen the task in today's quest list.
func TestCancelSideQuest_ReversesPointsAndQuestState(t *testing.T) {
	today := domain.GetToday(time.Now())
	questJSON := `[{"id":"q1","name":"Chair Squat","target":18,"progress":18,"unit":"x","difficulty":"medium"}]`
	base := &cancelReportRepoStub{
		report: &domain.Report{
			UserID: "user1", Name: "Budi",
			LastReportDate:     time.Now(),
			ActivityCount:      3,
			TotalPoints:        50,
			SeasonalPoints:     40,
			TotalSideQuests:    2,
			SeasonalSideQuests: 2,
		},
		dailyCountByKind: map[string]int{
			domain.ActivityKindRegularReport: 1,
			domain.ActivityKindSideQuest:     1,
		},
		questJSON: questJSON,
	}
	repo := &cancelLedgerRepoStub{
		cancelReportRepoStub: base,
		events: []domain.ReportActivityEvent{{
			EventID: "ev-sq", UserID: "user1", SeasonNumber: 1,
			Kind: domain.ActivityKindSideQuest, ActivityDate: today,
			OccurredAt: time.Now(), PointsDelta: 3, SideQuestCountDelta: 1,
			Source: "whatsapp", ActivityText: "Side quest: Chair Squat",
			MetadataJSON: "{}",
		}},
	}
	uc := NewCancelReportUsecase(repo)

	msg, err := uc.ExecuteSideQuest(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.cancelCalled || repo.cancelDeleteAll != true {
		t.Fatalf("expected atomic full sidequest cancel, got called=%v deleteAll=%v", repo.cancelCalled, repo.cancelDeleteAll)
	}
	if ce := repo.cancelEvent; ce == nil || ce.PointsDelta != -3 || ce.SideQuestCountDelta != -1 {
		t.Fatalf("compensating event wrong: %+v", ce)
	}
	r := repo.cancelReport
	if r.TotalPoints != 47 || r.SeasonalPoints != 37 {
		t.Fatalf("sidequest points must be reversed: got total=%d season=%d", r.TotalPoints, r.SeasonalPoints)
	}
	if r.TotalSideQuests != 1 || r.SeasonalSideQuests != 1 {
		t.Fatalf("sidequest counters must decrement: got %d/%d", r.TotalSideQuests, r.SeasonalSideQuests)
	}
	if r.ActivityCount != 3 {
		t.Fatalf("regular ActivityCount must not change, got %d", r.ActivityCount)
	}
	// Quest state reverted: task progress back to 0 so it can be redone.
	if !strings.Contains(repo.savedQuestJSON, `"progress":0`) {
		t.Fatalf("quest completion should be reverted, saved: %s", repo.savedQuestJSON)
	}
	if !strings.Contains(msg, "Chair Squat") {
		t.Fatalf("response should name the reopened quest, got: %s", msg)
	}
}

// TestCancelSideQuest_PreLedger_UsesQuestDifficultyFallback covers the
// no-ledger path: points are recovered from the reverted quest's difficulty
// instead of the ledger.
func TestCancelSideQuest_PreLedger_UsesQuestDifficultyFallback(t *testing.T) {
	questJSON := `[{"id":"q1","name":"Push-up","target":20,"progress":20,"unit":"x","difficulty":"hard"}]`
	repo := &cancelReportRepoStub{
		report: &domain.Report{
			UserID: "user1", Name: "Budi",
			LastReportDate:     time.Now(),
			ActivityCount:      3,
			TotalPoints:        50,
			SeasonalPoints:     40,
			TotalSideQuests:    1,
			SeasonalSideQuests: 1,
		},
		dailyCountByKind: map[string]int{
			domain.ActivityKindRegularReport: 1,
			domain.ActivityKindSideQuest:     1,
		},
		questText: "Side quest: Push-up",
		questJSON: questJSON,
	}
	uc := NewCancelReportUsecase(repo)

	msg, err := uc.ExecuteSideQuest(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := repo.upsertedReport
	if r == nil {
		t.Fatal("legacy path must upsert the report")
	}
	if r.TotalPoints != 45 || r.SeasonalPoints != 35 {
		t.Fatalf("hard quest = 5 pts fallback reversal, got total=%d season=%d", r.TotalPoints, r.SeasonalPoints)
	}
	if r.TotalSideQuests != 0 {
		t.Fatalf("expected TotalSideQuests=0, got %d", r.TotalSideQuests)
	}
	if !strings.Contains(repo.savedQuestJSON, `"progress":0`) {
		t.Fatalf("quest completion should be reverted, saved: %s", repo.savedQuestJSON)
	}
	if !strings.Contains(msg, "-5") {
		t.Fatalf("response should announce the deducted points, got: %s", msg)
	}
}

func TestCancelAllSideQuest_DeletesOnlySideQuestKind(t *testing.T) {
	repo := &cancelReportRepoStub{
		report: &domain.Report{
			UserID:             "user1",
			Name:               "Budi",
			LastReportDate:     time.Now(),
			ActivityCount:      3,
			TotalSideQuests:    2,
			SeasonalSideQuests: 2,
		},
		dailyCountByKind: map[string]int{
			domain.ActivityKindRegularReport: 1,
			domain.ActivityKindSideQuest:     2,
		},
	}
	uc := NewCancelReportUsecase(repo)

	_, err := uc.ExecuteAllSideQuest(context.Background(), "user1", "Budi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.deletedLogKind != domain.ActivityKindSideQuest {
		t.Fatalf("expected sidequest delete, got %q", repo.deletedLogKind)
	}
	if repo.dailyCountByKind[domain.ActivityKindRegularReport] != 1 {
		t.Fatalf("regular report count should not change, got %d", repo.dailyCountByKind[domain.ActivityKindRegularReport])
	}
	if repo.upsertedReport.TotalSideQuests != 0 {
		t.Fatalf("expected TotalSideQuests=0, got %d", repo.upsertedReport.TotalSideQuests)
	}
}
