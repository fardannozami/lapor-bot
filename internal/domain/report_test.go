package domain

import (
	"fmt"
	"reflect"
	"testing"
)

func TestDetermineAttributes(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []AttributeType
	}{
		{"stamina running", "lari 5km pagi", []AttributeType{AttrSta}},
		{"stamina indonesian verb", "berlari 5km lalu berenang", []AttributeType{AttrSta}},
		{"strength gym", "gym push-up squat", []AttributeType{AttrStr}},
		{"strength push day", "Push Day chest dan triceps", []AttributeType{AttrStr}},
		{"agility badminton", "main bulu tangkis dan padel", []AttributeType{AttrAgi}},
		{"vitality recovery", "yoga stretching mobility", []AttributeType{AttrVit}},
		{"hiking is stamina", "hiking jalan jauh", []AttributeType{AttrSta}},
		{"multi attr stable order", "gym lalu lari dan stretching", []AttributeType{AttrStr, AttrSta, AttrVit}},
		{"legacy does not match leg", "belajar legacy code", nil},
		{"rundown does not match run", "rundown meeting pagi", nil},
		{"push notification does not match push up", "push notification error", nil},
		{"fallback", "aktivitas sehat", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetermineAttributes(tt.text); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("DetermineAttributes(%q) = %#v, want %#v", tt.text, got, tt.want)
			}
		})
	}
}

func TestSelectReportAttribute(t *testing.T) {
	tests := []struct {
		name     string
		attrs    []AttributeType
		jobClass string
		seed     string
		want     AttributeType
	}{
		{"empty returns empty", nil, "fighter", "", ""},
		{"single activity match is kept", []AttributeType{AttrSta}, "ranger", "", AttrSta},
		// Activity directs the reward: a run (STA) still boosts STA even for a
		// STR-focused fighter — the job does not override a clear activity match.
		{"activity overrides job when only one matches", []AttributeType{AttrSta}, "fighter", "", AttrSta},
		// Job breaks ties when several attributes match: the hunter's specialty
		// wins among the matches.
		{"job wins among multiple matches (fighter→STR)", []AttributeType{AttrSta, AttrStr, AttrVit}, "fighter", "", AttrStr},
		{"job wins among multiple matches (ranger→STA)", []AttributeType{AttrStr, AttrSta, AttrVit}, "ranger", "", AttrSta},
		{"job wins among multiple matches (healer→VIT)", []AttributeType{AttrSta, AttrAgi, AttrVit}, "healer", "", AttrVit},
		// When the job's specialty is NOT among the matches, the first matched
		// attribute wins (deterministic, no arbitrary preference).
		{"job not among matches falls back to first", []AttributeType{AttrSta, AttrVit}, "fighter", "", AttrSta},
		// Mage has no single primary → the seed distributes the gain across the
		// matched attributes instead of always defaulting to the first (STR).
		// The same seed must always yield the same attribute (deterministic).
		{"mage multi-match is seed-determined", []AttributeType{AttrStr, AttrSta, AttrVit}, "mage", "userA|regular|2026-06-30|1|gym lari", AttrStr},
		{"unknown job multi-match is seed-determined", []AttributeType{AttrAgi, AttrVit}, "", "u|regular|d|1|x", AttrVit},
		{"single match unaffected by seed", []AttributeType{AttrAgi}, "mage", "any", AttrAgi},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectReportAttribute(tt.attrs, tt.jobClass, tt.seed); got != tt.want {
				t.Fatalf("SelectReportAttribute(%#v, %q, %q) = %q, want %q", tt.attrs, tt.jobClass, tt.seed, got, tt.want)
			}
		})
	}
}

// TestSelectReportAttribute_GrantsSinglePointPerReport is the fairness guard:
// a mixed session that matches three categories must yield exactly ONE
// attribute point (the job's specialty among the matches), not three. This
// prevents the old behavior where "gym lalu lari dan stretching" gave +3 while
// a focused session gave +1.
func TestSelectReportAttribute_GrantsSinglePointPerReport(t *testing.T) {
	mixed := DetermineAttributes("gym lalu lari dan stretching")
	if len(mixed) < 2 {
		t.Fatalf("sanity: expected mixed session to match several attributes, got %#v", mixed)
	}
	chosen := SelectReportAttribute(mixed, "fighter", "seed-irrelevant-when-primary-matches")
	// Exactly one attribute is selected (the fighter specialty STR is among
	// the matches here), so the report grants +1 STR — not +1 to each.
	if chosen != AttrStr {
		t.Fatalf("fighter mixed session should reward its specialty STR, got %q", chosen)
	}
}

// TestSelectReportAttribute_MageDistributesFairly verifies that Mage (no single
// primary attribute) does not bias toward one attribute: across many distinct
// report seeds, all candidate attributes get selected. This guards against the
// old always-first behavior that funneled every Mage gain into STR.
func TestSelectReportAttribute_MageDistributesFairly(t *testing.T) {
	attrs := []AttributeType{AttrStr, AttrSta, AttrAgi, AttrVit}
	seen := make(map[AttributeType]bool)
	const samples = 400
	for i := 0; i < samples; i++ {
		seed := fmt.Sprintf("mage-user|regular|2026-06-30|%d|mixed activity", i)
		chosen := SelectReportAttribute(attrs, "mage", seed)
		if chosen == "" {
			t.Fatalf("seed %d returned empty attribute", i)
		}
		seen[chosen] = true
	}
	if len(seen) != 4 {
		t.Fatalf("Mage expected to distribute across all 4 attributes over %d seeds, only got %v", samples, seen)
	}
}
