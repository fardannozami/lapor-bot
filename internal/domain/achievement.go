package domain

import "strings"

// Achievement represents a gamification achievement that users can unlock.
type Achievement struct {
	ID            string
	Name          string
	Description   string
	Points        int
	DisplayEmoji  string
	UnlockMessage string
	Check         func(report *Report) bool
}

// AllAchievements defines all available achievements in order.
var AllAchievements = []Achievement{
	{
		ID:            "first_report",
		Name:          "Pemula",
		Description:   "Laporan pertama",
		Points:        10,
		DisplayEmoji:  "🐣",
		UnlockMessage: "Selamat datang di perjalanan kebugaranmu! Satu laporan pertama adalah langkah terberat — dan kamu sudah melewatinya. Terus melangkah! 🚀",
		Check:         func(r *Report) bool { return r.ActivityCount >= 1 },
	},
	{
		ID:            "streak_1",
		Name:          "Konsisten",
		Description:   "1 minggu berturut-turut",
		Points:        25,
		DisplayEmoji:  "🔥",
		UnlockMessage: "Konsistensi adalah kunci! Satu minggu berturut-turut membuktikan kamu serius. Api semangatmu sudah menyala — jangan biarkan padam!",
		Check:         func(r *Report) bool { return r.MaxStreak >= 1 },
	},
	{
		ID:            "streak_2",
		Name:          "On Fire",
		Description:   "2 minggu berturut-turut",
		Points:        50,
		DisplayEmoji:  "⚡",
		UnlockMessage: "Dua minggu berturut-turut! Kamu bukan sekadar konsisten, kamu ON FIRE! Energimu menular ke seluruh grup! ⚡🔥",
		Check:         func(r *Report) bool { return r.MaxStreak >= 2 },
	},
	{
		ID:            "streak_3",
		Name:          "Gigih",
		Description:   "3 minggu berturut-turut",
		Points:        75,
		DisplayEmoji:  "💎",
		UnlockMessage: "Tiga minggu tanpa henti! Seperti berlian yang terbentuk dari tekanan, ketekunanmu mulai membentuk sesuatu yang berharga. Kilau mulai terlihat! ✨",
		Check:         func(r *Report) bool { return r.MaxStreak >= 3 },
	},
	{
		ID:            "streak_4",
		Name:          "Spartan",
		Description:   "4 minggu berturut-turut",
		Points:        100,
		DisplayEmoji:  "🛡️",
		UnlockMessage: "THIS IS SPARTA! Sebulan penuh konsisten. Tameng Spartan kini menjadi simbol pertahananmu terhadap rasa malas. Kau prajurit sejati! 🛡️💪",
		Check:         func(r *Report) bool { return r.MaxStreak >= 4 },
	},
	{
		ID:            "streak_8",
		Name:          "Titan",
		Description:   "8 minggu berturut-turut",
		Points:        150,
		DisplayEmoji:  "🏛️",
		UnlockMessage: "Kamu bukan manusia biasa — kamu TITAN! Dua bulan konsisten adalah pencapaian yang hanya diraih oleh mereka yang punya mental juara. Berdiri tegak di puncak Olympus-mu! ⚡",
		Check:         func(r *Report) bool { return r.MaxStreak >= 8 },
	},
	{
		ID:            "streak_12",
		Name:          "Centurion",
		Description:   "12 minggu berturut-turut",
		Points:        300,
		DisplayEmoji:  "⚔️",
		UnlockMessage: "CENTURION! Romawi kuno memberi gelar ini hanya untuk prajurit terbaik yang memimpin 100 orang. Kamu sudah membuktikan kepemimpinan melalui aksi, bukan kata-kata! ⚔️👑",
		Check:         func(r *Report) bool { return r.MaxStreak >= 12 },
	},
	{
		ID:            "activity_10",
		Name:          "10 Hari",
		Description:   "Total 10 hari aktif",
		Points:        20,
		DisplayEmoji:  "🌟",
		UnlockMessage: "10 hari aktif! Bintang kecilmu mulai bersinar. Perjalanan seribu mil dimulai dari langkah pertama — dan kamu sudah 10 langkah! 🌟",
		Check:         func(r *Report) bool { return r.ActivityCount >= 10 },
	},
	{
		ID:            "activity_25",
		Name:          "25 Hari",
		Description:   "Total 25 hari aktif",
		Points:        50,
		DisplayEmoji:  "⭐",
		UnlockMessage: "25 hari bergerak! Bintangmu semakin terang. Ini bukan lagi coba-coba — ini sudah menjadi gaya hidup! ⭐💪",
		Check:         func(r *Report) bool { return r.ActivityCount >= 25 },
	},
	{
		ID:            "activity_50",
		Name:          "Half Century",
		Description:   "Total 50 hari aktif",
		Points:        100,
		DisplayEmoji:  "🏅",
		UnlockMessage: "HALF CENTURY! 50 hari berkeringat. Kamu bukan pemula lagi — kamu atlet sejati yang layak dapat medali! 🏅🎖️",
		Check:         func(r *Report) bool { return r.ActivityCount >= 50 },
	},
	{
		ID:            "activity_100",
		Name:          "Century",
		Description:   "Total 100 hari aktif",
		Points:        200,
		DisplayEmoji:  "💯",
		UnlockMessage: "CENTURY! 100 HARI! 💯 Kamu adalah living proof bahwa komitmen mengalahkan motivasi sesaat. Hari ini kamu bukan cuma dapat badge — kamu dapat gelar LEGEND! 🏆",
		Check:         func(r *Report) bool { return r.ActivityCount >= 100 },
	},
	// --- Extended Streak Achievements ---
	{
		ID:            "streak_5",
		Name:          "Iron Will",
		Description:   "5 minggu berturut-turut",
		Points:        125,
		DisplayEmoji:  "🦾",
		UnlockMessage: "Tekadmu sekuat baja! Lima minggu membuktikan bahwa olahraga sudah menjadi bagian dari DNA-mu. Tak ada yang bisa menghentikanmu sekarang! 🦾🔥",
		Check:         func(r *Report) bool { return r.MaxStreak >= 5 },
	},
	{
		ID:            "streak_10",
		Name:          "Unstoppable",
		Description:   "10 minggu berturut-turut",
		Points:        200,
		DisplayEmoji:  "🚂",
		UnlockMessage: "UNSTOPPABLE! Seperti kereta yang terus melaju, tak ada yang bisa menghentikan momentummu. Sepuluh minggu — kamu inspirasi bagi seluruh grup! 🚂💨",
		Check:         func(r *Report) bool { return r.MaxStreak >= 10 },
	},
	{
		ID:            "streak_16",
		Name:          "Season Conqueror",
		Description:   "16 minggu berturut-turut (full season!)",
		Points:        400,
		DisplayEmoji:  "👑",
		UnlockMessage: "SEASON CONQUEROR! Kamu menaklukkan seluruh season tanpa jeda. Mahkota ini bukan diberikan — kamu merebutnya dengan keringat dan disiplin. LEGEND! 👑🔥",
		Check:         func(r *Report) bool { return r.MaxStreak >= 16 },
	},
	// --- Seasonal Activity Achievements ---
	{
		ID:            "season_active_7",
		Name:          "Season Starter",
		Description:   "7 hari aktif di season ini",
		Points:        25,
		DisplayEmoji:  "🌅",
		UnlockMessage: "7 hari di season ini! Matahari terbit menandai awal yang cerah. Kamu sudah membangun momentum — teruskan! 🌅💪",
		Check:         func(r *Report) bool { return r.SeasonalActivityCount >= 7 },
	},
	{
		ID:            "season_active_25",
		Name:          "Season Grinder",
		Description:   "25 hari aktif di season ini",
		Points:        60,
		DisplayEmoji:  "⚙️",
		UnlockMessage: "25 hari di season ini! Grinder sejati tidak pernah berhenti berputar. Roda gigi disiplinmu terus menghasilkan progres! ⚙️🔥",
		Check:         func(r *Report) bool { return r.SeasonalActivityCount >= 25 },
	},
	// --- Seasonal Point Achievements ---
	{
		ID:            "season_hunter",
		Name:          "Season Hunter",
		Description:   "Raih 300+ poin dalam 1 season",
		Points:        50,
		DisplayEmoji:  "🏹",
		UnlockMessage: "300 poin di season ini! Seperti pemburu yang sabar, kamu mengincar target demi target. Tepat sasaran! 🏹🎯",
		Check:         func(r *Report) bool { return r.SeasonalPoints >= 300 },
	},
	{
		ID:            "season_master",
		Name:          "Season Master",
		Description:   "Raih 500+ poin dalam 1 season",
		Points:        100,
		DisplayEmoji:  "🧙",
		UnlockMessage: "500 poin di season ini! Kamu menguasai seni konsistensi. Seperti penyihir yang meracik ramuan, kamu tahu persis formula sukses: kerja keras + konsistensi = hasil maksimal! 🧙✨",
		Check:         func(r *Report) bool { return r.SeasonalPoints >= 500 },
	},
}

