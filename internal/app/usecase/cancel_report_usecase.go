package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type CancelReportUsecase struct {
	repo domain.ReportRepository
}

type walkCancelRepository interface {
	DeleteLatestWalkSteps(ctx context.Context, userID string, date time.Time) error
}

// subtractWalkRepository reverses an exact walk-step amount. Implemented by
// the sqlite repository; used by cancel flows that know the reversed amount
// precisely (report_events metadata or parsed activity text).
type subtractWalkRepository interface {
	SubtractWalkSteps(ctx context.Context, userID string, date time.Time, amount int) error
}

// reportEventQueryRepository exposes the per-day report ledger so cancellations
// can reverse the exact points/steps that were granted instead of estimates.
type reportEventQueryRepository interface {
	GetReportEvents(ctx context.Context, userID string, activityDate time.Time, kind string) ([]domain.ReportActivityEvent, error)
}

// cancelActivityRepository persists a cancellation atomically: activity_logs
// counter decrement/delete + user_reports upsert + compensating ledger event
// + projection reversal, all in one transaction.
type cancelActivityRepository interface {
	CancelActivityEvents(ctx context.Context, report *domain.Report, cancelEvent domain.ReportActivityEvent, kind string, deleteAll bool) (remaining int, err error)
}

// cancelEventSource marks compensating ledger entries written by cancellations.
const cancelEventSource = "cancel"

// repeatCancelPoints is what a same-day repeat report granted (base halved),
// mirroring the report flow's `reportPoints / 2` for repeats. Used as the
// reversal amount when no ledger event exists for the cancelled report.
const repeatCancelPoints = baseReportPoints / 2

func NewCancelReportUsecase(repo domain.ReportRepository) *CancelReportUsecase {
	return &CancelReportUsecase{repo: repo}
}

func (uc *CancelReportUsecase) Execute(ctx context.Context, userID, name string) (string, error) {
	return uc.cancelToday(ctx, userID, name, domain.ActivityKindRegularReport, false)
}

func (uc *CancelReportUsecase) ExecuteAll(ctx context.Context, userID, name string) (string, error) {
	return uc.cancelToday(ctx, userID, name, domain.ActivityKindRegularReport, true)
}

func (uc *CancelReportUsecase) ExecuteSideQuest(ctx context.Context, userID, name string) (string, error) {
	return uc.cancelToday(ctx, userID, name, domain.ActivityKindSideQuest, false)
}

func (uc *CancelReportUsecase) ExecuteAllSideQuest(ctx context.Context, userID, name string) (string, error) {
	return uc.cancelToday(ctx, userID, name, domain.ActivityKindSideQuest, true)
}

func (uc *CancelReportUsecase) cancelToday(ctx context.Context, userID, name string, kind string, all bool) (string, error) {
	report, err := uc.repo.GetReport(ctx, userID)
	if err != nil {
		return "", err
	}

	if report == nil {
		return fmt.Sprintf("Halo %s, kamu belum pernah laporan. Belum ada yang bisa dibatalkan.", name), nil
	}

	now := time.Now()
	today := domain.GetToday(now)

	dailyCount, err := uc.repo.GetDailyActivityCountByKind(ctx, userID, today, kind)
	if err != nil {
		return "", err
	}
	if dailyCount == 0 {
		return fmt.Sprintf("Halo %s, tidak menemukan %s untuk hari ini.", report.Name, cancelItemLabel(kind)), nil
	}

	if kind == domain.ActivityKindSideQuest {
		return uc.cancelSideQuestToday(ctx, report, today, now, dailyCount, all)
	}

	dates, err := uc.repo.GetUserActivityDatesByKind(ctx, userID, kind)
	if err != nil {
		return "", err
	}
	if len(dates) == 0 {
		return fmt.Sprintf("Halo %s, tidak ada aktivitas yang tercatat.", report.Name), nil
	}

	if !containsDate(dates, today) {
		return fmt.Sprintf("Halo %s, tidak menemukan laporan untuk hari ini.", report.Name), nil
	}

	if !all && dailyCount > 1 {
		return uc.cancelRepeatReport(ctx, report, today, now, dates, dailyCount)
	}
	return uc.cancelFullDay(ctx, report, today, now, dates, dailyCount, all)
}

