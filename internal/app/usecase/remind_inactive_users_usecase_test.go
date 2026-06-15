package usecase

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

func TestBuildReminderMessage(t *testing.T) {
	now := time.Now()

	// Simulate the same users from the real announcement
	testUsers := []struct {
		id           string
		name         string
		maxStreak    int
		daysInactive int
	}{
		// Critical tier (60+ days)
		{"6281289281245", "S.Pratama", 1, 80},
		{"6282340599539", "Deimos", 1, 80},
		{"6283157206433", "Dea", 1, 80},
		{"6281334228595", "Aufa Wibowo", 1, 80},
		{"6287765797036", "Ydh", 1, 80},
		{"6287791194987", "Anwar Abdullah", 1, 80},
		{"628977275056", "Zainal", 1, 80},
		{"6285211486365", "Agus W", 1, 80},
		{"62895352983939", "Day", 1, 80},
		{"6281274429667", "Ahmad Amri", 1, 78},
		{"6285695744459", "Fauzan Azhari", 1, 77},
		{"6281383287559", "enjangsetiawan", 1, 75},
		{"6282231878575", "Novi", 1, 75},
		{"6285795094888", "Finn Christoffer", 1, 75},
		{"6282153481251", "Rakhel", 1, 74},
		{"628119292778", "M Fauzi", 2, 73},
		{"628978574686", "Fadli Asyhari", 1, 70},
		{"6281276307723", "Maulana", 1, 70},
		{"6285156354193", "cecep wandy", 1, 70},
		{"6285171231336", "Naufal Fawwaz", 1, 67},
		{"6289643373386", "Risa", 2, 60},
		// Warning tier (30-59 days)
		{"6281398447822", "yusnar", 1, 59},
		{"6285846132417", "Taupik Pirdian", 1, 58},
		{"6282293119116", "Fadli Rusandy", 4, 54},
		{"6281277753811", "Rezki Nasrullah", 1, 49},
		{"6287878970037", "Iqbal M", 2, 46},
		{"6282121455132", "Asep Ardi", 4, 45},
		{"6285237264054", "Ahmad Sufyan", 4, 44},
		{"6281919101990", "naufal nazaruddin", 5, 44},
		{"6285866473926", "nuribnuu", 3, 43},
		{"6285817110632", "Bayu Santoso", 4, 42},
		{"6285719602421", "Ismail S", 4, 41},
		{"628996423135", "Faqih Nur Fahmi", 6, 37},
		{"6282338588078", "swegrowthid", 3, 35},
		{"6282127070014", "Muhammad Syarifudin", 4, 34},
		{"6287868126244", "Mochammad Rizqi", 4, 33},
		{"6281264808425", "fajar setiawan", 1, 33},
		{"628986389685", "Agus Setyawan", 1, 32},
		{"6285769476484", "M Idrus Salam", 1, 28}, // actually mild (< 30)
		// Mild tier (7-29 days)
		{"6281329262095", "Haddawi", 2, 21},
		{"6285724230162", "fajardsutera", 4, 20},
		{"6282121632543", "aabar", 1, 18},
		{"628115420582", "Kasjful Kurniawan", 1, 16},
		{"6285740482440", "Arif Faizin", 3, 15},
		{"6287875228746", "Otniel", 1, 14},
		{"6289634879461", "fauzandp", 1, 11},
	}

	// Build inactiveUserInfo slice
	var users []inactiveUserInfo
	for _, tu := range testUsers {
		users = append(users, inactiveUserInfo{
			user: &domain.Report{
				UserID:         tu.id,
				Name:           tu.name,
				MaxStreak:      tu.maxStreak,
				LastReportDate: now.AddDate(0, 0, -tu.daysInactive),
				Achievements:   "", // no achievements yet
			},
			daysInactive: tu.daysInactive,
		})
	}

	weeklyGroups := groupInactiveUsersByWeeks(users)

	// Hall of fame
	var hallOfFame []inactiveUserInfo
	for _, u := range users {
		if u.user.MaxStreak >= 4 {
			hallOfFame = append(hallOfFame, u)
		}
	}

	// Comeback counts
	comebackCounts := make(map[string]int)
	for _, u := range users {
		for _, a := range domain.AllComebackAchievements {
			if u.daysInactive >= a.MinInactiveDays && !domain.HasAchievement(u.user.Achievements, a.ID) {
				comebackCounts[a.Name]++
				break
			}
		}
	}

	activeToday := []*domain.Report{
		{UserID: "628111111111", Name: "Active One", LastReportDate: now},
		{UserID: "628222222222", Name: "Active Two", LastReportDate: now},
	}

	msg, mentions := BuildReminderMessage(users, weeklyGroups, hallOfFame, comebackCounts, activeToday)

	// Print the actual message so you can see it
	fmt.Println("=== GENERATED MESSAGE ===")
	fmt.Println(msg)
	fmt.Println("=== END MESSAGE ===")
	fmt.Printf("\nTotal mentions: %d\n", len(mentions))
	fmt.Printf("Weekly groups: %d, Hall of Fame: %d\n", len(weeklyGroups), len(hallOfFame))

	// Basic assertions
	if len(mentions) == 0 {
		t.Error("expected mentions to be non-empty")
	}
	if msg == "" {
		t.Error("expected message to be non-empty")
	}
	if len(weeklyGroups) == 0 {
		t.Error("expected weekly groups to have users")
	}
	for _, want := range []string{"1 minggu belum lapor", "2 minggu belum lapor", "3+ minggu belum lapor"} {
		if !strings.Contains(msg, want) {
			t.Errorf("expected weekly reminder to contain %q, got %q", want, msg)
		}
	}
	if len(hallOfFame) == 0 {
		t.Error("expected hall of fame to have users")
	}
	if !strings.Contains(msg, "Apresiasi buat yang sudah olahraga hari ini") {
		t.Error("expected reminder to appreciate users who already exercised today")
	}
	if !strings.Contains(msg, "Ada *2 orang* yang sudah bergerak") {
		t.Error("expected reminder to appreciate active users collectively")
	}
	if !strings.Contains(msg, "#mysidequest") {
		t.Error("expected 15:15 reminder to include #mysidequest info")
	}
	for _, active := range activeToday {
		if strings.Contains(msg, active.Name) || strings.Contains(msg, active.UserID) {
			t.Errorf("active user %s should be appreciated collectively without name or ID", active.UserID)
		}
		mentionedJID := active.UserID + "@s.whatsapp.net"
		for _, mention := range mentions {
			if mention == mentionedJID {
				t.Errorf("active user %s should not be mentioned", active.UserID)
			}
		}
	}
}
