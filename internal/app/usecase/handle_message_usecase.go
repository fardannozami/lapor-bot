package usecase

import (
	"context"
	"strings"
)

type HandleMessageUsecase struct {
	reportUC      *ReportActivityUsecase
	leaderboardUC *GetLeaderboardUsecase
}

func NewHandleMessageUsecase(reportUC *ReportActivityUsecase, leaderboardUC *GetLeaderboardUsecase) *HandleMessageUsecase {
	return &HandleMessageUsecase{
		reportUC:      reportUC,
		leaderboardUC: leaderboardUC,
	}
}

func (uc *HandleMessageUsecase) Execute(ctx context.Context, userID, name, message string) (string, error) {
	msg := strings.ToLower(strings.TrimSpace(message))

	// Handle #lapor (di mana aja posisinya)
	if strings.Contains(msg, "#lapor") {
		return uc.reportUC.Execute(ctx, userID, name)
	}

	// Handle #leaderboard (di mana aja juga)
	if strings.Contains(msg, "#leaderboard") {
		return uc.leaderboardUC.Execute(ctx)
	}

	return "", nil
}