// cancelRepeatReport cancels only the latest of today's regular reports. The
// day stays active, streaks and counters are untouched, and exactly the
// points/steps that the cancelled report granted are reversed (from the
// ledger when available, else the repeat-points constant).
func (uc *CancelReportUsecase) cancelRepeatReport(ctx context.Context, report *domain.Report, today, now time.Time, dates []time.Time, dailyCount int) (string, error) {
	events := uc.todayEvents(ctx, report.UserID, today, domain.ActivityKindRegularReport)

	pointsReversed, walkReversed, texts := repeatReversalAmounts(events)
	if len(events) == 0 {
		// Pre-ledger fallback: a same-day repeat always granted base/2 points.
		pointsReversed = repeatCancelPoints
		if text := uc.latestActivityText(ctx, report.UserID, today, domain.ActivityKindRegularReport); text != "" {
			texts = []string{text}
			if steps, ok := domain.ParseWalkReport(text); ok && steps > 0 {
				walkReversed = steps
			}
		}
	}

	applyPointReversal(report, pointsReversed)
	applyWalkReversal(report, walkReversed)

	remaining := dailyCount - 1
	persisted := false
	if len(events) > 0 {
		if cr, ok := uc.repo.(cancelActivityRepository); ok {
			cancelEvent := buildCancelEvent(report.UserID, today, now, domain.ActivityKindRegularReport,
				-pointsReversed, -1, 0, -walkReversed, strings.Join(texts, "\n"))
			rem, err := cr.CancelActivityEvents(ctx, report, cancelEvent, domain.ActivityKindRegularReport, false)
			if err != nil {
				return "", err
			}
			remaining = rem
			persisted = true
		}
	}
	if !persisted {
		rem, err := uc.repo.DeleteLatestActivityLogByKind(ctx, report.UserID, today, domain.ActivityKindRegularReport)
		if err != nil {
			return "", err
		}
		remaining = rem
		uc.reverseWalkSteps(ctx, report.UserID, today, walkReversed)
		if err := uc.repo.UpsertReport(ctx, report); err != nil {
			return "", err
		}
	}

	if remaining == 0 {
		// The day emptied despite dailyCount > 1 (inconsistent kind counters
		// on legacy rows). Degrade to the full-day cancellation so counters,
		// streaks and seasonal fields are recalculated from the remaining
		// dates instead of keeping a phantom active day.
		return uc.cancelFullDay(ctx, report, today, now, dates, dailyCount, true)
	}

	msg := fmt.Sprintf("✅ Laporan terakhir hari ini telah dibatalkan, %s.\n\n", report.Name)
	msg += fmt.Sprintf("📌 Sisa laporan hari ini: %d/%d\n", remaining, MaxDailyReports)
	msg += fmt.Sprintf("📅 Total hari aktif tetap: %d\n", report.ActivityCount)
	if pointsReversed > 0 {
		msg += fmt.Sprintf("⭐ Poin dikurangi: -%d (Total poin sekarang: %d)\n", pointsReversed, report.TotalPoints)
	} else {
		msg += fmt.Sprintf("⭐ Total poin sekarang: %d\n", report.TotalPoints)
	}
	msg += "\nKalau ingin menghapus semua laporan hari ini, ketik /cancel-all."
	return msg, nil
}

