package api

import "testing"

func TestSlugify(t *testing.T) {
	testCases := []struct {
		in   string
		want string
	}{
		{"Hello, world!", "hello-world"},
		{" Now   is the  time ", "now-is-the-time"},
		{"", ""},
		{"------", "-"},
		{" @#$%^&*( ____ ", "-"},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			if got := slugify(tc.in); tc.want != got {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}
