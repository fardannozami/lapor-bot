package usecase

import (
	"fmt"
	"testing"
	"time"

	"github.com/fardannozami/whatsapp-gateway/internal/domain"
)

func TestBuildReminderMessage(t *testing.T) {
	now := time.Now()

	// Simulate the same users from the real announcement
	testUsers := []struct {
		id          string
		name        string
		maxStreak   int
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
		{"6285769476484", "M Idrus Salam", 1, 28},  // actually mild (< 30)
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

	// Bucket into tiers
	var critical, warning, mild []inactiveUserInfo
	for _, u := range users {
		switch {
		case u.daysInactive >= tierCriticalDays:
			critical = append(critical, u)
		case u.daysInactive >= tierWarningDays:
			warning = append(warning, u)
		default:
			mild = append(mild, u)
		}
	}

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

	msg, mentions := BuildReminderMessage(users, critical, warning, mild, hallOfFame, comebackCounts)

	// Print the actual message so you can see it
	fmt.Println("=== GENERATED MESSAGE ===")
	fmt.Println(msg)
	fmt.Println("=== END MESSAGE ===")
	fmt.Printf("\nTotal mentions: %d\n", len(mentions))
	fmt.Printf("Critical: %d, Warning: %d, Mild: %d, Hall of Fame: %d\n",
		len(critical), len(warning), len(mild), len(hallOfFame))

	// Basic assertions
	if len(mentions) == 0 {
		t.Error("expected mentions to be non-empty")
	}
	if msg == "" {
		t.Error("expected message to be non-empty")
	}
	if len(critical) == 0 {
		t.Error("expected critical tier to have users")
	}
	if len(hallOfFame) == 0 {
		t.Error("expected hall of fame to have users")
	}
}
