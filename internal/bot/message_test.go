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

func TestBuildIRCSlapMessageHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		sender telego.User
		target string
		want   string
	}{
		{
			name:   "with target",
			sender: telego.User{Username: "nick"},
			target: "bob",
			want:   "<i>* @nick slaps @bob around a bit with a large trout</i>",
		},
		{
			name:   "with @ prefix in target",
			sender: telego.User{Username: "nick"},
			target: "@bob",
			want:   "<i>* @nick slaps @bob around a bit with a large trout</i>",
		},
		{
			name:   "no target self-slap",
			sender: telego.User{Username: "nick"},
			target: "",
			want:   "<i>* @nick slaps @nick around a bit with a large trout</i>",
		},
		{
			name:   "html escaping",
			sender: telego.User{Username: "<me>"},
			target: "<target>",
			want:   "<i>* @&lt;me&gt; slaps @&lt;target&gt; around a bit with a large trout</i>",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := buildIRCSlapMessageHTML(tc.sender, tc.target); got != tc.want {
				t.Fatalf("buildIRCSlapMessageHTML() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestResolveSlapTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		commandText string
		message     *telego.Message
		want        string
	}{
		{
			name:        "explicit target",
			commandText: "bob",
			message:     &telego.Message{},
			want:        "bob",
		},
		{
			name:        "explicit target with @ prefix",
			commandText: "@bob",
			message:     &telego.Message{},
			want:        "@bob",
		},
		{
			name:        "no target no reply is empty",
			commandText: "",
			message:     &telego.Message{},
			want:        "",
		},
		{
			name:        "no target falls back to reply sender",
			commandText: "",
			message: &telego.Message{
				ReplyToMessage: &telego.Message{
					From: &telego.User{Username: "alice"},
				},
			},
			want: "alice",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := resolveSlapTarget(tc.commandText, tc.message); got != tc.want {
				t.Fatalf("resolveSlapTarget() = %q, want %q", got, tc.want)
			}
		})
	}
}