// cancelFullDay cancels all of today's regular reports. The active day is
// removed: counters and streaks are recalculated from the remaining dates,
// while points and seasonal progress are PRESERVED minus exactly what today's
// reports granted (ledger-exact when available) — a cancel must not wipe a
// season's worth of points, badges and walk totals.
func (uc *CancelReportUsecase) cancelFullDay(ctx context.Context, report *domain.Report, today, now time.Time, dates []time.Time, dailyCount int, all bool) (string, error) {
	remainingDates := removeDate(dates, today)
	if len(remainingDates) == len(dates) {
		return fmt.Sprintf("Halo %s, tidak menemukan laporan untuk hari ini.", report.Name), nil
	}

	events := uc.todayEvents(ctx, report.UserID, today, domain.ActivityKindRegularReport)
	pointsReversed, walkReversed, texts := netReversalAmounts(events)
	if len(events) == 0 {
		// Pre-ledger fallback estimate: first report base + repeat halves.
		// Streak bonuses bundled into today's grants are not knowable here, so
		// the estimate can under-reverse; ledger-days cancel exactly.
		pointsReversed = baseReportPoints + repeatCancelPoints*(dailyCount-1)
		if text := uc.latestActivityText(ctx, report.UserID, today, domain.ActivityKindRegularReport); text != "" {
			texts = []string{text}
			if steps, ok := domain.ParseWalkReport(text); ok && steps > 0 {
				walkReversed = steps
			}
		}
	}

	var newReport *domain.Report
	if len(remainingDates) == 0 {
		newReport = &domain.Report{
			UserID:          report.UserID,
			Name:            report.Name,
			ActivityCount:   0,
			LastReportDate:  time.Time{},
			Streak:          0,
			MaxStreak:       0,
			TotalPoints:     0,
			Achievements:    "",
			ComebackStreak:  0,
			InactiveDays:    0,
			CenturionCycles: 0,
			// Walk totals survive minus today's reversed steps: steps from
			// earlier days were earned and are not cancelled away.
			WalkStepsTotal:    max(0, report.WalkStepsTotal-walkReversed),
			SeasonalWalkSteps: max(0, report.SeasonalWalkSteps-walkReversed),
		}
	} else {
		// Date-derived counters/streaks only — points, badges and seasonal
		// fields are preserved below (minus today's reversal). Achievements
		// are intentionally NOT re-evaluated here: awarding badges during a
		// cancellation was a backdoor that made retired badges unlockable.
		newReport = recalculateReportFromDates(report.UserID, report.Name, remainingDates)
		newReport.TotalPoints = max(0, report.TotalPoints-pointsReversed)
		newReport.SeasonalPoints = max(0, report.SeasonalPoints-pointsReversed)
		newReport.SeasonalActivityCount = max(0, report.SeasonalActivityCount-1)
		newReport.SeasonalMaxStreak = seasonalMaxStreakFromDates(remainingDates, now)
		newReport.SeasonalAchievements = report.SeasonalAchievements
		newReport.Achievements = report.Achievements
		newReport.WalkStepsTotal = max(0, report.WalkStepsTotal-walkReversed)
		newReport.SeasonalWalkSteps = max(0, report.SeasonalWalkSteps-walkReversed)
		newReport.Level = domain.NumericLevelFromTotalPoints(newReport.TotalPoints)
	}
	preserveNonReportFields(newReport, report)

	persisted := false
	if len(events) > 0 {
		if cr, ok := uc.repo.(cancelActivityRepository); ok {
			cancelEvent := buildCancelEvent(report.UserID, today, now, domain.ActivityKindRegularReport,
				-pointsReversed, -dailyCount, 0, -walkReversed, strings.Join(texts, "\n"))
			if _, err := cr.CancelActivityEvents(ctx, newReport, cancelEvent, domain.ActivityKindRegularReport, true); err != nil {
				return "", err
			}
			persisted = true
		}
	}
	if !persisted {
		if err := uc.repo.DeleteActivityLogByKind(ctx, report.UserID, today, domain.ActivityKindRegularReport); err != nil {
			return "", err
		}
		uc.reverseWalkSteps(ctx, report.UserID, today, walkReversed)
		if err := uc.repo.UpsertReport(ctx, newReport); err != nil {
			return "", err
		}
	}

	msg := fmt.Sprintf("✅ Semua laporan hari ini telah dibatalkan, %s.\n\n", report.Name)
	if !all {
		msg = fmt.Sprintf("✅ Laporan hari ini telah dibatalkan, %s.\n\n", report.Name)
	}
	msg += fmt.Sprintf("📅 Total hari aktif: %d\n", newReport.ActivityCount)
	msg += fmt.Sprintf("🔥 Streak saat ini: %d minggu\n", newReport.Streak)
	if pointsReversed > 0 {
		msg += fmt.Sprintf("⭐ Poin dikurangi: -%d (Total poin: %d)\n", pointsReversed, newReport.TotalPoints)
	} else {
		msg += fmt.Sprintf("⭐ Total poin: %d\n", newReport.TotalPoints)
	}

	if newReport.InactiveDays > 0 {
		msg += fmt.Sprintf("🔄 Comeback streak: %d minggu\n", newReport.ComebackStreak)
	}

	msg += "\n_Kamu bisa lapor lagi hari ini dengan /lapor._"
	return msg, nil
}

