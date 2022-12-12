//go:build go1.19
// +build go1.19

package semver

import "testing"

func FuzzParse(f *testing.F) {
	testcase := []string{
		"1.2.3", "1.0", "1.3", "2", "0.4.2",
		"1.2.3-beta.1+build345", ">= 1.2.3",
		"1.2 - 1.4.5",
	}
	for _, t := range testcase {
		f.Add(t)
	}

	f.Fuzz(func(t *testing.T, d string) {
		// Test NewVersion
		_, _ = NewVersion(d)

		// Test StrictNewVersion
		_, _ = StrictNewVersion(d)

		// Test NewConstraint
		_, _ = NewConstraint(d)
	})
}

func FuzzRanges(f *testing.F) {
	testcase := []struct {
		ver        string
		constraint string
	}{
		{"1.2.3", ">= 1"},
		{"0-1", "x-1"},
		{"1.0", "< 2"},
		{"1.3", "~1.3"},
		{"2", "=2"},
		{"0.4.2", "0.1.2 - 1.3.2"},
		{"1.2.3-beta.1+build345", "^1.2.3-beta"},
	}
	for _, t := range testcase {
		f.Add(t.ver, t.constraint)
	}

	f.Fuzz(func(t *testing.T, v string, c string) {
		vv, err := NewVersion(v)
		if err != nil {
			return
		}

		cc, err := NewConstraint(c)
		if err != nil {
			return
		}

		r, err := cc.AsRanges()
		if err != nil {
			return
		}

		if EvalRanges(vv, r) != cc.Check(vv) {
			t.Fatalf("fail to match %v %s %s", r, c, v)
		}
	})
}
