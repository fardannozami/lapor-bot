package usecase

import (
	"context"
	"fmt"
	"strings"

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
	}
}

func (uc *HandleMessageUsecase) Execute(ctx context.Context, userID, name, message string) (MessageResponse, error) {
	msg := strings.ToLower(strings.TrimSpace(message))

	if strings.Contains(msg, "#help") {
		text := uc.helpUC.Execute()
		return MessageResponse{Text: text}, nil
	}

	if strings.Contains(msg, "#lapor") {
		workout := domain.ParseHevy(message)
		text, err := uc.reportUC.Execute(ctx, userID, name, workout)
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#cancel") {
		text, err := uc.cancelUC.Execute(ctx, userID, name)
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#motivasi") {
		text := uc.motivationUC.Execute()
		return MessageResponse{Text: text}, nil
	}

	if strings.Contains(msg, "#setname") {
		idx := strings.Index(msg, "#setname")
		newName := strings.TrimSpace(message[idx+len("#setname"):])
		text, err := uc.updateNameUC.Execute(ctx, userID, newName)
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#leaderboard-weekly") {
		text, err := uc.weeklyLeaderboardUC.Execute(ctx)
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#leaderboard-seasonal") {
		text, err := uc.leaderboardUC.ExecuteSeasonal(ctx)
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#leaderboard") {
		text, err := uc.leaderboardUC.Execute(ctx)
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#mystats") {
		text, err := uc.myStatsUC.Execute(ctx, userID, name)
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#achievements") {
		text, err := uc.achievementsUC.Execute(ctx)
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#comeback") {
		text, err := uc.comebackUC.Execute(ctx, userID, name)
		return MessageResponse{Text: text}, err
	}

	if strings.Contains(msg, "#strava") {
		authURL := uc.linkStravaUC.GetAuthURL(userID, name)
		text := fmt.Sprintf("🚴‍♂️ *Integrasi Strava* 🏃‍♂️\n\nKlik link di bawah ini untuk menghubungkan akun Strava kamu:\n\n%s\n\nSetelah berhasil, aktivitas larimu akan otomatis dilaporkan! 🎉", authURL)
		return MessageResponse{Text: text, IsPrivate: true}, nil
	}

	if strings.HasPrefix(msg, "!broadcast_update") {
		text := uc.broadcastUpdateUC.Execute()
		return MessageResponse{Text: text}, nil
	}

	return MessageResponse{}, nil
}
