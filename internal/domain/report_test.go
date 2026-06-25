package domain

import (
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
