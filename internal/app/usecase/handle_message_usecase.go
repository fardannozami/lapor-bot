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
	msg := strings.TrimSpace(message)

	// Handle #lapor
	if strings.HasPrefix(strings.ToLower(msg), "#lapor") {
		return uc.reportUC.Execute(ctx, userID, name)
	}

	// Handle #leaderboard
	if strings.HasPrefix(strings.ToLower(msg), "#leaderboard") {
		return uc.leaderboardUC.Execute(ctx)
	}

	return "", nil
}
