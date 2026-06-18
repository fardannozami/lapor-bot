package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

type MessageResponse struct {
	Text      string
	IsPrivate bool
}

type HandleMessageUsecase struct {
	reportUC            *ReportActivityUsecase
	leaderboardUC       *GetLeaderboardUsecase
	weeklyLeaderboardUC *GetWeeklyLeaderboardUsecase
	myStatsUC           *GetMyStatsUsecase
	achievementsUC      *GetAchievementsUsecase
	comebackUC          *ComebackChallengeUsecase
	cancelUC            *CancelReportUsecase
	updateNameUC        *UpdateNameUsecase
	linkStravaUC        *LinkStravaUsecase
	broadcastUpdateUC   *BroadcastUpdateUsecase
	motivationUC        *GetMotivationUsecase
	helpUC              *GetHelpUsecase
	jobUC               *JobUsecase
	goalUC              *GoalUsecase
	dailyQuestUC        *DailyQuestUsecase
}

func NewHandleMessageUsecase(
	reportUC *ReportActivityUsecase,
	leaderboardUC *GetLeaderboardUsecase,
	myStatsUC *GetMyStatsUsecase,
	achievementsUC *GetAchievementsUsecase,
	comebackUC *ComebackChallengeUsecase,
	cancelUC *CancelReportUsecase,
	updateNameUC *UpdateNameUsecase,
	linkStravaUC *LinkStravaUsecase,
	broadcastUpdateUC *BroadcastUpdateUsecase,
	motivationUC *GetMotivationUsecase,
	helpUC *GetHelpUsecase,
) *HandleMessageUsecase {
	return &HandleMessageUsecase{
		reportUC:            reportUC,
		leaderboardUC:       leaderboardUC,
		weeklyLeaderboardUC: NewGetWeeklyLeaderboardUsecase(leaderboardUC.repo),
		myStatsUC:           myStatsUC,
		achievementsUC:      achievementsUC,
		comebackUC:          comebackUC,
		cancelUC:            cancelUC,
		updateNameUC:        updateNameUC,
		linkStravaUC:        linkStravaUC,
		broadcastUpdateUC:   broadcastUpdateUC,
		motivationUC:        motivationUC,
		helpUC:              helpUC,
		jobUC:               NewJobUsecase(leaderboardUC.repo),
		goalUC:              NewGoalUsecase(leaderboardUC.repo),
		dailyQuestUC:        NewDailyQuestUsecase(leaderboardUC.repo),
	}
}

