package utils

import "testing"

func TestDefaultStringIfEmpty(t *testing.T) {
	t.Parallel()
	type args struct {
		str           string
		defaultString string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{
				str:           "",
				defaultString: "default",
			},
			want: "default",
		},
		{
			name: "not empty string",
			args: args{
				str:           "not empty",
				defaultString: "default",
			},
			want: "not empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultStringIfEmpty(tt.args.str, tt.args.defaultString); got != tt.want {
				t.Errorf("DefaultStringIfEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	t.Parallel()
	type args struct {
		str     string
		options SlugifyOptions
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{
				str:     "",
				options: SlugifyOptions{},
			},
			want: "",
		},
		{
			name: "not empty string",
			args: args{
				str:     "not empty",
				options: SlugifyOptions{},
			},
			want: "not_empty",
		},
		{
			name: "not empty string with separator",
			args: args{
				str:     "NOT EMPTY",
				options: SlugifyOptions{KeepHyphen: true},
			},
			want: "not_empty",
		},
		{
			name: "not empty string with separator and lowercase",
			args: args{
				str:     "not empty",
				options: SlugifyOptions{KeepHyphen: true},
			},
			want: "not_empty",
		},
		{
			name: "not empty string with separator and lowercase and keep colon",
			args: args{
				str:     "not:empty",
				options: SlugifyOptions{KeepColon: true},
			},
			want: "not:empty",
		},
		{
			name: "not empty string with separator and lowercase and keep hash",
			args: args{
				str:     "not#empty",
				options: SlugifyOptions{KeepHash: true},
			},
			want: "not#empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slugify(tt.args.str, tt.args.options); got != tt.want {
				t.Errorf("Slugify() = %v, want %v", got, tt.want)
			}
		})
	}
}
