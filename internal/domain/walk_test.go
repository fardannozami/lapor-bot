package domain

import "testing"

func TestParseWalkReport(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantSteps int
		wantOK    bool
	}{
		{name: "plain steps", text: "jalan kaki 8000", wantSteps: 8000, wantOK: true},
		{name: "dot separator", text: "/lapor jalan kaki 12.500", wantSteps: 12500, wantOK: true},
		{name: "comma separator", text: "jalan kaki 12,500 langkah", wantSteps: 12500, wantOK: true},
		{name: "k suffix", text: "walk 10k", wantSteps: 10000, wantOK: true},
		{name: "rb suffix", text: "jalan 8rb", wantSteps: 8000, wantOK: true},
		{name: "ribu suffix", text: "jalan pagi 8ribu langkah", wantSteps: 8000, wantOK: true},
		{name: "with strava link", text: "jalan kaki 12.500 https://www.strava.com/activities/123456789", wantSteps: 12500, wantOK: true},
		{name: "link before number", text: "https://strava.com/activities/98765 jalan sore 7500", wantSteps: 7500, wantOK: true},
		{name: "english walking", text: "Walking 6500 steps", wantSteps: 6500, wantOK: true},
		{name: "langkah keyword only", text: "langkah hari ini 4200", wantSteps: 4200, wantOK: true},
		{name: "no number falls back", text: "jalan pagi santai", wantSteps: 0, wantOK: false},
		// "30" is below MinWalkSteps: falls through to the regular /lapor flow.
		{name: "minutes not steps", text: "jalan 30 menit", wantSteps: 0, wantOK: false},
		{name: "below threshold ignored", text: "jalan kaki 99", wantSteps: 0, wantOK: false},
		{name: "absurd input rejected", text: "jalan kaki 999999999", wantSteps: 0, wantOK: false},
		{name: "no walk keyword", text: "lari 10 km", wantSteps: 0, wantOK: false},
		{name: "glued suffix token", text: "jalan kaki 12.500langkah", wantSteps: 12500, wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps, ok := ParseWalkReport(tt.text)
			if ok != tt.wantOK {
				t.Fatalf("ParseWalkReport(%q) ok=%v, want %v", tt.text, ok, tt.wantOK)
			}
			if steps != tt.wantSteps {
				t.Fatalf("ParseWalkReport(%q) steps=%d, want %d", tt.text, steps, tt.wantSteps)
			}
		})
	}
}

func TestFormatStepsID(t *testing.T) {
	tests := []struct {
		steps int
		want  string
	}{
		{steps: 0, want: "0"},
		{steps: 42, want: "42"},
		{steps: 999, want: "999"},
		{steps: 1000, want: "1.000"},
		{steps: 12500, want: "12.500"},
		{steps: 1234567, want: "1.234.567"},
	}
	for _, tt := range tests {
		if got := FormatStepsID(tt.steps); got != tt.want {
			t.Errorf("FormatStepsID(%d) = %q, want %q", tt.steps, got, tt.want)
		}
	}
}