// AllSeasonAchievements defines resettable season badges. Lifetime level/EXP is
// still preserved, but these badge unlock states reset at the season boundary.
var AllSeasonAchievements = []Achievement{
	{
		ID:            "first_report",
		Name:          "Awakened Hunter",
		Description:   "Laporan pertama di season ini",
		Points:        10,
		DisplayEmoji:  "🐣",
		UnlockMessage: "Awakening dimulai! Satu laporan pertama membuka gerbang dungeon season ini. Terus naikkan rank-mu! 🚪⚔️",
		Check:         func(r *Report) bool { return r.SeasonalActivityCount >= 1 },
	},
	{
		ID:            "streak_1",
		Name:          "E-Rank Consistency",
		Description:   "1 minggu streak di season ini",
		Points:        25,
		DisplayEmoji:  "🔥",
		UnlockMessage: "Quest mingguan pertama selesai. Api disiplinmu sudah menyala — jangan biarkan padam!",
		Check:         func(r *Report) bool { return r.SeasonalMaxStreak >= 1 },
	},
	{
		ID:            "streak_2",
		Name:          "D-Rank Momentum",
		Description:   "2 minggu streak di season ini",
		Points:        50,
		DisplayEmoji:  "⚡",
		UnlockMessage: "Momentum terbentuk! Dua minggu berturut-turut membuktikan kamu bukan hunter biasa. ⚡",
		Check:         func(r *Report) bool { return r.SeasonalMaxStreak >= 2 },
	},
	{
		ID:            "streak_3",
		Name:          "C-Rank Grinder",
		Description:   "3 minggu streak di season ini",
		Points:        75,
		DisplayEmoji:  "💎",
		UnlockMessage: "Tiga minggu grinding! Stat disiplinmu naik drastis. Dungeon rasa malas mulai terasa kecil. 💎",
		Check:         func(r *Report) bool { return r.SeasonalMaxStreak >= 3 },
	},
	{
		ID:            "streak_4",
		Name:          "B-Rank Vanguard",
		Description:   "4 minggu streak di season ini",
		Points:        100,
		DisplayEmoji:  "🛡️",
		UnlockMessage: "Sebulan penuh bertahan di garis depan. Kamu layak membawa tameng B-Rank Vanguard! 🛡️",
		Check:         func(r *Report) bool { return r.SeasonalMaxStreak >= 4 },
	},
	{
		ID:            "streak_8",
		Name:          "A-Rank Hunter",
		Description:   "8 minggu streak di season ini",
		Points:        150,
		DisplayEmoji:  "🏛️",
		UnlockMessage: "Delapan minggu tanpa menyerah. Aura A-Rank mulai terasa — kamu jadi standar baru di grup! 🏛️",
		Check:         func(r *Report) bool { return r.SeasonalMaxStreak >= 8 },
	},
	{
		ID:            "streak_12",
		Name:          "S-Rank Hunter",
		Description:   "12 minggu streak di season ini",
		Points:        300,
		DisplayEmoji:  "⚔️",
		UnlockMessage: "S-RANK! Dua belas minggu menaklukkan quest. Ini bukan motivasi sesaat — ini sistem hidup. ⚔️👑",
		Check:         func(r *Report) bool { return r.SeasonalMaxStreak >= 12 },
	},
	{
		ID:            "activity_10",
		Name:          "Daily Quest x10",
		Description:   "10 hari aktif di season ini",
		Points:        20,
		DisplayEmoji:  "🌟",
		UnlockMessage: "10 daily quest selesai! Hal kecil yang diulang mulai jadi kekuatan besar. 🌟",
		Check:         func(r *Report) bool { return r.SeasonalActivityCount >= 10 },
	},
	{
		ID:            "activity_25",
		Name:          "Dungeon Grinder",
		Description:   "25 hari aktif di season ini",
		Points:        50,
		DisplayEmoji:  "⭐",
		UnlockMessage: "25 hari aktif! Kamu terus masuk dungeon meski tidak selalu mudah. Respect, hunter. ⭐",
		Check:         func(r *Report) bool { return r.SeasonalActivityCount >= 25 },
	},
	{
		ID:            "activity_50",
		Name:          "Raid Captain",
		Description:   "50 hari aktif di season ini",
		Points:        100,
		DisplayEmoji:  "🏅",
		UnlockMessage: "50 hari aktif dalam satu season. Kamu bukan cuma ikut raid — kamu memimpin ritmenya! 🏅",
		Check:         func(r *Report) bool { return r.SeasonalActivityCount >= 50 },
	},
	{
		ID:            "activity_100",
		Name:          "Season Monarch",
		Description:   "100 hari aktif di season ini",
		Points:        200,
		DisplayEmoji:  "💯",
		UnlockMessage: "MONARCH OF THE SEASON! 100 hari aktif adalah bukti bahwa sistemmu sudah melampaui mood. 💯👑",
		Check:         func(r *Report) bool { return r.SeasonalActivityCount >= 100 },
	},
	{
		ID:            "streak_16",
		Name:          "Season Conqueror",
		Description:   "16 minggu streak di season ini",
		Points:        400,
		DisplayEmoji:  "👑",
		UnlockMessage: "SEASON CONQUEROR! Kamu menaklukkan seluruh season dengan konsistensi. LEGEND! 👑🔥",
		Check:         func(r *Report) bool { return r.SeasonalMaxStreak >= 16 },
	},
	{
		ID:            "season_hunter",
		Name:          "Point Hunter",
		Description:   "Raih 300+ poin dalam season ini",
		Points:        50,
		DisplayEmoji:  "🏹",
		UnlockMessage: "300 poin season! Kamu memburu progress seperti hunter yang tahu targetnya. 🏹🎯",
		Check:         func(r *Report) bool { return r.SeasonalPoints >= 300 },
	},
	{
		ID:            "season_master",
		Name:          "Shadow Monarch",
		Description:   "Raih 500+ poin dalam season ini",
		Points:        100,
		DisplayEmoji:  "🌑",
		UnlockMessage: "500 poin season! Bayangan alasan sudah kamu taklukkan. Rise. 🌑",
		Check:         func(r *Report) bool { return r.SeasonalPoints >= 500 },
	},
}