// cancelSideQuestToday cancels today's side quests and reverses everything the
// cancelled quests granted: the difficulty points, the side-quest counters,
// the quest completion state in today's quest list (so the task can be
// completed again), and any walk steps credited through the easy cardio quest.
func (uc *CancelReportUsecase) cancelSideQuestToday(ctx context.Context, report *domain.Report, today, now time.Time, dailyCount int, all bool) (string, error) {
	cancelAll := all || dailyCount == 1
	events := uc.todayEvents(ctx, report.UserID, today, domain.ActivityKindSideQuest)

	countReversed := 1
	var pointsReversed, walkReversed int
	var texts []string
	if cancelAll {
		countReversed = dailyCount
		pointsReversed, walkReversed, texts = netReversalAmounts(events)
	} else {
		pointsReversed, walkReversed, texts = repeatReversalAmounts(events)
	}
	if len(texts) == 0 {
		// Pre-ledger fallback: read the newest stored text BEFORE the logs are
		// deleted (the old flow read it after deletion and could match nothing).
		if text := uc.latestActivityText(ctx, report.UserID, today, domain.ActivityKindSideQuest); text != "" {
			texts = []string{text}
		}
	}

	// Revert the quest completion state first (separate daily_quests storage).
	// The reverted difficulty points double as the exact points fallback when
	// no ledger events exist for today.
	revertedPoints, revertedNames := uc.revertQuestCompletions(ctx, report.UserID, today, texts)
	if len(events) == 0 {
		if revertedPoints > 0 {
			pointsReversed = revertedPoints
		}
		if walkReversed <= 0 {
			walkReversed = walkStepsFromTexts(texts)
		}
	}

	decrementSideQuestCount(report, countReversed)
	applyPointReversal(report, pointsReversed)
	applyWalkReversal(report, walkReversed)

	remaining := 0
	if !cancelAll {
		remaining = dailyCount - 1
	}
	persisted := false
	if len(events) > 0 {
		if cr, ok := uc.repo.(cancelActivityRepository); ok {
			cancelEvent := buildCancelEvent(report.UserID, today, now, domain.ActivityKindSideQuest,
				-pointsReversed, 0, -countReversed, -walkReversed, strings.Join(texts, "\n"))
			rem, err := cr.CancelActivityEvents(ctx, report, cancelEvent, domain.ActivityKindSideQuest, cancelAll)
			if err != nil {
				return "", err
			}
			remaining = rem
			persisted = true
		}
	}
	if !persisted {
		if cancelAll {
			if err := uc.repo.DeleteActivityLogByKind(ctx, report.UserID, today, domain.ActivityKindSideQuest); err != nil {
				return "", err
			}
		} else {
			rem, err := uc.repo.DeleteLatestActivityLogByKind(ctx, report.UserID, today, domain.ActivityKindSideQuest)
			if err != nil {
				return "", err
			}
			remaining = rem
		}
		uc.reverseWalkSteps(ctx, report.UserID, today, walkReversed)
		if err := uc.repo.UpsertReport(ctx, report); err != nil {
			return "", err
		}
	}

	var msg string
	switch {
	case cancelAll && all:
		msg = fmt.Sprintf("✅ Semua side quest hari ini telah dibatalkan, %s.\n\n", report.Name)
	case cancelAll:
		msg = fmt.Sprintf("✅ Side quest hari ini telah dibatalkan, %s.\n\n", report.Name)
	default:
		msg = fmt.Sprintf("✅ Side quest terakhir hari ini telah dibatalkan, %s.\n\n", report.Name)
	}
	if !cancelAll {
		msg += fmt.Sprintf("📌 Sisa side quest hari ini: %d/%d\n", remaining, MaxDailySideQuests)
	}
	msg += fmt.Sprintf("🧩 Total side quest: %d\n", report.TotalSideQuests)
	if len(revertedNames) > 0 {
		msg += fmt.Sprintf("↩️ Quest dikembalikan: %s\n", strings.Join(revertedNames, ", "))
	}
	if pointsReversed > 0 {
		msg += fmt.Sprintf("⭐ Poin dikurangi: -%d (Total poin sekarang: %d)\n", pointsReversed, report.TotalPoints)
	}
	if !cancelAll {
		msg += "\nKalau ingin menghapus semua side quest hari ini, ketik /cancel-all sidequest."
	} else {
		msg += "\n_Kamu bisa lapor side quest lagi hari ini dengan /lapor sidequest._"
	}
	return msg, nil
}

