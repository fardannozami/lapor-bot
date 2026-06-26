package usecase

import (
	"context"
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
