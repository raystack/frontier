package debounce

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimiter_Fn(t *testing.T) {
	var counter uint64

	type fields struct {
		after time.Duration
		want  uint64
	}
	type args struct {
		f func()
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "limit number of calls",
			fields: fields{
				after: 10 * time.Millisecond,
				want:  3,
			},
			args: args{
				f: func() {
					atomic.AddUint64(&counter, 1)
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New(tt.fields.after)

			for i := 0; i < 3; i++ {
				for j := 0; j < 10; j++ {
					d.Fn(tt.args.f)
				}

				time.Sleep(20 * time.Millisecond)
			}

			assert.Equal(t, tt.fields.want, atomic.LoadUint64(&counter))
		})
	}
}

func TestLimiter_Call(t *testing.T) {
	var counter uint64

	type fields struct {
		after time.Duration
		want  uint64
	}
	type args struct {
		f func()
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "limit number of calls",
			fields: fields{
				after: 10 * time.Millisecond,
				want:  3,
			},
			args: args{
				f: func() {
					atomic.AddUint64(&counter, 1)
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New(tt.fields.after)

			for i := 0; i < 3; i++ {
				for j := 0; j < 10; j++ {
					d.Call(tt.args.f)
					time.Sleep(1 * time.Millisecond)
				}
				// even though we call the function 10 times, it should only be called once
				// notice we don't have to wait for the timer to expire as the first call
				// scheduled the function to be called after the duration
				time.Sleep(1 * time.Millisecond)
			}

			assert.Equal(t, tt.fields.want, atomic.LoadUint64(&counter))
		})
	}
}
