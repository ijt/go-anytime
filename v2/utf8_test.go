package anytime

import (
	"testing"
	"unicode/utf8"
)

func Test_fixUTF8(t *testing.T) {
	type args struct {
		s string
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
			name: "ascii",
			args: args{
				s: "abc",
			},
			want: "abc",
		},
		{
			name: "invalid",
			args: args{
				s: "\xd8",
			},
			want: "",
		},
		{
			name: "valid and invalid",
			args: args{
				s: "abc\xd8",
			},
			want: "abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixUTF8(tt.args.s); got != tt.want {
				t.Errorf("fixUTF8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Fuzz_fixUTF8(f *testing.F) {
	f.Add("\xd8")
	f.Add("abc")
	f.Add("\xd8abc")
	f.Fuzz(func(t *testing.T, s string) {
		s2 := fixUTF8(s)
		if !utf8.Valid([]byte(s2)) {
			t.Fatalf("got invalid utf8 result: %q", s2)
		}
	})
}