func (uc *HandleMessageUsecase) Execute(ctx context.Context, userID, name, message string) (MessageResponse, error) {
	msg := strings.ToLower(strings.TrimSpace(message))

	if msg == "" {
		return MessageResponse{}, nil
	}

	if strings.Contains(msg, "#tutorial") {
		text := uc.helpUC.ExecuteTutorial()
		return MessageResponse{Text: text}, nil
	}

	if strings.Contains(msg, "#help") {
		text := uc.helpUC.Execute()
		return MessageResponse{Text: text}, nil
	}

	// if strings.Contains(msg, "#attributes") || strings.Contains(msg, "#attributs") || strings.Contains(msg, "#atributs") {
	// 	text := uc.helpUC.ExecuteAttributes()
	// 	return MessageResponse{Text: text}, nil
	// }

	if strings.Contains(msg, "#lapor-kemarin") {
		workout := domain.ParseHevy(message)
		text, err := uc.reportUC.ExecuteYesterdayWithMessage(ctx, userID, name, message, workout)
		return MessageResponse{Text: text}, err
	}

	// if strings.Contains(msg, "#mysidequest") {
	// 	text, err := uc.dailyQuestUC.ViewQuest(ctx, userID, name, time.Now())
	// 	return MessageResponse{Text: text}, err
	// }

	if arg, ok := extractSideQuestReport(message); ok {
		if arg == "" {
			text, err := uc.dailyQuestUC.ViewQuest(ctx, userID, name, time.Now())
			return MessageResponse{Text: text}, err
		}
		text, err := uc.dailyQuestUC.UpdateProgress(ctx, userID, name, []string{arg}, uc.reportUC, time.Now())
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#lapor") {
		workout := domain.ParseHevy(message)
		text, err := uc.reportUC.ExecuteWithMessage(ctx, userID, name, message, workout)
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#cancel-all") {
		text, err := uc.cancelUC.ExecuteAll(ctx, userID, name)
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#cancel") {
		text, err := uc.cancelUC.Execute(ctx, userID, name)
		return MessageResponse{Text: text}, err
	}

	// if strings.Contains(msg, "#motivasi") {
	// 	text := uc.motivationUC.Execute()
	// 	return MessageResponse{Text: text}, nil
	// }

	// if strings.Contains(msg, "#jobs") {
	// 	return MessageResponse{Text: uc.jobUC.List()}, nil
	// }

	// if strings.Contains(msg, "#goal") {
	// 	text, err := uc.goalUC.Execute(ctx, userID, name, message)
	// 	return MessageResponse{Text: text}, err
	// }

	// if strings.Contains(msg, "#job") {
	// 	idx := strings.Index(msg, "#job")
	// 	jobID := strings.TrimSpace(msg[idx+len("#job"):])
	// 	if jobID == "" {
	// 		return MessageResponse{Text: "Pilih job dengan format: #job <id>. Cek daftar job dengan #jobs."}, nil
	// 	}
	// 	text, err := uc.jobUC.Select(ctx, userID, name, jobID)
	// 	return MessageResponse{Text: text}, err
	// }

	// if strings.Contains(msg, "#setname") {
	// 	idx := strings.Index(msg, "#setname")
	// 	newName := strings.TrimSpace(message[idx+len("#setname"):])
	// 	text, err := uc.updateNameUC.Execute(ctx, userID, newName)
	// 	return MessageResponse{Text: text}, err
	// }

	// if strings.Contains(msg, "#leaderboard-weekly") {
	// 	text, err := uc.weeklyLeaderboardUC.Execute(ctx)
	// 	return MessageResponse{Text: text}, err
	// }

	// if strings.Contains(msg, "#leaderboard-seasonal") {
	// 	text, err := uc.leaderboardUC.ExecuteSeasonal(ctx)
	// 	return MessageResponse{Text: text}, err
	// }

	// if strings.Contains(msg, "#ranks") {
	// 	text, err := uc.leaderboardUC.ExecuteRanks(ctx)
	// 	return MessageResponse{Text: text}, err
	// }

	// if strings.Contains(msg, "#leaderboard") {
	// 	text, err := uc.leaderboardUC.Execute(ctx)
	// 	return MessageResponse{Text: text}, err
	// }

	// if strings.Contains(msg, "#mystats") {
	// 	text, err := uc.myStatsUC.Execute(ctx, userID, name)
	// 	return MessageResponse{Text: text}, err
	// }

	// if strings.Contains(msg, "#achievements") {
	// 	text, err := uc.achievementsUC.Execute(ctx)
	// 	return MessageResponse{Text: text}, err
	// }

	// if strings.Contains(msg, "#comeback") {
	// 	text, err := uc.comebackUC.Execute(ctx, userID, name)
	// 	return MessageResponse{Text: text}, err
	// }

	// if strings.Contains(msg, "#strava") {
	// 	authURL := uc.linkStravaUC.GetAuthURL(userID, name)
	// 	text := fmt.Sprintf("🚴‍♂️ *Integrasi Strava* 🏃‍♂️\n\nKlik link di bawah ini untuk menghubungkan akun Strava kamu:\n\n%s\n\nSetelah berhasil, aktivitas larimu akan otomatis dilaporkan! 🎉", authURL)
	// 	return MessageResponse{Text: text, IsPrivate: true}, nil
	// }

	// if strings.HasPrefix(msg, "!broadcast_update") {
	// 	text := uc.broadcastUpdateUC.Execute()
	// 	return MessageResponse{Text: text}, nil
	// }

	return MessageResponse{
		Text: "📋 Maaf, command yang kamu kirim belum tersedia nih! 😅\n\n" +
			"Coba cek command yang bisa dipakai dengan `#help` atau langsung kunjungi:\n" +
			"🌐 https://lapor-bot.web.id/\n\n" +
			"Di sana kamu bisa lihat klasemen, stats personal, dan info lainnya! ✨",
	}, nil
}

func extractSideQuestReport(message string) (string, bool) {
	lower := strings.ToLower(message)
	for _, command := range []string{"#lapor sidequest", "#lapor-sidequest"} {
		idx := strings.Index(lower, command)
		if idx == -1 {
			continue
		}
		return strings.TrimSpace(message[idx+len(command):]), true
	}
	return "", false
}
