package anytime

import (
	"fmt"
	"math/rand"
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