// ComebackAchievement represents achievements earned by returning after inactivity.
// These are checked separately because they need InactiveDays context.
type ComebackAchievement struct {
	ID                string
	Name              string
	Description       string
	Points            int
	DisplayEmoji      string
	UnlockMessage     string
	MinInactiveDays   int
	MinComebackStreak int
}

// BadgeSummary is a compact display model for recent badge notifications.
type BadgeSummary struct {
	ID           string
	Name         string
	DisplayEmoji string
}

// AllComebackAchievements defines achievements for users who return after inactivity.
var AllComebackAchievements = []ComebackAchievement{
	{
		ID:                "comeback_4",
		Name:              "Comeback Kid",
		Description:       "Kembali dan raih 4 minggu streak setelah absen lama",
		Points:            30,
		DisplayEmoji:      "🔄",
		UnlockMessage:     "COMEBACK KID! Jatuh itu biasa — bangkit itu luar biasa. Kamu membuktikan bahwa masa lalu tidak menentukan masa depan. Selamat kembali, champ! 🔄💪",
		MinInactiveDays:   7,
		MinComebackStreak: 4,
	},
	{
		ID:                "comeback_hero",
		Name:              "Comeback Hero",
		Description:       "Kembali dan raih 8 minggu streak setelah absen lama",
		Points:            75,
		DisplayEmoji:      "🦸",
		UnlockMessage:     "COMEBACK HERO! Dari absen menjadi pahlawan. Transformasimu luar biasa — kamu adalah superhero bagi dirimu sendiri! 🦸✨",
		MinInactiveDays:   14,
		MinComebackStreak: 8,
	},
	{
		ID:                "phoenix",
		Name:              "Phoenix",
		Description:       "Kembali dan raih 12 minggu streak setelah absen lama",
		Points:            150,
		DisplayEmoji:      "🐦‍🔥",
		UnlockMessage:     "PHOENIX! Kamu terbakar, menjadi abu, dan bangkit kembali lebih kuat! Seperti burung legendaris, kamu adalah simbol harapan dan transformasi. RESPECT! 🔥🐦‍🔥",
		MinInactiveDays:   30,
		MinComebackStreak: 12,
	},
}