// revertQuestCompletions marks today's quest tasks incomplete again for every
// task named in the cancelled activity texts ("Side quest: A, B (N langkah)")
// and returns the difficulty points of the reverted tasks. Best-effort: when
// the quest list is missing or unparseable nothing is reverted.
func (uc *CancelReportUsecase) revertQuestCompletions(ctx context.Context, userID string, today time.Time, activityTexts []string) (points int, names []string) {
	if len(activityTexts) == 0 {
		return 0, nil
	}
	tasksJSON, err := uc.repo.GetDailyQuest(ctx, userID, today.Format("2006-01-02"))
	if err != nil || tasksJSON == "" {
		return 0, nil
	}
	var tasks []domain.QuestTask
	if err := json.Unmarshal([]byte(tasksJSON), &tasks); err != nil {
		return 0, nil
	}

	changed := false
	for _, text := range activityTexts {
		for _, taskName := range parseSideQuestTaskNames(text) {
			for i := range tasks {
				if tasks[i].Name == taskName && tasks[i].Progress >= tasks[i].Target {
					tasks[i].Progress = 0
					points += sideQuestPoints(tasks[i].Difficulty)
					names = append(names, taskName)
					changed = true
					break
				}
			}
		}
	}
	if !changed {
		return 0, nil
	}

	bytes, err := json.Marshal(tasks)
	if err != nil {
		return 0, nil
	}
	if err := uc.repo.SaveDailyQuest(ctx, userID, today.Format("2006-01-02"), string(bytes)); err != nil {
		return 0, nil
	}
	return points, names
}

