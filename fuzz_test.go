//go:build go1.19
// +build go1.19

package semver

import "testing"

func FuzzReverse(f *testing.F) {
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
