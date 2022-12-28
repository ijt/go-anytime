package anytime

import (
	"testing"
	"time"
)

func TestParse_ok(t *testing.T) {
	type args struct {
		b    []byte
		opts []Option
	}
	tests := []struct {
		name string
		args args
		want LocatedRange
	}{
		{
			name: "now",
			args: args{
				b: []byte("now"),
			},
			want: LocatedRange{
				RangeFn: func(ref time.Time, dir Direction) Range {
					return Range{
						start:    ref,
						Duration: time.Second,
					}
				},
				Pos:  0,
				Text: []byte("now"),
			},
		},
		{
			name: "now with leading verbiage",
			args: args{
				b: []byte("the time is now"),
			},
			want: LocatedRange{
				RangeFn: func(ref time.Time, dir Direction) Range {
					return Range{
						start:    ref,
						Duration: time.Second,
					}
				},
				Pos:  12,
				Text: []byte("now"),
			},
		},
		{
			name: "now with trailing verbiage",
			args: args{
				b: []byte("now is the time"),
			},
			want: LocatedRange{
				RangeFn: func(ref time.Time, dir Direction) Range {
					return Range{
						start:    ref,
						Duration: time.Second,
					}
				},
				Pos:  0,
				Text: []byte("now"),
			},
		},
		{
			name: "now with leading and trailing verbiage",
			args: args{
				b: []byte("Do you think that now is the time?"),
			},
			want: LocatedRange{
				RangeFn: func(ref time.Time, dir Direction) Range {
					return Range{
						start:    ref,
						Duration: time.Second,
					}
				},
				Pos:  18,
				Text: []byte("now"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse("input string", tt.args.b, tt.args.opts...)
			if err != nil {
				t.Fatal(err)
			}
			gotLocRange := got.(LocatedRange)
			now := time.Now()
			gotNow := gotLocRange.RangeFn(now, Future)
			wantNow := tt.want.RangeFn(now, Future)
			if !gotNow.Equal(wantNow) {
				t.Errorf("gotNow = %v, wantNow %v", gotNow, wantNow)
			}
		})
	}
}

func TestParse_fail(t *testing.T) {
	type args struct {
		b    []byte
		opts []Option
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "empty",
			args: args{
				b: []byte{},
			},
		},
		{
			name: "verbiage",
			args: args{
				b: []byte("not a date or a time"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse("input string", tt.args.b, tt.args.opts...)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
