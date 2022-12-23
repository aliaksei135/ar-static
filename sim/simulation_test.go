package sim

import (
	"testing"

	"github.com/aliaksei135/ar-static/hist"
	"github.com/aliaksei135/ar-static/util"
)

func TestTraffic_Setup(t *testing.T) {
	alt_hist := hist.CreateHistogram(util.GetDataFromCSV("../test_data/alts.csv"), 200)
	x_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists("../test_data/x.csv")), 500)
	y_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists("../test_data/y.csv")), 500)
	type args struct {
		bounds         [6]float64
		target_density float64
	}
	tests := []struct {
		name string
		tfc  *Traffic
		args args
	}{
		{"Setup", &Traffic{Seed: 321, AltitudeDistr: alt_hist, XDistr: x_hist, YDistr: y_hist}, args{[6]float64{0, 1e4, 0, 1e4, 0, 1524}, 4e-9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tfc.Setup(tt.args.bounds, tt.args.target_density)
		})
	}
}

func TestOwnship_Step(t *testing.T) {
	path := [][3]float64{{1, 1, 200}, {300, 600, 800}, {2000, 5000, 900}, {3000, 6000, 200}}
	ownship := Ownship{Path: path, Velocity: 10.0}
	ownship.Setup()

	type args struct {
		timestep float64
	}
	tests := []struct {
		name    string
		ownship *Ownship
		args    args
	}{
		{"Step", &ownship, args{timestep: 1.0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ownship.Step()
		})
	}
}

func TestSimulation_Run(t *testing.T) {
	alt_hist := hist.CreateHistogram(util.GetDataFromCSV("../test_data/alts.csv"), 40)
	x_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists("../test_data/x.csv")), 500)
	y_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists("../test_data/y.csv")), 500)
	traffic := Traffic{Seed: 321, AltitudeDistr: alt_hist, XDistr: x_hist, YDistr: y_hist}
	traffic.Setup([6]float64{-145176.17270300398, -101964.24515822314, 6569893.199178016, 6595219.236650961, 0, 1524}, 1e-9)

	ownship := Ownship{Path: util.GetPathDataFromCSV("../test_data/path.csv"), Velocity: 70.0}
	// ownship.Setup()

	sim := Simulation{Traffic: traffic, Ownship: ownship, ConflictDistances: [2]float64{20, 20}}

	tests := []struct {
		name string
		sim  *Simulation
	}{
		{"Run Sim", &sim},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sim.Run()
		})
	}
}
