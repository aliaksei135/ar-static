package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aliaksei135/ar-static/hist"
	"github.com/aliaksei135/ar-static/sim"
	"github.com/aliaksei135/ar-static/util"
	"github.com/google/uuid"
	"gonum.org/v1/gonum/mat"

	"runtime"

	"strings"

	_ "github.com/mattn/go-sqlite3"

	_ "net/http/pprof"

	"github.com/urfave/cli/v2"
)

var (
	SIM_ID, _         = uuid.NewUUID()
	DEBUG_WRITE_MUTEX sync.Mutex
)

const (
	DEBUG = false
)

func simulateBatch(batch_size, batch_id int, chan_out chan []int64, bounds [6]float64, alt_hist, x_hist, y_hist hist.Histogram, target_density, own_velocity float64, path [][3]float64, conflict_dists [2]float64) {
	f, _ := os.OpenFile(fmt.Sprintf("debug/%v.csv", SIM_ID), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	csvWriter := csv.NewWriter(f)
	for i := 0; i < batch_size; i++ {
		seed := rand.Int63()
		traffic := sim.Traffic{Seed: seed, AltitudeDistr: alt_hist, XDistr: x_hist, YDistr: y_hist}
		traffic.Setup(bounds, target_density)

		ownship := sim.Ownship{Path: path, Velocity: own_velocity}
		ownship.Setup()

		sim := sim.Simulation{Traffic: traffic, Ownship: ownship, ConflictDistances: conflict_dists}
		sim.Run()
		sim.End()
		pos_sum := 0.0
		samples := int(math.Min(600, float64(len(sim.Traffic.Positions.RawMatrix().Data)-1)))
		for j := 0; j < samples; j++ {
			pos_sum += sim.Traffic.Positions.RawMatrix().Data[j]
		}
		chan_out <- []int64{int64(pos_sum), seed, int64(float64(sim.T) * sim.Timestep), int64(len(sim.ConflictLog))}
		if i%int(batch_size/20) == 0 {
			fmt.Printf("Completed %v sims (%v %%) in batch %v \n", i, 100*i/batch_size, batch_id)
		}

		if DEBUG && i%int(batch_size/5) == 0 {
			traffic_positions := mat.DenseCopyOf(&traffic.Positions)

			n_agents := traffic_positions.RawMatrix().Rows
			traffic_density := fmt.Sprint(float64(n_agents) / traffic.TotalVolume)

			agent_strs := make([][]string, n_agents)
			for k := 0; k < n_agents; k++ {
				posX := fmt.Sprint(traffic_positions.At(k, 0))
				posY := fmt.Sprint(traffic_positions.At(k, 1))
				posZ := fmt.Sprint(traffic_positions.At(k, 2))

				agent_str := []string{posX, posY, posZ, fmt.Sprint(n_agents), traffic_density}
				agent_strs[k] = agent_str
			}
			DEBUG_WRITE_MUTEX.Lock()
			_ = csvWriter.WriteAll(agent_strs)
			DEBUG_WRITE_MUTEX.Unlock()
		}
	}
	fmt.Printf("Completed batch %v \n", batch_id)
}

func main() {
	log.SetFlags(0)
	start := time.Now()

	if DEBUG {
		go func() {
			fmt.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	app := &cli.App{
		Version:     "0.1a",
		Usage:       "Specific Traffic ABS",
		Description: "Agent Based Traffic MAC Simulation",
		Flags: []cli.Flag{
			&cli.Float64SliceFlag{
				Name:     "bounds",
				Usage:    "W,E,S,N,B,T bounds in metres",
				Required: true,
			},
			&cli.Float64Flag{
				Name:     "target-density",
				Usage:    "Target background traffic density in ac/m^3",
				Required: true,
			},
			&cli.PathFlag{
				Name:     "altDataPath",
				Usage:    "Path to altitude data in metres as CSV",
				Required: true,
			},
			&cli.PathFlag{
				Name:     "xDataPath",
				Usage:    "Path to x data in m as CSV",
				Required: true,
			},
			&cli.PathFlag{
				Name:     "yDataPath",
				Usage:    "Path to y data in m as CSV",
				Required: true,
			},
			&cli.PathFlag{
				Name:     "ownPath",
				Usage:    "Path for ownship. Should be a nx3 CSV",
				Required: true,
			},
			&cli.Float64Flag{
				Name:  "ownVelocity",
				Usage: "Speed of the ownship along the defined path in m/s",
				Value: 60.0,
			},
			&cli.IntFlag{
				Name:  "simOps",
				Usage: "The total number of simulation runs to be done.",
				Value: 1e2,
			},
			&cli.Float64SliceFlag{
				Name:  "conflictDists",
				Usage: "X,Y distances in metres which define a conflict",
				Value: cli.NewFloat64Slice(15.0, 6.0),
			},
			&cli.PathFlag{
				Name:  "dbPath",
				Usage: "A path to the SQLite3 DB the results should be written to",
				Value: "./results.db",
			},
		},
		Action: func(ctx *cli.Context) error {
			bounds := (*[6]float64)(util.CheckSliceLen(ctx.Float64Slice("bounds"), 6))
			target_density := ctx.Float64("target-density")
			alt_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists(ctx.Path("altDataPath"))), 50)
			x_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists(ctx.Path("xDataPath"))), 50)
			y_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists(ctx.Path("yDataPath"))), 50)
			own_path := util.GetPathDataFromCSV(util.CheckPathExists(ctx.Path("ownPath")))
			own_velocity := ctx.Float64("ownVelocity")
			conflict_dist := (*[2]float64)(util.CheckSliceLen(ctx.Float64Slice("conflictDists"), 2))
			dbPath := ctx.Path("dbPath")
			simOps := ctx.Int("simOps")

			if strings.HasPrefix(strings.ToLower(dbPath), "s3://") {
				dbPath = filepath.Join(os.TempDir(), "results.db")
			}
			db, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()

			_, err = db.Exec("CREATE TABLE IF NOT EXISTS sims(id, seed, timesteps, n_conflicts)")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Created/Opened output database")

			result_chan := make(chan []int64)

			n_batches := runtime.NumCPU()
			batch_size := int(simOps / n_batches)
			fmt.Printf("Running %v batches of %v simulations\n", n_batches, batch_size)

			pathLength := util.GetPathLength(own_path)
			timestep := conflict_dist[0] / own_velocity
			expectedSteps := pathLength / own_velocity
			expectedSecs := expectedSteps / timestep
			batchHours := (expectedSecs * float64(batch_size)) / 3600
			batchTimesteps := expectedSecs * float64(n_batches) * float64(batch_size)
			fmt.Printf("Simulating %v total hrs, %v hrs per simulation, %v hrs per batch\n", batchHours*float64(n_batches), expectedSecs/3600, batchHours)
			fmt.Printf("Simulating %v total timesteps, %v timesteps per simulation, %v timesteps per batch\n", batchTimesteps*float64(n_batches), expectedSteps, batchTimesteps)

			for i := 0; i < n_batches; i++ {
				go simulateBatch(batch_size, i, result_chan, *bounds, alt_hist, x_hist, y_hist, target_density, own_velocity, own_path, *conflict_dist)
			}

			sim_results := make([][]int64, n_batches*batch_size)

			result_count := 0
			for results := range result_chan {
				sim_results[result_count] = results

				result_count++
				if result_count >= n_batches*batch_size {
					break
				}
			}
			fmt.Printf("Formatting %v results for database insertion\n", len(sim_results))
			value_fmt := "(%v, %v, %v, %v)"
			string_results := make([]string, len(sim_results))
			for idx, row := range sim_results {
				string_results[idx] = fmt.Sprintf(value_fmt, row[0], row[1], row[2], row[3])
			}
			values_str := strings.Join(string_results, ",")
			fmt.Println("Inserting results into database")
			_, err = db.Exec("INSERT INTO sims VALUES " + values_str)
			if err != nil {
				log.Fatal(err)
				return err
			}

			_, S3Upload := os.LookupEnv("S3_UPLOAD_RESULTS")
			if S3Upload {
				fmt.Println("Uploading results to S3...")
				util.UploadToS3(dbPath)
				fmt.Println("Uploaded results to S3")
			}

			elapsed := time.Since(start).Seconds()
			fmt.Printf("Completed successfully in %v seconds.\n %v ms per simulation.\n %v secs per simulated hour.\n", elapsed, elapsed*1000/float64(n_batches*batch_size), elapsed/(batchHours*float64(n_batches)))
			fmt.Print("Exiting...\n")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
