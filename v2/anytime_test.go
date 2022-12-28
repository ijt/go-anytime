package anytime

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"
)

func TestReplaceAllRangesByFunc_ok(t *testing.T) {
	now := time.UnixMilli(rand.Int63())
	type args struct {
		s   string
		ref time.Time
		f   func(source string, r Range) string
		dir Direction
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				s: "",
			},
			want: "",
		},
		{
			name: "now",
			args: args{
				s:   "now",
				ref: now,
				f: func(source string, r Range) string {
					return fmt.Sprintf("%v", r.Start().UnixMilli())
				},
			},
			want: fmt.Sprintf("%v", now.UnixMilli()),
		},
		{
			name: "now with trailing verbiage",
			args: args{
				s:   "now is the time",
				ref: now,
				f: func(source string, r Range) string {
					return fmt.Sprintf("%v", r.Start().UnixMilli())
				},
			},
			want: fmt.Sprintf("%v is the time", now.UnixMilli()),
		},
		{
			name: "now with leading verbiage",
			args: args{
				s:   "the time is now",
				ref: now,
				f: func(source string, r Range) string {
					return fmt.Sprintf("%v", r.Start().UnixMilli())
				},
			},
			want: fmt.Sprintf("the time is %v", now.UnixMilli()),
		},
		{
			name: "now with leading and trailing verbiage",
			args: args{
				s:   "Without a doubt now is the time",
				ref: now,
				f: func(source string, r Range) string {
					return fmt.Sprintf("%v", r.Start().UnixMilli())
				},
			},
			want: fmt.Sprintf("Without a doubt %v is the time", now.UnixMilli()),
		},
		{
			name: "two nows",
			args: args{
				s:   "now now",
				ref: now,
				f: func(source string, r Range) string {
					return fmt.Sprintf("%v", r.Start().UnixMilli())
				},
			},
			want: fmt.Sprintf("%v %v", now.UnixMilli(), now.UnixMilli()),
		},
		{
			name: "two nows with no space between them",
			args: args{
				s:   "nownow",
				ref: now,
				f: func(source string, r Range) string {
					return fmt.Sprintf("%v", r.Start().UnixMilli())
				},
			},
			want: "nownow",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceAllRangesByFunc(tt.args.s, tt.args.ref, tt.args.f, tt.args.dir)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
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
