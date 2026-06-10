package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type cancelReportRepoStub struct {
	report         *domain.Report
	dates          []time.Time
	deletedLog     bool
	deletedLogDate time.Time
	latestDeleted  bool
	dailyCount     int
	remainingCount int
	deletedReport  bool
	upsertedReport *domain.Report
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

func (r *cancelReportRepoStub) DeleteActivityLog(ctx context.Context, userID string, activityDate time.Time) error {
	r.deletedLog = true
	r.deletedLogDate = activityDate
	r.dailyCount = 0
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
	if msg != "Halo Budi, kamu belum laporan hari ini. Tidak ada yang perlu dibatalkan." {
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
		dates: []time.Time{},
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
		dates: []time.Time{yesterday},
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
