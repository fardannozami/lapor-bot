package usecase

import (
	"context"
	"strings"
)

type HandleMessageUsecase struct {
	reportUC    *ReportActivityUsecase
	leaderboardUC  *GetLeaderboardUsecase
	myStatsUC      *GetMyStatsUsecase
	achievementsUC *GetAchievementsUsecase
	comebackUC     *ComebackChallengeUsecase
	updateNameUC   *UpdateNameUsecase
}

func NewHandleMessageUsecase(
	reportUC *ReportActivityUsecase,
	leaderboardUC *GetLeaderboardUsecase,
	myStatsUC *GetMyStatsUsecase,
	achievementsUC *GetAchievementsUsecase,
	comebackUC *ComebackChallengeUsecase,
	updateNameUC *UpdateNameUsecase,
) *HandleMessageUsecase {
	return &HandleMessageUsecase{
		reportUC:       reportUC,
		leaderboardUC:  leaderboardUC,
		myStatsUC:      myStatsUC,
		achievementsUC: achievementsUC,
		comebackUC:     comebackUC,
		updateNameUC:   updateNameUC,
	}
}

func (uc *HandleMessageUsecase) Execute(ctx context.Context, userID, name, message string) (string, error) {
	msg := strings.ToLower(strings.TrimSpace(message))

	// Handle #lapor (di mana aja posisinya)
	if strings.Contains(msg, "#lapor") {
		return uc.reportUC.Execute(ctx, userID, name)
	}

	// Handle #setname
	if strings.Contains(msg, "#setname") {
		// Extract name from message: everything after "#setname"
		idx := strings.Index(msg, "#setname")
		newName := strings.TrimSpace(message[idx+len("#setname"):])
		return uc.updateNameUC.Execute(ctx, userID, newName)
	}

	// Handle #leaderboard (di mana aja juga)
	if strings.Contains(msg, "#leaderboard") {
		return uc.leaderboardUC.Execute(ctx)
	}

	// Handle #mystats
	if strings.Contains(msg, "#mystats") {
		return uc.myStatsUC.Execute(ctx, userID, name)
	}

	// Handle #achievements
	if strings.Contains(msg, "#achievements") {
		return uc.achievementsUC.Execute(ctx)
	}

	// Handle #comeback
	if strings.Contains(msg, "#comeback") {
		return uc.comebackUC.Execute(ctx, userID, name)
	}

	return "", nil
}