// parseSideQuestTaskNames extracts the completed task names from a stored
// side-quest activity text of the form "Side quest: A, B (4.000 langkah)".
func parseSideQuestTaskNames(activityText string) []string {
	text := strings.TrimSpace(activityText)
	if !strings.HasPrefix(text, "Side quest:") {
		return nil
	}
	text = strings.TrimSpace(strings.TrimPrefix(text, "Side quest:"))
	// Strip a trailing annotation such as "(4.000 langkah)".
	if idx := strings.LastIndex(text, " ("); idx >= 0 && strings.HasSuffix(text, ")") {
		text = text[:idx]
	}
	var names []string
	for _, part := range strings.Split(text, ",") {
		if name := strings.TrimSpace(part); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// walkStepsFromTexts sums the walk-step amounts parseable from side-quest
// activity texts (only the easy cardio quest carries steps).
func walkStepsFromTexts(texts []string) int {
	total := 0
	for _, text := range texts {
		if steps, ok := domain.ParseWalkReport(text); ok && steps > 0 {
			total += steps
		}
	}
	return total
}

// todayEvents fetches today's ledger events for exact reversals. Returns nil
// when the repo has no ledger support or the day predates the ledger.
func (uc *CancelReportUsecase) todayEvents(ctx context.Context, userID string, date time.Time, kind string) []domain.ReportActivityEvent {
	qr, ok := uc.repo.(reportEventQueryRepository)
	if !ok {
		return nil
	}
	events, err := qr.GetReportEvents(ctx, userID, date, kind)
	if err != nil {
		return nil
	}
	return events
}

// latestActivityText reads the newest stored activity text for the day/kind.
// Callers must invoke it BEFORE deleting the day's logs.
func (uc *CancelReportUsecase) latestActivityText(ctx context.Context, userID string, date time.Time, kind string) string {
	text, err := uc.repo.GetLatestActivityText(ctx, userID, date, kind)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(text)
}

// reverseWalkSteps removes steps from the day's walk ledger: exactly `amount`
// when known (SubtractWalkSteps), otherwise the marker-based latest addition.
// Unknown amounts reverse nothing — reversing unearned steps is worse than
// leaving a residual.
func (uc *CancelReportUsecase) reverseWalkSteps(ctx context.Context, userID string, date time.Time, amount int) {
	if amount > 0 {
		if sw, ok := uc.repo.(subtractWalkRepository); ok {
			_ = sw.SubtractWalkSteps(ctx, userID, date, amount)
			return
		}
		if wa, ok := uc.repo.(walkCancelRepository); ok {
			_ = wa.DeleteLatestWalkSteps(ctx, userID, date)
		}
	}
}

// buildCancelEvent assembles the compensating ledger entry for a cancellation.
// Deltas are negative so a replay of the ledger nets to the remaining value;
// walkDelta is likewise negated into the metadata's walk_steps field.
func buildCancelEvent(userID string, activityDate, occurredAt time.Time, kind string, pointsDelta, regularDelta, sideQuestDelta, walkDelta int, activityText string) domain.ReportActivityEvent {
	seasonNumber, _ := GetCurrentSessionInfo(occurredAt)
	metadata := "{}"
	if walkDelta != 0 {
		if b, err := json.Marshal(map[string]int{"walk_steps": walkDelta}); err == nil {
			metadata = string(b)
		}
	}
	return domain.ReportActivityEvent{
		EventID:             reportActivityEventID(userID, kind, activityDate, occurredAt, pointsDelta, regularDelta, sideQuestDelta),
		UserID:              userID,
		SeasonNumber:        seasonNumber,
		Kind:                kind,
		ActivityDate:        activityDate,
		OccurredAt:          occurredAt,
		PointsDelta:         pointsDelta,
		RegularCountDelta:   regularDelta,
		SideQuestCountDelta: sideQuestDelta,
		RuleVersion:         1,
		Source:              cancelEventSource,
		ActivityText:        activityText,
		MetadataJSON:        metadata,
	}
}

// repeatReversalAmounts returns the points/steps granted by the LATEST forward
// (non-cancel) event — the counterpart of "delete the latest report". The
// event's own activity text rides along for quest-state reversal.
func repeatReversalAmounts(events []domain.ReportActivityEvent) (points, walk int, texts []string) {
	latest := latestForwardEvent(events)
	if latest == nil {
		return 0, 0, nil
	}
	return max(0, latest.PointsDelta), max(0, walkStepsFromEventMetadata(latest.MetadataJSON)), []string{latest.ActivityText}
}

// netReversalAmounts sums ALL of today's events (forward grants plus earlier
// compensations) so a full-day cancel reverses exactly the remaining credited
// value. Texts come from forward events only.
func netReversalAmounts(events []domain.ReportActivityEvent) (points, walk int, texts []string) {
	if len(events) == 0 {
		return 0, 0, nil
	}
	for _, ev := range events {
		points += ev.PointsDelta
		walk += walkStepsFromEventMetadata(ev.MetadataJSON)
		if ev.Source != cancelEventSource && ev.ActivityText != "" {
			texts = append(texts, ev.ActivityText)
		}
	}
	return max(0, points), max(0, walk), texts
}

func latestForwardEvent(events []domain.ReportActivityEvent) *domain.ReportActivityEvent {
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Source != cancelEventSource && (events[i].RegularCountDelta > 0 || events[i].SideQuestCountDelta > 0) {
			return &events[i]
		}
	}
	return nil
}

// walkStepsFromEventMetadata mirrors the repo's metadata convention
// ({"walk_steps": N}); kept local so the usecase can compute reversals without
// importing repo internals.
func walkStepsFromEventMetadata(metadataJSON string) int {
	if metadataJSON == "" {
		return 0
	}
	var m struct {
		WalkSteps int `json:"walk_steps"`
	}
	if err := json.Unmarshal([]byte(metadataJSON), &m); err != nil {
		return 0
	}
	return m.WalkSteps
}

// applyPointReversal subtracts reversed points from lifetime and seasonal
// totals (floored at 0) and re-derives the numeric level.
func applyPointReversal(report *domain.Report, points int) {
	if points <= 0 {
		return
	}
	report.TotalPoints = max(0, report.TotalPoints-points)
	report.SeasonalPoints = max(0, report.SeasonalPoints-points)
	report.Level = domain.NumericLevelFromTotalPoints(report.TotalPoints)
}

// applyWalkReversal subtracts reversed steps from the profile walk totals
// (floored at 0). The activity_logs ledger is reversed separately.
func applyWalkReversal(report *domain.Report, steps int) {
	if steps <= 0 {
		return
	}
	report.WalkStepsTotal = max(0, report.WalkStepsTotal-steps)
	report.SeasonalWalkSteps = max(0, report.SeasonalWalkSteps-steps)
}

// seasonalMaxStreakFromDates recomputes the season max streak from the
// remaining in-season active dates after a cancellation.
func seasonalMaxStreakFromDates(dates []time.Time, now time.Time) int {
	_, seasonStart := GetCurrentSessionInfo(now)
	weeks := make([]time.Time, 0, len(dates))
	seenWeeks := make(map[string]bool)
	for _, d := range dates {
		if d.Before(seasonStart) {
			continue
		}
		ws := domain.GetStartOfISOWeek(d)
		key := ws.Format(time.DateOnly)
		if !seenWeeks[key] {
			seenWeeks[key] = true
			weeks = append(weeks, ws)
		}
	}
	_, maxStreak, _, _ := calculateStreaksFromWeeks(weeks)
	return maxStreak
}

func decrementSideQuestCount(report *domain.Report, count int) {
	report.TotalSideQuests -= count
	if report.TotalSideQuests < 0 {
		report.TotalSideQuests = 0
	}
	report.SeasonalSideQuests -= count
	if report.SeasonalSideQuests < 0 {
		report.SeasonalSideQuests = 0
	}
}

func cancelItemLabel(kind string) string {
	if kind == domain.ActivityKindSideQuest {
		return "side quest"
	}
	return "laporan"
}

func preserveNonReportFields(next, current *domain.Report) {
	next.JobClass = current.JobClass
	next.StreakFreezes = current.StreakFreezes
	next.GoalsCompleted = current.GoalsCompleted
	next.TotalSideQuests = current.TotalSideQuests
	next.SeasonalSideQuests = current.SeasonalSideQuests
	// Attributes are preserved as-is on cancellation. Reversing them exactly
	// requires knowing which attribute each cancelled report granted — that is
	// recorded in report_events metadata ("attr") for new events, so a future
	// change can reverse attribute gains for ledger-era reports. Pre-ledger
	// grants are not recoverable and stays as-is (known limitation).
	next.Str = current.Str
	next.Sta = current.Sta
	next.Agi = current.Agi
	next.Vit = current.Vit
	next.End = current.End
	next.Luk = current.Luk
}

func removeDate(dates []time.Time, target time.Time) []time.Time {
	var result []time.Time
	for _, d := range dates {
		if !d.Equal(target) {
			result = append(result, d)
		}
	}
	return result
}

func containsDate(dates []time.Time, target time.Time) bool {
	for _, d := range dates {
		if d.Equal(target) {
			return true
		}
	}
	return false
}

// recalculateReportFromDates rebuilds ONLY the date-derived fields (active-day
// counters, streaks, last report date) from the remaining active dates.
// Points, achievements and seasonal progress are owned by the caller: a
// cancellation reverses what the cancelled reports granted instead of
// recomputing — and awarding — achievements from scratch.
func recalculateReportFromDates(userID, name string, dates []time.Time) *domain.Report {
	if len(dates) == 0 {
		return nil
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	totalDates := len(dates)
	lastReportDate := dates[len(dates)-1]

	weeks := make([]time.Time, 0, len(dates))
	seenWeeks := make(map[string]bool)
	for _, d := range dates {
		ws := domain.GetStartOfISOWeek(d)
		key := ws.Format(time.DateOnly)
		if !seenWeeks[key] {
			seenWeeks[key] = true
			weeks = append(weeks, ws)
		}
	}

	streak, maxStreak, comebackStreak, inactiveDays := calculateStreaksFromWeeks(weeks)

	activityCount := totalDates
	centurionCycles := 0
	if totalDates > 100 {
		centurionCycles = (totalDates - 1) / 100
		activityCount = ((totalDates - 1) % 100) + 1
	}

	report := &domain.Report{
		UserID:          userID,
		Name:            name,
		ActivityCount:   activityCount,
		LastReportDate:  lastReportDate,
		Streak:          streak,
		MaxStreak:       maxStreak,
		ComebackStreak:  comebackStreak,
		InactiveDays:    inactiveDays,
		CenturionCycles: centurionCycles,
		TotalPoints:     0,
		Achievements:    "",
	}

	if maxStreak == 0 && streak > 0 {
		report.MaxStreak = streak
	}

	return report
}

func calculateStreaksFromWeeks(weeks []time.Time) (streak, maxStreak, comebackStreak, inactiveDays int) {
	if len(weeks) == 0 {
		return 0, 0, 0, 0
	}

	streak = 1
	maxStreak = 1
	comebackStreak = 1
	inactiveDays = 0

	var lastGapDays int

	for i := 1; i < len(weeks); i++ {
		prevWeek := weeks[i-1]
		currWeek := weeks[i]
		weeksDiff := int(math.Round(currWeek.Sub(prevWeek).Hours() / (24 * 7)))

		if weeksDiff == 1 {
			streak++
			comebackStreak++
		} else {
			if streak > maxStreak {
				maxStreak = streak
			}
			lastGapDays = int(math.Round(currWeek.Sub(prevWeek.AddDate(0, 0, 7)).Hours() / 24))
			streak = 1
			comebackStreak = 1
		}
	}

	if streak > maxStreak {
		maxStreak = streak
	}

	if len(weeks) >= 2 {
		lastTwoDiff := int(math.Round(weeks[len(weeks)-1].Sub(weeks[len(weeks)-2]).Hours() / (24 * 7)))
		if lastTwoDiff > 1 {
			inactiveDays = int(math.Round(weeks[len(weeks)-1].Sub(weeks[len(weeks)-2].AddDate(0, 0, 7)).Hours() / 24))
		} else if lastGapDays > 0 {
			inactiveDays = lastGapDays
		}
	}

	return streak, maxStreak, comebackStreak, inactiveDays
}
