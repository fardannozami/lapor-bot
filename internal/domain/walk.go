package domain

import (
	"strconv"
	"strings"
	"unicode"
)

const (
	// MinWalkSteps is the smallest plausible step count for a walk report.
	// Values below this (e.g. "jalan 30 menit") are treated as ordinary
	// activity text, not as a steps report.
	MinWalkSteps = 100
	// MaxWalkStepsPerReport rejects typos and absurd inputs; even a 100-mile
	// ultra is roughly 200k steps.
	MaxWalkStepsPerReport = 250000
)

// walkKeywords are matched with the same boundary-aware matcher used for
// activity attributes so "jalan pagi" matches but "jalan123" does not.
var walkKeywords = []string{"jalan kaki", "jalan", "walk", "walking", "langkah", "steps"}

// walkNumberSuffixes expand compact Indonesian step notation ("8rb", "8ribu",
// "8k") to the full count. Longest suffix first so "8ribu" strips "ribu".
var walkNumberSuffixes = []struct {
	suffix     string
	multiplier int
}{
	{"ribu", 1000},
	{"rb", 1000},
	{"k", 1000},
}

// ParseWalkReport extracts a step count from a walk activity text such as
// "/lapor jalan kaki 12.500 https://strava.com/activities/123".
//
// Detection rules:
//   - text must contain a walking keyword (matched on the normalized text,
//     same matcher as activity attributes);
//   - every whitespace token of the RAW text is scanned for a number;
//     thousand separators ("." ",") and rb/ribu/k suffixes are expanded;
//   - the LARGEST candidate wins so digits inside pasted Strava links can
//     never beat the real step count (real counts dwarf URL digits);
//   - result must fall within [MinWalkSteps, MaxWalkStepsPerReport].
//
// Returns steps=0, ok=false when the text is not a steps report; callers
// then fall through to the regular /lapor flow unchanged.
func ParseWalkReport(text string) (int, bool) {
	normalized := normalizeActivityText(text)
	if !containsAnyActivityKeyword(normalized, walkKeywords) {
		return 0, false
	}
	// Number scanning runs on the raw (lowercased) text: normalize collapses
	// "12.500" into "12 500" which would corrupt the count. A thin space is
	// inserted at digit->letter boundaries so "12500langkah" still parses.
	steps := largestStepCandidate(splitDigitLetterBoundaries(strings.ToLower(text)))
	if steps < MinWalkSteps || steps > MaxWalkStepsPerReport {
		return 0, false
	}
	return steps, true
}

// largestStepCandidate returns the largest numeric candidate across all
// whitespace-separated tokens of the raw text. Non-numeric tokens (links,
// words) are skipped.
func largestStepCandidate(raw string) int {
	best := 0
	for _, token := range strings.Fields(raw) {
		if value, ok := parseStepToken(token); ok && value > best {
			best = value
		}
	}
	return best
}

// splitDigitLetterBoundaries inserts a space wherever a digit is followed by
// a letter, EXCEPT when the letters form a known step suffix ("8rb", "8k",
// "8ribu") so compact notation stays one token. Glued unit words like
// "12500langkah" still split into number + word.
func splitDigitLetterBoundaries(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 4)
	for i := 0; i < len(s); i++ {
		r := rune(s[i])
		if i > 0 && unicode.IsDigit(rune(s[i-1])) && unicode.IsLetter(r) {
			if hasStepSuffix(s[i:]) {
				// keep glued: "10k", "8rb", "8ribu"
			} else {
				b.WriteByte(' ')
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

// hasStepSuffix reports whether s starts with a known compact step suffix.
func hasStepSuffix(s string) bool {
	for _, suf := range walkNumberSuffixes {
		if strings.HasPrefix(s, suf.suffix) {
			return true
		}
	}
	return false
}

// parseStepToken parses one token into a step count. Supported shapes:
// "8000", "12.500", "12,500", "8rb", "8ribu", "8k", "12500langkah"
// (trailing punctuation glued to digits is tolerated).
func parseStepToken(token string) (int, bool) {
	for _, suf := range walkNumberSuffixes {
		if strings.HasSuffix(token, suf.suffix) && len(token) > len(suf.suffix) {
			head := strings.TrimSuffix(token, suf.suffix)
			value, err := strconv.Atoi(stripStepSeparators(head))
			if err != nil || value <= 0 {
				continue
			}
			return value * suf.multiplier, true
		}
	}

	cleaned := strings.TrimFunc(stripStepSeparators(token), func(r rune) bool {
		return r == '!' || r == '?' || r == ':' || r == ','
	})
	if cleaned == "" {
		return 0, false
	}
	value, err := strconv.Atoi(cleaned)
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}

// stripStepSeparators removes '.' and ',' thousand separators:
// "12.500" and "12,500" both become "12500".
func stripStepSeparators(token string) string {
	token = strings.ReplaceAll(token, ".", "")
	return strings.ReplaceAll(token, ",", "")
}

// FormatStepsID renders a step count with Indonesian dot separators,
// e.g. 12500 -> "12.500".
func FormatStepsID(steps int) string {
	s := strconv.Itoa(steps)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
	}
	for i := pre; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte('.')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
