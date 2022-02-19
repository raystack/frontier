package logger

import (
	"go.uber.org/zap/zapcore"
	"testing"
)

func Test_atomicLevel(t *testing.T) {
	type args struct {
		level string
	}
	tests := []struct {
		name string
		args args
		want zapcore.Level
	}{
		{
			name: "debug",
			args: args{
				level: "debug",
			},
			want: zapcore.DebugLevel,
		},
		{
			name: "info",
			args: args{
				level: "info",
			},
			want: zapcore.InfoLevel,
		},
		{
			name: "warn",
			args: args{
				level: "warn",
			},
			want: zapcore.WarnLevel,
		},
		{
			name: "error",
			args: args{
				level: "error",
			},
			want: zapcore.ErrorLevel,
		},
		{
			name: "dpanic",
			args: args{
				level: "dpanic",
			},
			want: zapcore.InfoLevel,
		},
		{
			name: "panic",
			args: args{
				level: "panic",
			},
			want: zapcore.InfoLevel,
		},
		{
			name: "fatal",
			args: args{
				level: "fatal",
			},
			want: zapcore.FatalLevel,
		},
		{
			name: "invalid",
			args: args{
				level: "invalid",
			},
			want: zapcore.InfoLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := atomicLevel(tt.args.level); got != tt.want {
				t.Errorf("atomicLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
