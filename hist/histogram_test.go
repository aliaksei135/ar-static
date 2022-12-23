package hist

import (
	"math/rand"

	"github.com/aliaksei135/ar-static/util"

	"reflect"
	"testing"
)

func TestCreateHistogram(t *testing.T) {
	type args struct {
		data     []float64
		num_bins int
	}
	tests := []struct {
		name string
		args args
		want Histogram
	}{
		{"Altitudes", args{util.GetDataFromCSV("../test_data/alts.csv"), 20},
			Histogram{
				cdf:           []float64{0.02650602409638554, 0.02891566265060241, 0.04578313253012048, 0.05783132530120482, 0.07951807228915662, 0.12289156626506025, 0.1855421686746988, 0.25783132530120484, 0.3289156626506024, 0.3710843373493976, 0.4867469879518072, 0.5602409638554217, 0.6289156626506024, 0.7, 0.7506024096385542, 0.8120481927710843, 0.8650602409638555, 0.9108433734939759, 0.9602409638554217, 1},
				bin_midpoints: []float64{179.79002624671915, 375.3280839895013, 570.8661417322835, 766.4041994750655, 961.9422572178478, 1157.48031496063, 1353.0183727034118, 1548.556430446194, 1744.0944881889764, 1939.6325459317586, 2135.1706036745404, 2330.7086614173227, 2526.246719160105, 2721.7847769028867, 2917.322834645669, 3112.8608923884512, 3308.3989501312335, 3503.9370078740158, 3699.475065616798, 3895.01312335958},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateHistogram(tt.args.data, tt.args.num_bins); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateHistogram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHistogram_Sample(t *testing.T) {
	alt_histogram := CreateHistogram(util.GetDataFromCSV("../test_data/alts.csv"), 20)
	type args struct {
		num int
	}
	tests := []struct {
		name string
		hist *Histogram
		args args
		want []float64
	}{
		{"Altitudes", &alt_histogram, args{1}, []float64{2721.7847769028867}},
		{"Altitudes Multiple", &alt_histogram, args{20}, []float64{2721.7847769028867, 2135.1706036745404, 1353.0183727034118, 3895.01312335958, 2721.7847769028867, 1157.48031496063, 1744.0944881889764, 3112.8608923884512, 1548.556430446194, 3895.01312335958, 3503.9370078740158, 2917.322834645669, 1353.0183727034118, 1548.556430446194, 1939.6325459317586, 1548.556430446194, 3308.3989501312335, 2526.246719160105, 1939.6325459317586, 3699.475065616798}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.hist.Sample(tt.args.num, *rand.New(rand.NewSource(324))); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Histogram.Sample() = %v, want %v", got, tt.want)
			}
		})
	}
}
