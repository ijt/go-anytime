package anytime

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"
)

func TestReplaceAllRangesByFunc_nows(t *testing.T) {
	now := time.UnixMilli(rand.Int63())
	nowRx := regexp.MustCompile(`(?i)\bnow\b`)
	f := func(source string, r Range) string {
		return fmt.Sprintf("%v", r.Start().UnixMilli())
	}
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"one space", " "},
		{"verbiage", "This string contains no times or dates."},
		{"now", "now"},
		{"space now", " now"},
		{"now space", "now "},
		{"space now space", " now "},
		{"all caps now", "NOW"},
		{"now verbiage", "now is the time"},
		{"verbiage now", "the time is now"},
		{"verbiage now verbiage", "Without a doubt now is the time."},
		{"verbiage now punctuation verbiage", "If you don't know me by now, you will never know me."},
		{"now now", "now now"},
		{"nownow", "nownow"},
		{"noise with two nows", "a;slas 😅dflasdjfla now laksjdfsdf  xxc,mnv as2w0  @#R$@$ 😑now😵‍💫  ;xlc x;c,nv.s,hriop4qu-u98dsvfjkldfljs $!@@#$WERTwe5u682470sZ)(*&Y)*("},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceAllRangesByFunc(tt.input, now, f, Future)
			if err != nil {
				t.Fatal(err)
			}
			want := nowRx.ReplaceAllString(tt.input, fmt.Sprintf("%v", now.UnixMilli()))
			if got != want {
				t.Errorf("got = %v, want %v", got, want)
			}
		})
	}
}

func BenchmarkReplaceAllRangesByFunc(b *testing.B) {
	now := time.UnixMilli(rand.Int63())
	s := `
Do It Now
by Berton Braley
IF WITH PLEASURE you are viewing any work a man is doing,
If you like him or you love him, tell him now;
Don't withhold your approbation till the parson makes oration
And he lies with snowy lilies on his brow;
No matter how you shout it he won't really care about it;
He won't know how many teardrops you have shed;
If you think some praise is due him now's the time to slip it to him,
For he cannot read his tombstone when he's dead.`
	nowRx := regexp.MustCompile(`(?i)\bnow\b`)
	want := nowRx.ReplaceAllString(s, fmt.Sprintf("%v", now.UnixMilli()))
	f := func(source string, r Range) string {
		return fmt.Sprintf("%v", r.Start().UnixMilli())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := ReplaceAllRangesByFunc(s, now, f, Past)
		if err != nil {
			b.Fatal(err)
		}
		if got != want {
			b.Errorf("got = %v\n\nwant %v", got, want)
		}
	}
}
