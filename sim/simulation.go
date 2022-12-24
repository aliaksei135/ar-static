package sim

import (
	"math"
	"math/rand"

	"github.com/aliaksei135/ar-static/hist"

	"gonum.org/v1/gonum/mat"
)

type Traffic struct {
	//Setup
	x_bounds    [2]float64
	y_bounds    [2]float64
	z_bounds    [2]float64
	TotalVolume float64

	//Randomness
	XDistr        hist.Histogram
	YDistr        hist.Histogram
	AltitudeDistr hist.Histogram
	// VelocityDistr     hist.Histogram
	// TrackDistr        hist.Histogram
	// VerticalRateDistr hist.Histogram
	// SurfaceEntrance   bool
	randomSource  *rand.Rand
	target_agents int

	//State
	Positions mat.Dense
	Seed      int64
	Timestep  float64
}

func (tfc *Traffic) Setup(bounds [6]float64, target_density float64) {

	tfc.x_bounds[0] = bounds[0]
	tfc.x_bounds[1] = bounds[1]
	tfc.y_bounds[0] = bounds[2]
	tfc.y_bounds[1] = bounds[3]
	tfc.z_bounds[0] = bounds[4]
	tfc.z_bounds[1] = bounds[5]

	tfc.randomSource = rand.New(rand.NewSource(tfc.Seed))

	tfc.TotalVolume = math.Abs(tfc.x_bounds[1]-tfc.x_bounds[0]) * math.Abs(tfc.y_bounds[1]-tfc.y_bounds[0]) * math.Abs(tfc.z_bounds[1]-tfc.z_bounds[0])
	tfc.target_agents = int(math.Ceil(target_density * tfc.TotalVolume))

	tfc.Positions = *mat.NewDense(tfc.target_agents, 3, nil)

	tfc.AddAgents(tfc.target_agents)
}

func (tfc *Traffic) AddAgents(n_agents int) {
	xs := tfc.XDistr.Sample(n_agents, *tfc.randomSource)
	ys := tfc.YDistr.Sample(n_agents, *tfc.randomSource)
	zs := tfc.AltitudeDistr.Sample(n_agents, *tfc.randomSource)
	for i := 0; i < n_agents; i++ {
		tfc.Positions.Set(i, 0, xs[i])
		tfc.Positions.Set(i, 1, ys[i])
		tfc.Positions.Set(i, 2, zs[i])
	}
}

type Ownship struct {
	Path         [][3]float64
	position     [3]float64
	Velocity     float64
	Timestep     float64
	pathIndex    int
	stepVelocity float64
}

func (ownship *Ownship) Setup() {
	ownship.pathIndex = 1
	ownship.position = ownship.Path[0]
	ownship.stepVelocity = ownship.Velocity * ownship.Timestep
}

func (ownship *Ownship) Step() {
	sub_goal := ownship.Path[ownship.pathIndex]
	var vecToGoal [3]float64
	var stepToGoal [3]float64
	for i := range ownship.position {
		vecToGoal[i] = sub_goal[i] - ownship.position[i]
	}

	goalMagnitude := math.Sqrt((vecToGoal[0] * vecToGoal[0]) + (vecToGoal[1] * vecToGoal[1]) + (vecToGoal[2] * vecToGoal[2]))

	for i := range vecToGoal {
		stepToGoal[i] = (vecToGoal[i] * ownship.stepVelocity) / goalMagnitude
	}

	if ownship.stepVelocity > goalMagnitude {
		ownship.pathIndex += 1
	}
	for i := range stepToGoal {
		ownship.position[i] += stepToGoal[i]
	}
}

type Simulation struct {
	Traffic           Traffic
	Ownship           Ownship
	ConflictDistances [2]float64
	ConflictLog       [][3]float64
	T                 int
	Timestep          float64
	conflictRows      []int
}

func (sim *Simulation) Run() {
	sim.Timestep = sim.ConflictDistances[0] / sim.Ownship.Velocity
	sim.Ownship.Timestep = sim.Timestep
	sim.Ownship.Setup()

	for {
		if sim.Ownship.pathIndex >= len(sim.Ownship.Path) {
			sim.End()
			break
		}
		sim.Ownship.Step()

		for i := 0; i < sim.Traffic.Positions.RawMatrix().Rows; i++ {
			skip := false
			for c := 0; c < len(sim.conflictRows); c++ {
				if sim.conflictRows[c] == i {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			xy_dist := math.Sqrt((sim.Traffic.Positions.At(i, 0)-sim.Ownship.position[0])*(sim.Traffic.Positions.At(i, 0)-sim.Ownship.position[0]) + ((sim.Traffic.Positions.At(i, 1) - sim.Ownship.position[1]) * (sim.Traffic.Positions.At(i, 1) - sim.Ownship.position[1])))
			z_dist := math.Abs(sim.Traffic.Positions.At(i, 2) - sim.Ownship.position[2])
			if xy_dist < sim.ConflictDistances[0] && z_dist < sim.ConflictDistances[1] {
				sim.ConflictLog = append(sim.ConflictLog, [3]float64{sim.Ownship.position[0], sim.Ownship.position[1], sim.Ownship.position[2]})
				sim.conflictRows = append(sim.conflictRows, i)
			}
		}
		sim.T++
	}
}

func (sim *Simulation) End() {

}