// CheckComebackAchievements evaluates comeback achievements against the report.
func CheckComebackAchievements(report *Report) []ComebackAchievement {
	var newlyUnlocked []ComebackAchievement
	for _, a := range AllComebackAchievements {
		if !HasAchievement(report.Achievements, a.ID) &&
			report.InactiveDays >= a.MinInactiveDays &&
			report.ComebackStreak >= a.MinComebackStreak {
			newlyUnlocked = append(newlyUnlocked, a)
		}
	}
	return newlyUnlocked
}

// HasAchievement checks if a report's achievements string contains the given achievement ID.
func HasAchievement(achievements string, id string) bool {
	if achievements == "" {
		return false
	}
	for _, a := range strings.Split(achievements, ",") {
		if strings.TrimSpace(a) == id {
			return true
		}
	}
	return false
}

// AddAchievement appends an achievement ID to the achievements string.
func AddAchievement(achievements string, id string) string {
	if achievements == "" {
		return id
	}
	return achievements + "," + id
}

// FindBadgeSummary returns compact badge display data by ID.
func FindBadgeSummary(id string) (BadgeSummary, bool) {
	for _, a := range AllSeasonAchievements {
		if a.ID == id {
			return BadgeSummary{ID: a.ID, Name: a.Name, DisplayEmoji: a.DisplayEmoji}, true
		}
	}
	for _, a := range AllComebackAchievements {
		if a.ID == id {
			return BadgeSummary{ID: a.ID, Name: a.Name, DisplayEmoji: a.DisplayEmoji}, true
		}
	}
	for _, a := range AllAchievements {
		if a.ID == id {
			return BadgeSummary{ID: a.ID, Name: a.Name, DisplayEmoji: a.DisplayEmoji}, true
		}
	}
	return BadgeSummary{}, false
}

