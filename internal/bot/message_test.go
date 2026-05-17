package bot

import (
	"testing"

	"github.com/mymmrac/telego"
)

func TestResolveDisplayName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user telego.User
		want string
	}{
		{
			name: "username has priority",
			user: telego.User{
				Username:  "nick",
				FirstName: "John",
			},
			want: "nick",
		},
		{
			name: "fallback to first name",
			user: telego.User{
				FirstName: "John",
			},
			want: "John",
		},
		{
			name: "fallback to someone",
			user: telego.User{},
			want: "someone",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := resolveDisplayName(tc.user); got != tc.want {
				t.Fatalf("resolveDisplayName() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildIRCMeMessageHTML(t *testing.T) {
	t.Parallel()

	got := buildIRCMeMessageHTML(
		telego.User{
			Username: "<nick>",
		},
		"1 < 2 & 3",
	)

	want := "<i>@&lt;nick&gt; 1 &lt; 2 &amp; 3</i>"
	if got != want {
		t.Fatalf("buildIRCMeMessageHTML() = %q, want %q", got, want)
	}
}
