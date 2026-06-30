package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

const (
	MaxDailyRegularReports = 3
	MaxDailySideQuests     = 3
	MaxDailyReports        = MaxDailyRegularReports
	ReportEventLedgerYear  = 2026
	ReportEventLedgerMonth = time.June
	ReportEventLedgerDay   = 24
	AttributeGainPerReport = 1
	jobSetupURL            = "https://lapor-bot.web.id/"

	// Scoring model. Base points are fixed per trigger so every user earns the
	// same amount for the same command (lapor / lapor sidequest). Streak
	// bonuses are additive and capped so a long streak cannot dominate the
	// base reward (best practice: streak share stays ≤ ~30% of a report).
	baseReportPoints         = 10
	sideQuestFallbackPoints  = 5
	seasonalFirstReportBonus = 5

	// Weekly streak bonus is worth more per step than the daily one, but both
	// apply additively when a user keeps both streaks alive — so a user with
	// an active daily AND weekly streak always earns more than with either
	// alone, while the weekly component stays the larger of the two.
	weeklyStreakBonusPerStep = 2
	weeklyStreakBonusCap     = 10 // max +20 pts per report from weekly streak
	dailyStreakBonusPerStep  = 1
	dailyStreakBonusCap      = 5  // max +5 pts per report from daily streak
)

type typedActivityRepository interface {
	UpsertReportWithActivityKind(ctx context.Context, report *domain.Report, activityDate time.Time, kind string) error
	GetDailyActivityCountByKind(ctx context.Context, userID string, date time.Time, kind string) (int, error)
}

type eventActivityRepository interface {
	UpsertReportWithActivityEvent(ctx context.Context, report *domain.Report, event domain.ReportActivityEvent) error
}

// GoalCompletionNotifier is called when a user's weekly goal is completed.
// userID and name identify the user; activity and targetDays describe the completed goal;
// totalCompleted is the cumulative goals_completed count.
type GoalCompletionNotifier func(ctx context.Context, userID string, name string, activity string, targetDays int, totalCompleted int)

type ReportActivityUsecase struct {
	repo         domain.ReportRepository
	goalNotifier GoalCompletionNotifier
	locks        sync.Map
}

type reportActivityOptions struct {
	sideQuestCount  int
	activityText    string
	now             time.Time
	sideQuestPoints int // total points from side quest difficulty multipliers (computed in DailyQuestUsecase)
}

func NewReportActivityUsecase(repo domain.ReportRepository) *ReportActivityUsecase {
	return &ReportActivityUsecase{repo: repo}
}

// SetGoalNotifier sets a callback that fires when a user completes their weekly goal.
func (uc *ReportActivityUsecase) SetGoalNotifier(fn GoalCompletionNotifier) {
	uc.goalNotifier = fn
}

