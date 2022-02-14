package main

import (
	"testing"
)

func TestGuessOverlapName(t *testing.T) {
	testcases := []struct {
		name     string
		s1       string
		s2       string
		expected string
	}{
		{
			name:     "empty",
			s1:       "",
			s2:       "",
			expected: "",
		},
		{
			name:     "shared basename",
			s1:       "a/b/cde",
			s2:       "x/y/cde",
			expected: "cde",
		},
		{
			name:     "common prefix",
			s1:       "myprog-amd64",
			s2:       "myprog-arm64",
			expected: "myprog",
		},
		{
			name:     "common suffix",
			s1:       "amd64-myprog",
			s2:       "arm64-myprog",
			expected: "myprog",
		},
		{
			name:     "suffixed extra components",
			s1:       "./myprog-darwin-amd64",
			s2:       "./myprog-darwin-arm64",
			expected: "myprog",
		},
		{
			name:     "snake-delim suffix",
			s1:       "./myprog_darwin_aarch64",
			s2:       "./myprog_darwin_arm64",
			expected: "myprog",
		},
		{
			name:     "wth",
			s1:       "!?#",
			s2:       "--_",
			expected: "",
		},
		{
			name:     "non-arch suffix",
			s1:       "./myprog-foo-bar_baz",
			s2:       "./myprog-foo-bar_ball",
			expected: "myprog-foo-bar_ba",
		},
		{
			name:     "non-arch prefix",
			s1:       "./ayymd-foo-bar_",
			s2:       "./ayymd-foo_bar_",
			expected: "ayymd-foo",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := guessOverlapName(tc.s1, tc.s2)
			if actual != tc.expected {
				t.Errorf("expected %q but got %q", tc.expected, actual)
			}
		})
	}
}
