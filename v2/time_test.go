package anytime

import (
	"reflect"
	"testing"
	"time"
)

func Test_truncateWeek(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "Sunday",
			args: args{
				t: time.Date(2022, 10, 2, 23, 59, 59, 999999, time.UTC),
			},
			want: time.Date(2022, 10, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "monday",
			args: args{
				t: time.Date(2022, 10, 3, 23, 59, 59, 999999, time.UTC),
			},
			want: time.Date(2022, 10, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "saturday",
			args: args{
				t: time.Date(2022, 10, 15, 23, 59, 59, 999999, time.UTC),
			},
			want: time.Date(2022, 10, 9, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateWeek(tt.args.t).Start(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("truncateWeek() = %v, want %v", got, tt.want)
			}
		})
	}
}
