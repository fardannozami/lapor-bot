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
	reportUC    *ReportActivityUsecase
	leaderboardUC  *GetLeaderboardUsecase
	myStatsUC      *GetMyStatsUsecase
	achievementsUC *GetAchievementsUsecase
	comebackUC     *ComebackChallengeUsecase
	updateNameUC   *UpdateNameUsecase
	linkStravaUC   *LinkStravaUsecase
}

func NewHandleMessageUsecase(
	reportUC *ReportActivityUsecase,
	leaderboardUC *GetLeaderboardUsecase,
	myStatsUC *GetMyStatsUsecase,
	achievementsUC *GetAchievementsUsecase,
	comebackUC *ComebackChallengeUsecase,
	updateNameUC *UpdateNameUsecase,
	linkStravaUC *LinkStravaUsecase,
) *HandleMessageUsecase {
	return &HandleMessageUsecase{
		reportUC:       reportUC,
		leaderboardUC:  leaderboardUC,
		myStatsUC:      myStatsUC,
		achievementsUC: achievementsUC,
		comebackUC:     comebackUC,
		updateNameUC:   updateNameUC,
		linkStravaUC:   linkStravaUC,
	}
}

func (uc *HandleMessageUsecase) Execute(ctx context.Context, userID, name, message string) (MessageResponse, error) {
	msg := strings.ToLower(strings.TrimSpace(message))

	// Handle #lapor (di mana aja posisinya)
	if strings.Contains(msg, "#lapor") {
		workout := domain.ParseHevy(message)
		text, err := uc.reportUC.Execute(ctx, userID, name, workout)
		return MessageResponse{Text: text}, err
	}

	// Handle #setname
	if strings.Contains(msg, "#setname") {
		// Extract name from message: everything after "#setname"
		idx := strings.Index(msg, "#setname")
		newName := strings.TrimSpace(message[idx+len("#setname"):])
		text, err := uc.updateNameUC.Execute(ctx, userID, newName)
		return MessageResponse{Text: text}, err
	}

	// Handle #leaderboard (di mana aja juga)
	if strings.Contains(msg, "#leaderboard") {
		text, err := uc.leaderboardUC.Execute(ctx)
		return MessageResponse{Text: text}, err
	}

	// Handle #mystats
	if strings.Contains(msg, "#mystats") {
		text, err := uc.myStatsUC.Execute(ctx, userID, name)
		return MessageResponse{Text: text}, err
	}

	// Handle #achievements
	if strings.Contains(msg, "#achievements") {
		text, err := uc.achievementsUC.Execute(ctx)
		return MessageResponse{Text: text}, err
	}

	// Handle #comeback
	if strings.Contains(msg, "#comeback") {
		text, err := uc.comebackUC.Execute(ctx, userID, name)
		return MessageResponse{Text: text}, err
	}

	// Handle #strava
	if strings.Contains(msg, "#strava") {
		authURL := uc.linkStravaUC.GetAuthURL(userID, name)
		text := fmt.Sprintf("🚴‍♂️ *Integrasi Strava* 🏃‍♂️\n\nKlik link di bawah ini untuk menghubungkan akun Strava kamu:\n\n%s\n\nSetelah berhasil, aktivitas larimu akan otomatis dilaporkan! 🎉", authURL)
		return MessageResponse{Text: text, IsPrivate: true}, nil
	}

	return MessageResponse{}, nil
}
