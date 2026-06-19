package usecase

import (
	"context"
	"strings"
	"time"
	"unicode"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

const commandPrefix = "/"

const unknownCommandMessage = "📋 Maaf, command yang kamu kirim belum tersedia nih! 😅\n\n" +
	"Coba cek command yang bisa dipakai dengan `/help` atau langsung kunjungi:\n" +
	"🌐 https://lapor-bot.web.id/\n\n" +
	"Di sana kamu bisa lihat klasemen, stats personal, dan info lainnya! ✨"

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
	trimmedMessage := strings.TrimSpace(message)
	msg := strings.ToLower(trimmedMessage)

	if msg == "" {
		return MessageResponse{}, nil
	}

	if !strings.HasPrefix(msg, commandPrefix) {
		return MessageResponse{}, nil
	}

	if hasCommand(msg, "/tutorial") {
		text := uc.helpUC.ExecuteTutorial()
		return MessageResponse{Text: text}, nil
	}

	if hasCommand(msg, "/help") {
		text := uc.helpUC.Execute()
		return MessageResponse{Text: text}, nil
	}

	// if hasCommand(msg, "/attributes") || hasCommand(msg, "/attributs") || hasCommand(msg, "/atributs") {
	// 	text := uc.helpUC.ExecuteAttributes()
	// 	return MessageResponse{Text: text}, nil
	// }

	if hasCommand(msg, "/lapor-kemarin") {
		workout := domain.ParseHevy(trimmedMessage)
		text, err := uc.reportUC.ExecuteYesterdayWithMessage(ctx, userID, name, trimmedMessage, workout)
		return MessageResponse{Text: text}, err
	}

	// if hasCommand(msg, "/mysidequest") {
	// 	text, err := uc.dailyQuestUC.ViewQuest(ctx, userID, name, time.Now())
	// 	return MessageResponse{Text: text}, err
	// }

	if arg, ok := extractSideQuestReport(trimmedMessage); ok {
		if arg == "" {
			text, err := uc.dailyQuestUC.ViewQuest(ctx, userID, name, time.Now())
			return MessageResponse{Text: text}, err
		}
		text, err := uc.dailyQuestUC.UpdateProgress(ctx, userID, name, []string{arg}, uc.reportUC, time.Now())
		return MessageResponse{Text: text}, err
	}

	if hasCommand(msg, "/lapor") {
		workout := domain.ParseHevy(trimmedMessage)
		text, err := uc.reportUC.ExecuteWithMessage(ctx, userID, name, trimmedMessage, workout)
		return MessageResponse{Text: text}, err
	}

	if hasCommand(msg, "/cancel-all") {
		text, err := uc.cancelUC.ExecuteAll(ctx, userID, name)
		return MessageResponse{Text: text}, err
	}

	if hasCommand(msg, "/cancel") {
		text, err := uc.cancelUC.Execute(ctx, userID, name)
		return MessageResponse{Text: text}, err
	}

	// if hasCommand(msg, "/motivasi") {
	// 	text := uc.motivationUC.Execute()
	// 	return MessageResponse{Text: text}, nil
	// }

	// if hasCommand(msg, "/jobs") {
	// 	return MessageResponse{Text: uc.jobUC.List()}, nil
	// }

	// if hasCommand(msg, "/goal") {
	// 	text, err := uc.goalUC.Execute(ctx, userID, name, message)
	// 	return MessageResponse{Text: text}, err
	// }

	// if hasCommand(msg, "/job") {
	// 	idx := strings.Index(msg, "/job")
	// 	jobID := strings.TrimSpace(msg[idx+len("/job"):])
	// 	if jobID == "" {
	// 		return MessageResponse{Text: "Pilih job dengan format: /job <id>. Cek daftar job dengan /jobs."}, nil
	// 	}
	// 	text, err := uc.jobUC.Select(ctx, userID, name, jobID)
	// 	return MessageResponse{Text: text}, err
	// }

	// if hasCommand(msg, "/setname") {
	// 	idx := strings.Index(msg, "/setname")
	// 	newName := strings.TrimSpace(message[idx+len("/setname"):])
	// 	text, err := uc.updateNameUC.Execute(ctx, userID, newName)
	// 	return MessageResponse{Text: text}, err
	// }

	// if hasCommand(msg, "/leaderboard-weekly") {
	// 	text, err := uc.weeklyLeaderboardUC.Execute(ctx)
	// 	return MessageResponse{Text: text}, err
	// }

	// if hasCommand(msg, "/leaderboard-seasonal") {
	// 	text, err := uc.leaderboardUC.ExecuteSeasonal(ctx)
	// 	return MessageResponse{Text: text}, err
	// }

	// if hasCommand(msg, "/ranks") {
	// 	text, err := uc.leaderboardUC.ExecuteRanks(ctx)
	// 	return MessageResponse{Text: text}, err
	// }

	// if hasCommand(msg, "/leaderboard") {
	// 	text, err := uc.leaderboardUC.Execute(ctx)
	// 	return MessageResponse{Text: text}, err
	// }

	// if hasCommand(msg, "/mystats") {
	// 	text, err := uc.myStatsUC.Execute(ctx, userID, name)
	// 	return MessageResponse{Text: text}, err
	// }

	// if hasCommand(msg, "/achievements") {
	// 	text, err := uc.achievementsUC.Execute(ctx)
	// 	return MessageResponse{Text: text}, err
	// }

	// if hasCommand(msg, "/comeback") {
	// 	text, err := uc.comebackUC.Execute(ctx, userID, name)
	// 	return MessageResponse{Text: text}, err
	// }

	// if hasCommand(msg, "/strava") {
	// 	authURL := uc.linkStravaUC.GetAuthURL(userID, name)
	// 	text := fmt.Sprintf("🚴‍♂️ *Integrasi Strava* 🏃‍♂️\n\nKlik link di bawah ini untuk menghubungkan akun Strava kamu:\n\n%s\n\nSetelah berhasil, aktivitas larimu akan otomatis dilaporkan! 🎉", authURL)
	// 	return MessageResponse{Text: text, IsPrivate: true}, nil
	// }

	// if strings.HasPrefix(msg, "!broadcast_update") {
	// 	text := uc.broadcastUpdateUC.Execute()
	// 	return MessageResponse{Text: text}, nil
	// }

	return MessageResponse{Text: unknownCommandMessage}, nil
}

func extractSideQuestReport(message string) (string, bool) {
	lower := strings.ToLower(message)
	for _, command := range []string{"/lapor sidequest", "/lapor-sidequest"} {
		if !hasCommand(lower, command) {
			continue
		}
		return strings.TrimSpace(message[len(command):]), true
	}
	return "", false
}

func hasCommand(message, command string) bool {
	if !strings.HasPrefix(message, command) {
		return false
	}
	if len(message) == len(command) {
		return true
	}
	return unicode.IsSpace(rune(message[len(command)]))
}