// RecentAchievementSummaries returns the latest achieved badges in newest-first order.
func RecentAchievementSummaries(achievements string, limit int) []BadgeSummary {
	if achievements == "" || limit <= 0 {
		return nil
	}

	ids := strings.Split(achievements, ",")
	summaries := make([]BadgeSummary, 0, limit)
	seen := make(map[string]bool, limit)
	for i := len(ids) - 1; i >= 0 && len(summaries) < limit; i-- {
		id := strings.TrimSpace(ids[i])
		if id == "" || seen[id] {
			continue
		}
		if summary, ok := FindBadgeSummary(id); ok {
			summaries = append(summaries, summary)
			seen[id] = true
		}
	}
	return summaries
}

// CheckNewAchievements evaluates all achievements against the report and returns newly unlocked ones.
func CheckNewAchievements(report *Report) []Achievement {
	var newlyUnlocked []Achievement
	for _, a := range AllAchievements {
		if !HasAchievement(report.Achievements, a.ID) && a.Check(report) {
			newlyUnlocked = append(newlyUnlocked, a)
		}
	}
	return newlyUnlocked
}

// CheckNewSeasonAchievements evaluates resettable season badges.
func CheckNewSeasonAchievements(report *Report) []Achievement {
	var newlyUnlocked []Achievement
	for _, a := range AllSeasonAchievements {
		if !HasAchievement(report.SeasonalAchievements, a.ID) && a.Check(report) {
			newlyUnlocked = append(newlyUnlocked, a)
		}
	}
	return newlyUnlocked
}
