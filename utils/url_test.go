package utils

import (
	"testing"

	"github.com/coroot/coroot/timeseries"
	"github.com/stretchr/testify/assert"
)

var now = timeseries.Now()

func TestParseTime(t *testing.T) {
	type args struct {
		now timeseries.Time
		val string
		def timeseries.Time
	}
	tests := []struct {
		name string
		args args
		want timeseries.Time
	}{
		{
			"case_LoadWorldByRequest_1",
			args{
				now: now,
				val: "now-12h",
				def: now.Add(-timeseries.Hour),
			},
			now.Add(-12 * timeseries.Hour),
		},
		{
			"case_LoadWorldByRequest_2",
			args{
				now: now,
				val: "now-12h",
				def: now,
			},
			now.Add(-12 * timeseries.Hour),
		},
		{
			"case_LoadWorldByRequest_3",
			args{
				now: now,
				val: "now",
				def: now,
			},
			now,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ParseTime(tt.args.now, tt.args.val, tt.args.def), "ParseTime(%v, %v, %v)", tt.args.now, tt.args.val, tt.args.def)
		})
	}
}