func (uc *ReportActivityUsecase) userLock(userID string) *sync.Mutex {
	lock, _ := uc.locks.LoadOrStore(userID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (uc *ReportActivityUsecase) Execute(ctx context.Context, userID, name string, workout *domain.HevyWorkout) (string, error) {
	return uc.execute(ctx, userID, name, workout, reportActivityOptions{})
}

func (uc *ReportActivityUsecase) ExecuteWithMessage(ctx context.Context, userID, name, message string, workout *domain.HevyWorkout) (string, error) {
	return uc.execute(ctx, userID, name, workout, reportActivityOptions{
		activityText: message,
	})
}

func (uc *ReportActivityUsecase) ExecuteSideQuest(ctx context.Context, userID, name, activityText string, completedCount, sideQuestPoints int, now time.Time) (string, error) {
	if completedCount < 1 {
		completedCount = 1
	}
	return uc.execute(ctx, userID, name, nil, reportActivityOptions{
		sideQuestCount:  completedCount,
		activityText:    activityText,
		now:             now,
		sideQuestPoints: sideQuestPoints,
	})
}

func (uc *ReportActivityUsecase) execute(ctx context.Context, userID, name string, workout *domain.HevyWorkout, opts reportActivityOptions) (string, error) {
	lock := uc.userLock(userID)
	lock.Lock()
	defer lock.Unlock()

	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	now := opts.now
	if now.IsZero() {
		now = time.Now()
	}
	today := domain.GetToday(now)
	isSideQuest := opts.sideQuestCount > 0
	activityKind := domain.ActivityKindRegularReport
	if isSideQuest {
		activityKind = domain.ActivityKindSideQuest
	}

	var dailyCount int
	dailyCount, err = uc.getDailyActivityCount(ctx, userID, today, activityKind)
	if err != nil {
		return "", err
	}

	dailyLimit := MaxDailyRegularReports
	if isSideQuest {
		dailyLimit = MaxDailySideQuests
	}
	if dailyCount >= dailyLimit {
		label := "laporan utama"
		if isSideQuest {
			label = "side quest"
		}
		displayName := name
		if report != nil {
			displayName = report.Name
		}
		msg := fmt.Sprintf("%s sudah mencapai batas %s %dx hari ini. Batas laporan utama dan side quest terpisah: masing-masing %d kali per hari.", displayName, label, dailyLimit, MaxDailyRegularReports)
		if !isSideQuest {
			msg += " Kalau tadi salah input, pakai /cancel untuk hapus laporan terakhir atau /cancel-all untuk hapus semua laporan hari ini."
		}
		msg += " 🙏"
		msg += fmt.Sprintf("\n\n💬 _\"%s\"_", RandomQuote())
		return msg, nil
	}

	isRepeatReport := !isSideQuest && dailyCount > 0
	isFullReport := !isSideQuest && !isRepeatReport

	streakFreezeUsed := false

	if report != nil {
		storedName := report.Name
		name = reportName(storedName, name)
		shouldUpdateStoredName := isUnknownName(storedName) && strings.TrimSpace(storedName) != name
		if shouldUpdateStoredName {
			report.Name = name
		}

		lastReport := report.LastReportDate
		lastReportDate := domain.GetToday(lastReport)

		if isSideQuest {
			if shouldUpdateStoredName {
				if err := uc.repo.UpsertReport(ctx, report); err != nil {
					return "", err
				}
			}
			name = report.Name
		} else if lastReportDate.Equal(today) {
			if shouldUpdateStoredName {
				if err := uc.repo.UpsertReport(ctx, report); err != nil {
					return "", err
				}
			}
			report.LastReportDate = now
			name = report.Name
		} else {
			daysSinceLastReport := int(math.Round(today.Sub(lastReportDate).Hours() / 24))

			currentWeekStart := domain.GetStartOfISOWeek(today)
			lastWeekStart := domain.GetStartOfISOWeek(lastReportDate)
			weeksSinceLastReport := int(math.Round(currentWeekStart.Sub(lastWeekStart).Hours() / (24 * 7)))

			if isFullReport {
				if currentWeekStart.Equal(lastWeekStart) {
				} else if weeksSinceLastReport == 1 {
					report.Streak++
					report.ComebackStreak++
				} else if weeksSinceLastReport == 2 && report.StreakFreezes > 0 {
					report.StreakFreezes--
					report.Streak++
					report.ComebackStreak++
					streakFreezeUsed = true
				} else {
					report.InactiveDays = daysSinceLastReport
					report.ComebackStreak = 1
					report.Streak = 1
				}
				report.ActivityCount++
				report.SeasonalActivityCount++
				if report.Streak > report.SeasonalMaxStreak {
					report.SeasonalMaxStreak = report.Streak
				}

				if report.ActivityCount > 100 {
					report.ActivityCount = 1
					report.CenturionCycles++
				}
			}

			report.LastReportDate = now
			name = report.Name
		}
	} else {
		name = reportName("", name)
		report = &domain.Report{
			UserID:                userID,
			Name:                  name,
			Streak:                1,
			ActivityCount:         1,
			SeasonalActivityCount: 1,
			SeasonalMaxStreak:     1,
			LastReportDate:        now,
			MaxStreak:             0,
			TotalPoints:           0,
			SeasonalPoints:        0,
			Achievements:          "",
			SeasonalAchievements:  "",
			ComebackStreak:        1,
			InactiveDays:          0,
			StreakFreezes:         1,
		}
	}

	if report.MaxStreak == 0 {
		report.MaxStreak = 1
	}
	oldNumericLevel := domain.NumericLevelFromTotalPoints(report.TotalPoints)
	oldLifetimeTier := domain.GetLevel(report.TotalPoints)
	oldSeasonRank := domain.GetSeasonRank(report.SeasonalPoints)

	newRecord := false
	if isFullReport && report.Streak > report.MaxStreak {
		report.MaxStreak = report.Streak
		if report.Streak > 1 {
			newRecord = true
		}
	}

	basePoints := baseReportPoints
	weeklyStreakBonus := 0
	dailyStreakBonus := 0
	seasonalFirstBonus := 0
	if isFullReport {
		weeklyStreakBonus = cappedStreakBonus(int64(report.Streak)-1, weeklyStreakBonusCap, weeklyStreakBonusPerStep)
		dailyStreak := uc.computeDailyStreak(ctx, userID, today)
		dailyStreakBonus = cappedStreakBonus(int64(dailyStreak)-1, dailyStreakBonusCap, dailyStreakBonusPerStep)
		if report.SeasonalActivityCount == 1 {
			seasonalFirstBonus = seasonalFirstReportBonus
		}
	}
	if isSideQuest {
		if opts.sideQuestPoints > 0 {
			basePoints = opts.sideQuestPoints
		} else {
			basePoints = sideQuestFallbackPoints // fallback for callers that don't pass difficulty points
		}
		weeklyStreakBonus = 0
		dailyStreakBonus = 0
		seasonalFirstBonus = 0
		report.TotalSideQuests += opts.sideQuestCount
		report.SeasonalSideQuests += opts.sideQuestCount
	}
	// streakBonus is the additive sum of weekly + daily components. Weekly
	// per-step (2) > daily per-step (1), and both caps apply, so the weekly
	// bonus is always the larger contributor and the two stack when both
	// streaks are active.
	streakBonus := weeklyStreakBonus + dailyStreakBonus
	reportPoints := basePoints + streakBonus + seasonalFirstBonus
	if isRepeatReport && !isSideQuest {
		reportPoints = reportPoints / 2
	}
	report.TotalPoints += reportPoints
	report.SeasonalPoints += reportPoints

	activityForParse := opts.activityText
	if workout != nil {
		activityForParse += " " + workout.Title
		for _, ex := range workout.Exercises {
			activityForParse += " " + ex
		}
	}
	attributesActive := hasSelectedJob(report)
	var statGains []string
	if attributesActive {
		// A report grants a single, fair attribute point. The activity directs
		// which attribute is rewarded; the job breaks ties among multiple
		// matches and is the fallback when nothing matches. Granting +1 to
		// every matched attribute would make mixed sessions worth several
		// times the attribute points of focused ones.
		attrs, _ := domain.ResolveReportAttributes(activityForParse, report.JobClass)
		chosen := domain.SelectReportAttribute(attrs, report.JobClass,
			attributeSelectionSeed(userID, activityKind, today.Format(time.DateOnly), dailyCount+1, activityForParse))
		statGains = applyAttributeGains(report, []domain.AttributeType{chosen}, AttributeGainPerReport)
	}

	var newAchievements []domain.Achievement
	var comebackAchievements []domain.ComebackAchievement
	pointsGained := 0

	if isFullReport && !isSideQuest {
		newAchievements = domain.CheckNewSeasonAchievements(report)
		for _, ach := range newAchievements {
			report.SeasonalAchievements = domain.AddAchievement(report.SeasonalAchievements, ach.ID)
			if !domain.HasAchievement(report.Achievements, ach.ID) {
				report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
			}
			report.TotalPoints += ach.Points
			report.SeasonalPoints += ach.Points
			pointsGained += ach.Points
		}

		comebackAchievements = domain.CheckComebackAchievements(report)
		for _, ach := range comebackAchievements {
			report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
			report.TotalPoints += ach.Points
			pointsGained += ach.Points
		}
	}

	freezeAwarded := false
	if isFullReport && !isSideQuest {
		for _, ach := range newAchievements {
			if ach.ID == "streak_4" && report.StreakFreezes < 2 {
				report.StreakFreezes++
				freezeAwarded = true
			}
		}
	}

	report.Level = domain.NumericLevelFromTotalPoints(report.TotalPoints)
	leveledUp := report.Level > oldNumericLevel
	newLifetimeTier := domain.GetLevel(report.TotalPoints)
	newSeasonRank := domain.GetSeasonRank(report.SeasonalPoints)
	lifetimeTierUp := newLifetimeTier.Tier > oldLifetimeTier.Tier
	seasonRankUp := newSeasonRank.Tier > oldSeasonRank.Tier
	totalPointsGained := reportPoints + pointsGained

	if err := uc.upsertReportWithActivity(ctx, report, reportActivityEventInput{
		activityDate:        today,
		kind:                activityKind,
		occurredAt:          now,
		pointsDelta:         totalPointsGained,
		regularCountDelta:   boolToInt(!isSideQuest),
		sideQuestCountDelta: opts.sideQuestCount,
		activityText:        goalActivityTextWithFallback(workout, opts.activityText),
	}); err != nil {
		return "", err
	}
	goalCompleted := false
	if !isSideQuest {
		goalCompleted, err = NewGoalUsecase(uc.repo).RecordActivity(ctx, userID, now, goalActivityTextWithFallback(workout, opts.activityText))
		if err != nil {
			return "", err
		}
	}

	isComeback := isFullReport && report.InactiveDays > 3 && report.Streak == 1
	var response string

	if isComeback {
		response = fmt.Sprintf("🎉 WELCOME BACK, %s! 🎉\n", name)
		response += fmt.Sprintf("Kamu kembali setelah %d hari absen. Itu butuh keberanian! 💪\n", report.InactiveDays)
		response += "Streak kamu direset, tapi totalmu tetap tersimpan.\n"
		if report.JobClass != "" {
			response += fmt.Sprintf("🧭 Job: %s\n", domain.FormatJobClass(report.JobClass))
		}
		response += fmt.Sprintf("\n📊 Level: Lv.%d • %s (Total: %d pts)\n", report.Level, domain.FormatLevel(report.TotalPoints), report.TotalPoints)
		response += fmt.Sprintf("📅 Total hari aktif: %d\n", report.ActivityCount)

		nextComebackAch := uc.getNextComebackTarget(report)
		if nextComebackAch != nil {
			response += fmt.Sprintf("\n🔥 Comeback Challenge dimulai! Raih %d minggu berturut-turut untuk unlock \"%s\"!", nextComebackAch.MinComebackStreak, nextComebackAch.Name)
		} else {
			response += "\n🔥 Comeback Challenge dimulai! Ayo bangun streak-mu kembali!"
		}
	} else {
		cyclePrefix := ""
		if report.CenturionCycles > 0 {
			cyclePrefix = fmt.Sprintf("[C%d] ", report.CenturionCycles+1)
		}

		// Only the final point result is shown; the per-component breakdown
		// (base / weekly / daily / seasonal / repeat ½) is intentionally hidden
		// from the user-facing message for both /lapor and /lapor sidequest.
		expBreakdown := fmt.Sprintf("⭐ +%d pts", reportPoints)
		if len(statGains) > 0 {
			expBreakdown += fmt.Sprintf("\n💪 Attributes: %s", strings.Join(statGains, ", "))
		}

		if isSideQuest {
			response = fmt.Sprintf("Side quest diterima, %s%s menyelesaikan %d side quest. Ini bonus terpisah dari 3 slot laporan utama harian. 🔥\n%s",
				cyclePrefix, name, opts.sideQuestCount, expBreakdown)
		} else if isRepeatReport {
			response = fmt.Sprintf("Laporan diterima (laporan ke-%d hari ini), %s%s sudah berkeringat %d hari. Lanjutkan 🔥 (streak %d minggu)\n%s",
				dailyCount+1, cyclePrefix, name, report.ActivityCount, report.Streak, expBreakdown)
		} else {
			response = fmt.Sprintf("Laporan diterima, %s%s sudah berkeringat %d hari. Lanjutkan 🔥 (streak %d minggu)\n%s",
				cyclePrefix, name, report.ActivityCount, report.Streak, expBreakdown)
		}
		if report.JobClass != "" {
			response += fmt.Sprintf("\n🧭 Job: %s", domain.FormatJobClass(report.JobClass))
		}
	}
	if !attributesActive {
		response += "\n\n" + formatJobSetupNotice()
	}

	if isFullReport && report.ActivityCount == 1 && report.CenturionCycles > 0 && !isComeback {
		response = "🔥 *ERA BARU DIMULAI!* 🔥\n" +
			fmt.Sprintf("Selamat %s, Anda telah menyelesaikan 100 hari sebelumnya.\n", name) +
			fmt.Sprintf("Sekarang Anda memulai *Siklus %d (Hari ke-1)*. Terus jaga konsistensi! 💪\n\n", report.CenturionCycles+1) +
			response
	} else if isFullReport && report.ActivityCount == 100 {
		response = "🎊 *LUAR BIASA!* 🎊\n" +
			fmt.Sprintf("Selamat %s, Anda mencapai *HARI KE-100*! 💯\n", name) +
			"Anda sekarang resmi menyandang gelar *CENTURION 🛡️*. Nama Anda telah diabdikan dalam jajaran legenda grup!\n\n" +
			response
	}

	if workout != nil {
		response += "\n" + uc.formatWorkout(workout)
	}

	if streakFreezeUsed {
		response += fmt.Sprintf("\n\n❄️ *Streak Freeze terpakai!* Streak kamu aman. (Sisa freeze: %d)", report.StreakFreezes)
	}

	if newRecord {
		response += fmt.Sprintf("\n\n🏆 New Personal Best Streak: %d minggu!", report.MaxStreak)
	}

	if goalCompleted {
		response += "\n\n🎯 *Goal minggu ini tercapai!* Konsistensi harianmu sudah sesuai target. Mantap, champ! 🏆"
		if uc.goalNotifier != nil {
			goalsCompleted := 1
			if report.GoalsCompleted > 0 {
				goalsCompleted = report.GoalsCompleted + 1
			}
			uc.goalNotifier(ctx, userID, name, "", 0, goalsCompleted)
		}
	}

	if leveledUp {
		response += fmt.Sprintf("\n\n⚔️ *LEVEL UP!* Lv.%d → Lv.%d", oldNumericLevel, report.Level)
	}
	if lifetimeTierUp {
		response += fmt.Sprintf("\n🎖️ *TIER LIFETIME UP!* %s %s → %s %s", oldLifetimeTier.Name, oldLifetimeTier.Icon, newLifetimeTier.Name, newLifetimeTier.Icon)
	}
	if seasonRankUp {
		response += fmt.Sprintf("\n🏹 *RANK SEASON UP!* %s %s → %s %s", oldSeasonRank.Name, oldSeasonRank.Icon, newSeasonRank.Name, newSeasonRank.Icon)
	}

	if len(newAchievements)+len(comebackAchievements) > 0 {
		response += "\n\n🏅 *Badge baru:*"
		for _, ach := range newAchievements {
			response += fmt.Sprintf("\n%s %s (+%d pts)", ach.DisplayEmoji, ach.Name, ach.Points)
		}
		for _, ach := range comebackAchievements {
			response += fmt.Sprintf("\n%s %s (+%d pts)", ach.DisplayEmoji, ach.Name, ach.Points)
		}

		if freezeAwarded {
			response += fmt.Sprintf("\n\n❄️ Bonus: +1 Streak Freeze! (Total: %d)", report.StreakFreezes)
		}

		response += "\nDetail badge & cerita unlock: https://lapor-bot.web.id/"
	}

	response += fmt.Sprintf("\n\n💬 _\"%s\"_", RandomQuote())

	if totalPointsGained > 0 {
		response += fmt.Sprintf("\n\n💰 Total: +%d points (Lifetime: %d | Season: %d)", totalPointsGained, report.TotalPoints, report.SeasonalPoints)
	}

	response += fmt.Sprintf("\n%s", domain.FormatNumericLevelProgressBar(report.TotalPoints))
	response += fmt.Sprintf("\n%s", domain.FormatProgressBar(report.TotalPoints))
	if attributesActive {
		response += fmt.Sprintf("\n\n%s", formatCurrentAttributes(report))
	}

	return response, nil
}

func (uc *ReportActivityUsecase) ExecuteYesterday(ctx context.Context, userID, name string, workout *domain.HevyWorkout) (string, error) {
	return uc.executeYesterday(ctx, userID, name, "", workout)
}

func (uc *ReportActivityUsecase) ExecuteYesterdayWithMessage(ctx context.Context, userID, name, message string, workout *domain.HevyWorkout) (string, error) {
	return uc.executeYesterday(ctx, userID, name, message, workout)
}

func (uc *ReportActivityUsecase) executeYesterday(ctx context.Context, userID, name, activityText string, workout *domain.HevyWorkout) (string, error) {
	lock := uc.userLock(userID)
	lock.Lock()
	defer lock.Unlock()

	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	now := time.Now()
	today := domain.GetToday(now)
	yesterday := today.AddDate(0, 0, -1)

	dailyCount, err := uc.getDailyActivityCount(ctx, userID, yesterday, domain.ActivityKindRegularReport)
	if err != nil {
		return "", err
	}
	if dailyCount >= MaxDailyReports {
		return fmt.Sprintf("%s sudah mencapai batas maksimal %dx laporan untuk hari kemarin. 🙏", report.Name, MaxDailyReports), nil
	}

	if report != nil && domain.GetToday(report.LastReportDate).Equal(today) {
		return fmt.Sprintf("%s sudah laporan hari ini. Gunakan /lapor untuk laporan tambahan hari ini. 💪", report.Name), nil
	}

	if report != nil {
		lastReportDate := domain.GetToday(report.LastReportDate)
		daysSinceLastReport := int(math.Round(today.Sub(lastReportDate).Hours() / 24))
		if daysSinceLastReport < 2 {
			return fmt.Sprintf("%s sudah laporan dalam 2 hari terakhir. Tidak perlu lapor ulang untuk hari kemarin. 🙏", report.Name), nil
		}
	}

	streakFreezeUsed := false

	if report != nil {
		storedName := report.Name
		name = reportName(storedName, name)
		shouldUpdateStoredName := isUnknownName(storedName) && strings.TrimSpace(storedName) != name
		if shouldUpdateStoredName {
			report.Name = name
		}

		lastReportDate := domain.GetToday(report.LastReportDate)
		currentWeekStart := domain.GetStartOfISOWeek(yesterday)
		lastWeekStart := domain.GetStartOfISOWeek(lastReportDate)
		weeksSinceLastReport := int(math.Round(currentWeekStart.Sub(lastWeekStart).Hours() / (24 * 7)))

		if currentWeekStart.Equal(lastWeekStart) {
		} else if weeksSinceLastReport == 1 {
			report.Streak++
			report.ComebackStreak++
		} else if weeksSinceLastReport == 2 && report.StreakFreezes > 0 {
			report.StreakFreezes--
			report.Streak++
			report.ComebackStreak++
			streakFreezeUsed = true
		} else {
			report.InactiveDays = int(math.Round(today.Sub(lastReportDate).Hours()/24)) - 1
			report.ComebackStreak = 1
			report.Streak = 1
		}
		report.ActivityCount++
		report.SeasonalActivityCount++
		if report.Streak > report.SeasonalMaxStreak {
			report.SeasonalMaxStreak = report.Streak
		}

		if report.ActivityCount > 100 {
			report.ActivityCount = 1
			report.CenturionCycles++
		}

		report.LastReportDate = now.AddDate(0, 0, -1)
		name = report.Name
	} else {
		name = reportName("", name)
		report = &domain.Report{
			UserID:                userID,
			Name:                  name,
			Streak:                1,
			ActivityCount:         1,
			SeasonalActivityCount: 1,
			SeasonalMaxStreak:     1,
			LastReportDate:        now.AddDate(0, 0, -1),
			MaxStreak:             0,
			TotalPoints:           0,
			SeasonalPoints:        0,
			Achievements:          "",
			SeasonalAchievements:  "",
			ComebackStreak:        1,
			InactiveDays:          0,
			StreakFreezes:         1,
		}
	}

	if report.MaxStreak == 0 {
		report.MaxStreak = 1
	}
	oldNumericLevel := domain.NumericLevelFromTotalPoints(report.TotalPoints)

	newRecord := false
	if report.Streak > report.MaxStreak {
		report.MaxStreak = report.Streak
		if report.Streak > 1 {
			newRecord = true
		}
	}

	basePoints := 5
	weeklyStreakBonus := 0
	seasonalFirstBonus := 0
	if report.Streak > 1 {
		// Yesterday catch-up reports earn ½ XP: halve the (already capped)
		// weekly streak bonus so inflation stays bounded here too.
		weeklyStreakBonus = cappedStreakBonus(int64(report.Streak)-1, weeklyStreakBonusCap, weeklyStreakBonusPerStep) / 2
		if weeklyStreakBonus < 0 {
			weeklyStreakBonus = 0
		}
	}
	if report.SeasonalActivityCount == 1 {
		seasonalFirstBonus = 2
	}
	streakBonus := weeklyStreakBonus
	reportPoints := basePoints + streakBonus + seasonalFirstBonus
	report.TotalPoints += reportPoints
	report.SeasonalPoints += reportPoints

	activityForParse := activityText
	if workout != nil {
		activityForParse += " " + workout.Title
		for _, ex := range workout.Exercises {
			activityForParse += " " + ex
		}
	}
	attributesActive := hasSelectedJob(report)
	var statGains []string
	if attributesActive {
		// A report grants a single, fair attribute point. The activity directs
		// which attribute is rewarded; the job breaks ties among multiple
		// matches and is the fallback when nothing matches. Granting +1 to
		// every matched attribute would make mixed sessions worth several
		// times the attribute points of focused ones.
		attrs, _ := domain.ResolveReportAttributes(activityForParse, report.JobClass)
		chosen := domain.SelectReportAttribute(attrs, report.JobClass,
			attributeSelectionSeed(userID, domain.ActivityKindRegularReport, yesterday.Format(time.DateOnly), dailyCount+1, activityForParse))
		statGains = applyAttributeGains(report, []domain.AttributeType{chosen}, AttributeGainPerReport)
	}

	newAchievements := domain.CheckNewSeasonAchievements(report)
	pointsGained := 0

	for _, ach := range newAchievements {
		report.SeasonalAchievements = domain.AddAchievement(report.SeasonalAchievements, ach.ID)
		if !domain.HasAchievement(report.Achievements, ach.ID) {
			report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
		}
		report.TotalPoints += ach.Points
		report.SeasonalPoints += ach.Points
		pointsGained += ach.Points
	}

	comebackAchievements := domain.CheckComebackAchievements(report)
	for _, ach := range comebackAchievements {
		report.Achievements = domain.AddAchievement(report.Achievements, ach.ID)
		report.TotalPoints += ach.Points
		pointsGained += ach.Points
	}

	freezeAwarded := false
	for _, ach := range newAchievements {
		if ach.ID == "streak_4" && report.StreakFreezes < 2 {
			report.StreakFreezes++
			freezeAwarded = true
		}
	}

	report.Level = domain.NumericLevelFromTotalPoints(report.TotalPoints)
	leveledUp := report.Level > oldNumericLevel
	totalPointsGained := reportPoints + pointsGained

	if err := uc.upsertReportWithActivity(ctx, report, reportActivityEventInput{
		activityDate:        yesterday,
		kind:                domain.ActivityKindRegularReport,
		occurredAt:          now,
		pointsDelta:         totalPointsGained,
		regularCountDelta:   1,
		sideQuestCountDelta: 0,
		activityText:        goalActivityTextWithFallback(workout, activityText),
	}); err != nil {
		return "", err
	}
	goalCompleted, err := NewGoalUsecase(uc.repo).RecordActivity(ctx, userID, yesterday, goalActivityText(workout))
	if err != nil {
		return "", err
	}

	isComeback := report.InactiveDays > 3 && report.Streak == 1
	var response string

	if isComeback {
		response = fmt.Sprintf("🎉 WELCOME BACK, %s! 🎉\n", name)
		response += fmt.Sprintf("Kamu kembali setelah %d hari absen. Itu butuh keberanian! 💪\n", report.InactiveDays)
		response += "Streak kamu direset, tapi totalmu tetap tersimpan.\n"
		if report.JobClass != "" {
			response += fmt.Sprintf("🧭 Job: %s\n", domain.FormatJobClass(report.JobClass))
		}
		response += fmt.Sprintf("\n📊 Level: Lv.%d • %s (Total: %d pts)\n", report.Level, domain.FormatLevel(report.TotalPoints), report.TotalPoints)
		response += fmt.Sprintf("📅 Total hari aktif: %d\n", report.ActivityCount)

		nextComebackAch := uc.getNextComebackTarget(report)
		if nextComebackAch != nil {
			response += fmt.Sprintf("\n🔥 Comeback Challenge dimulai! Raih %d minggu berturut-turut untuk unlock \"%s\"!", nextComebackAch.MinComebackStreak, nextComebackAch.Name)
		} else {
			response += "\n🔥 Comeback Challenge dimulai! Ayo bangun streak-mu kembali!"
		}
	} else {
		cyclePrefix := ""
		if report.CenturionCycles > 0 {
			cyclePrefix = fmt.Sprintf("[C%d] ", report.CenturionCycles+1)
		}

		// Only the final point result is shown; the per-component breakdown is
		// intentionally hidden from the user-facing message.
		expBreakdown := fmt.Sprintf("⭐ +%d pts", reportPoints)
		expBreakdown += " (lapor kemarin)"

		if len(statGains) > 0 {
			expBreakdown += fmt.Sprintf("\n💪 Attributes: %s", strings.Join(statGains, ", "))
		}

		response = fmt.Sprintf("Laporan kemarin diterima, %s%s sudah berkeringat %d hari. Lanjutkan 🔥 (streak %d minggu)\n%s",
			cyclePrefix, name, report.ActivityCount, report.Streak, expBreakdown)
		if report.JobClass != "" {
			response += fmt.Sprintf("\n🧭 Job: %s", domain.FormatJobClass(report.JobClass))
		}
	}
	if !attributesActive {
		response += "\n\n" + formatJobSetupNotice()
	}

	if report.ActivityCount == 1 && report.CenturionCycles > 0 && !isComeback {
		response = "🔥 *ERA BARU DIMULAI!* 🔥\n" +
			fmt.Sprintf("Selamat %s, Anda telah menyelesaikan 100 hari sebelumnya.\n", name) +
			fmt.Sprintf("Sekarang Anda memulai *Siklus %d (Hari ke-1)*. Terus jaga konsistensi! 💪\n\n", report.CenturionCycles+1) +
			response
	} else if report.ActivityCount == 100 {
		response = "🎊 *LUAR BIASA!* 🎊\n" +
			fmt.Sprintf("Selamat %s, Anda mencapai *HARI KE-100*! 💯\n", name) +
			"Anda sekarang resmi menyandang gelar *CENTURION 🛡️*. Nama Anda telah diabdikan dalam jajaran legenda grup!\n\n" +
			response
	}

	if workout != nil {
		response += "\n" + uc.formatWorkout(workout)
	}

	if streakFreezeUsed {
		response += fmt.Sprintf("\n\n❄️ *Streak Freeze terpakai!* Streak kamu aman. (Sisa freeze: %d)", report.StreakFreezes)
	}

	if newRecord {
		response += fmt.Sprintf("\n\n🏆 New Personal Best Streak: %d minggu!", report.MaxStreak)
	}

	if goalCompleted {
		response += "\n\n🎯 *Goal minggu ini tercapai!* Konsistensi harianmu sudah sesuai target. Mantap, champ! 🏆"
		if uc.goalNotifier != nil {
			goalsCompleted := 1
			if report.GoalsCompleted > 0 {
				goalsCompleted = report.GoalsCompleted + 1
			}
			uc.goalNotifier(ctx, userID, name, "", 0, goalsCompleted)
		}
	}

	if leveledUp {
		response += fmt.Sprintf("\n\n⚔️ *LEVEL UP!* Lv.%d → Lv.%d", oldNumericLevel, report.Level)
	}

	if len(newAchievements)+len(comebackAchievements) > 0 {
		response += "\n\n🏅 *Badge baru:*"
		for _, ach := range newAchievements {
			response += fmt.Sprintf("\n%s %s (+%d pts)", ach.DisplayEmoji, ach.Name, ach.Points)
		}
		for _, ach := range comebackAchievements {
			response += fmt.Sprintf("\n%s %s (+%d pts)", ach.DisplayEmoji, ach.Name, ach.Points)
		}

		if freezeAwarded {
			response += fmt.Sprintf("\n\n❄️ Bonus: +1 Streak Freeze! (Total: %d)", report.StreakFreezes)
		}

		response += "\nDetail badge & cerita unlock: https://lapor-bot.web.id/"
	}

	response += fmt.Sprintf("\n\n💬 _\"%s\"_", RandomQuote())

	if totalPointsGained > 0 {
		response += fmt.Sprintf("\n\n💰 Total: +%d points (Lifetime: %d | Season: %d)", totalPointsGained, report.TotalPoints, report.SeasonalPoints)
	}

	response += fmt.Sprintf("\n%s", domain.FormatNumericLevelProgressBar(report.TotalPoints))
	response += fmt.Sprintf("\n%s", domain.FormatProgressBar(report.TotalPoints))
	if attributesActive {
		response += fmt.Sprintf("\n\n%s", formatCurrentAttributes(report))
	}

	return response, nil
}

func (uc *ReportActivityUsecase) getDailyActivityCount(ctx context.Context, userID string, date time.Time, kind string) (int, error) {
	if repo, ok := uc.repo.(typedActivityRepository); ok {
		return repo.GetDailyActivityCountByKind(ctx, userID, date, kind)
	}
	return uc.repo.GetDailyActivityCount(ctx, userID, date)
}

type reportActivityEventInput struct {
	activityDate        time.Time
	kind                string
	occurredAt          time.Time
	pointsDelta         int
	regularCountDelta   int
	sideQuestCountDelta int
	activityText        string
}

func (uc *ReportActivityUsecase) upsertReportWithActivity(ctx context.Context, report *domain.Report, input reportActivityEventInput) error {
	if repo, ok := uc.repo.(eventActivityRepository); ok && ReportEventLedgerEnabled(input.occurredAt) {
		seasonNumber, _ := GetCurrentSessionInfo(input.occurredAt)
		event := domain.ReportActivityEvent{
			EventID:             reportActivityEventID(report.UserID, input.kind, input.activityDate, input.occurredAt, input.pointsDelta, input.regularCountDelta, input.sideQuestCountDelta),
			UserID:              report.UserID,
			SeasonNumber:        seasonNumber,
			Kind:                input.kind,
			ActivityDate:        input.activityDate,
			OccurredAt:          input.occurredAt,
			PointsDelta:         input.pointsDelta,
			RegularCountDelta:   input.regularCountDelta,
			SideQuestCountDelta: input.sideQuestCountDelta,
			RuleVersion:         1,
			Source:              "whatsapp",
			ActivityText:        input.activityText,
			MetadataJSON:        "{}",
		}
		return repo.UpsertReportWithActivityEvent(ctx, report, event)
	}
	if repo, ok := uc.repo.(typedActivityRepository); ok {
		return repo.UpsertReportWithActivityKind(ctx, report, input.activityDate, input.kind)
	}
	return uc.repo.UpsertReportWithActivity(ctx, report, input.activityDate)
}

// ReportEventLedgerEnabled gates the Season 2 ledger/projection dual-write.
// Schema can be deployed earlier, but event capture starts on the configured
// date in WIB so we avoid partial-day data.
func ReportEventLedgerEnabled(now time.Time) bool {
	loc := time.FixedZone("WIB", 7*3600)
	start := time.Date(ReportEventLedgerYear, ReportEventLedgerMonth, ReportEventLedgerDay, 0, 0, 0, 0, loc)
	return !now.In(loc).Before(start)
}

func reportActivityEventID(userID, kind string, activityDate, occurredAt time.Time, pointsDelta, regularCountDelta, sideQuestCountDelta int) string {
	seed := fmt.Sprintf("%s|%s|%s|%s|%d|%d|%d",
		userID,
		kind,
		activityDate.Format(time.DateOnly),
		occurredAt.UTC().Format(time.RFC3339Nano),
		pointsDelta,
		regularCountDelta,
		sideQuestCountDelta,
	)
	sum := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(sum[:])
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// cappedStreakBonus computes a capped linear streak bonus.
// steps is the streak length minus one (the number of "extra" periods), clamped
// to [0, cap]; the bonus is steps * perStep. Capping prevents a long streak
// from dominating the fixed base reward (runaway XP).
func cappedStreakBonus(steps, cap, perStep int64) int {
	if steps < 0 {
		steps = 0
	}
	if steps > cap {
		steps = cap
	}
	return int(steps * perStep)
}

// attributeSelectionSeed builds a stable, per-report-slot seed used to
// distribute attribute gains fairly for jobs without a single primary
// attribute (Mage). The same report context always yields the same seed, so
// the chosen attribute is deterministic; varying the slot/date/activity
// spreads Mage's gains across its candidate attributes over time.
func attributeSelectionSeed(userID, kind, date string, slot int, activityText string) string {
	return fmt.Sprintf("%s|%s|%s|%d|%s", userID, kind, date, slot, activityText)
}

// computeDailyStreak returns the number of consecutive days (ending at today)
// the user has logged any activity. today is counted even though the
// activity_log row for the current report is upserted after scoring, so the
// bonus reflects the streak this report is extending.
func (uc *ReportActivityUsecase) computeDailyStreak(ctx context.Context, userID string, today time.Time) int {
	dates, err := uc.repo.GetUserActivityDates(ctx, userID)
	if err != nil {
		// Fail open: treat as the first day so we never block a report on a
		// read error. The bonus is simply not awarded.
		return 1
	}
	return dailyStreakFromDates(dates, today)
}

// dailyStreakFromDates counts consecutive calendar days ending at today,
// including today itself. priorDates are the user's historical activity
// dates (which do not yet contain today for the current report).
func dailyStreakFromDates(priorDates []time.Time, today time.Time) int {
	active := make(map[string]bool, len(priorDates)+1)
	for _, d := range priorDates {
		active[d.Format(time.DateOnly)] = true
	}
	active[today.Format(time.DateOnly)] = true

	streak := 0
	for d := today; active[d.Format(time.DateOnly)]; d = d.AddDate(0, 0, -1) {
		streak++
	}
	return streak
}

func reportName(storedName, incomingName string) string {
	storedName = strings.TrimSpace(storedName)
	incomingName = strings.TrimSpace(incomingName)

	if !isUnknownName(storedName) {
		return storedName
	}
	if !isUnknownName(incomingName) {
		return incomingName
	}
	return "Teman"
}

func isUnknownName(name string) bool {
	clean := strings.TrimSpace(name)
	return clean == "" ||
		clean == "-" ||
		strings.EqualFold(clean, "unknown") ||
		strings.EqualFold(clean, "teman") ||
		strings.EqualFold(clean, "user")
}

func (uc *ReportActivityUsecase) formatWorkout(workout *domain.HevyWorkout) string {
	if workout == nil {
		return ""
	}

	res := fmt.Sprintf("\n🏋️‍♂️ *%s Detail:*\n", workout.Title)
	if len(workout.Exercises) > 0 {
		res += "\n"
		for _, ex := range workout.Exercises {
			res += fmt.Sprintf("%s\n", ex)
		}
	}
	if workout.Time != "" {
		res += fmt.Sprintf("... ⏱️ Time: %s", workout.Time)
	}
	return res
}

func (uc *ReportActivityUsecase) getNextComebackTarget(report *domain.Report) *domain.ComebackAchievement {
	for i := range domain.AllComebackAchievements {
		a := &domain.AllComebackAchievements[i]
		if !domain.HasAchievement(report.Achievements, a.ID) &&
			report.InactiveDays >= a.MinInactiveDays {
			return a
		}
	}
	return nil
}

func applyAttributeGains(report *domain.Report, attrs []domain.AttributeType, statPoints int) []string {
	if report == nil || statPoints <= 0 || len(attrs) == 0 {
		return nil
	}

	var statGains []string
	for _, attr := range attrs {
		switch attr {
		case domain.AttrStr:
			report.Str = domain.ClampedAttribute(report.Str) + statPoints
			statGains = append(statGains, fmt.Sprintf("STR +%d", statPoints))
		case domain.AttrSta:
			report.Sta = domain.ClampedAttribute(report.Sta) + statPoints
			statGains = append(statGains, fmt.Sprintf("STA +%d", statPoints))
		case domain.AttrAgi:
			report.Agi = domain.ClampedAttribute(report.Agi) + statPoints
			statGains = append(statGains, fmt.Sprintf("AGI +%d", statPoints))
		case domain.AttrVit:
			report.Vit = domain.ClampedAttribute(report.Vit) + statPoints
			statGains = append(statGains, fmt.Sprintf("VIT +%d", statPoints))
		}
	}
	return statGains
}

func hasSelectedJob(report *domain.Report) bool {
	return report != nil && strings.TrimSpace(report.JobClass) != ""
}

func formatJobSetupNotice() string {
	return fmt.Sprintf("ℹ️ Atribut belum ditampilkan karena kamu belum memilih job. Laporan tetap masuk; setup job di dashboard agar atribut aktif: %s", jobSetupURL)
}

func formatCurrentAttributes(report *domain.Report) string {
	return fmt.Sprintf("🛡️ Stats: STR %d | STA %d | AGI %d | VIT %d",
		domain.ClampedAttribute(report.Str),
		domain.ClampedAttribute(report.Sta),
		domain.ClampedAttribute(report.Agi),
		domain.ClampedAttribute(report.Vit),
	)
}
