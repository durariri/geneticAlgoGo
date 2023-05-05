package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"sort"
)

type GeneticAlgorithmSettings struct {
	PopulationSize           int
	MutationRate             int
	CrossoverRate            int
	NumGenerations           int
	KeepBestAcrossPopulation bool
}

type GeneticAlgorithmRunner interface {
	GenerateInitialPopulation(populationSize int) []Arg
	PerformCrossover(individual1, individual2 Arg, crossoverRate int) Arg
	PerformMutation(individual Arg, mutationRate int) Arg
	Sort([]Arg)
}

func createStochasticProbableListOfIndividuals(population []Arg) []Arg {
	totalCount, populationLength := 0, len(population)
	for j := 0; j < populationLength; j++ {
		totalCount += j
	}

	probableIndividuals := make([]Arg, 0, totalCount)
	for index, individual := range population {
		for i := 0; i < index; i++ {
			probableIndividuals = append(probableIndividuals, individual)
		}
	}

	return probableIndividuals
}

func Run(geneticAlgoRunner GeneticAlgorithmRunner, settings GeneticAlgorithmSettings) (Arg, []float64, error) {
	fitnessHistory := make([]float64, 0)

	population := geneticAlgoRunner.GenerateInitialPopulation(settings.PopulationSize)

	bestSoFar := population[len(population)-1]
	fmt.Printf("First Best: x: %f  y: %f  F(x, y): %f\n", bestSoFar.x, bestSoFar.y, calculate(bestSoFar))
	geneticAlgoRunner.Sort(population)

	bestSoFar = population[len(population)-1]
	fmt.Printf("First Best: x: %f  y: %f  F(x, y): %f\n", bestSoFar.x, bestSoFar.y, calculate(bestSoFar))
	for i := 0; i < settings.NumGenerations; i++ {

		newPopulation := make([]Arg, 0, settings.PopulationSize)

		if settings.KeepBestAcrossPopulation {
			newPopulation = append(newPopulation, bestSoFar)
		}

		// perform crossovers with random selection
		probabilisticListOfPerformers := createStochasticProbableListOfIndividuals(population)

		newPopIndex := 0
		if settings.KeepBestAcrossPopulation {
			newPopIndex = 1
		}
		for ; newPopIndex < settings.PopulationSize; newPopIndex++ {
			indexSelection1 := rand.Int() % len(probabilisticListOfPerformers)
			indexSelection2 := rand.Int() % len(probabilisticListOfPerformers)

			// crossover
			newIndividual := geneticAlgoRunner.PerformCrossover(
				probabilisticListOfPerformers[indexSelection1],
				probabilisticListOfPerformers[indexSelection2], settings.CrossoverRate)

			// mutate
			if rand.Intn(101) < settings.MutationRate {
				newIndividual = geneticAlgoRunner.PerformMutation(newIndividual, settings.MutationRate)
			}

			newPopulation = append(newPopulation, newIndividual)
		}

		population = newPopulation

		// sort by performance
		geneticAlgoRunner.Sort(population)

		// keep the best so far
		bestSoFar = population[len(population)-1]
		if i%50 == 0 {
			fmt.Printf("Best: x: %f  y: %f  F(x, y): %f\n", bestSoFar.x, bestSoFar.y, calculate(bestSoFar))
		}
		fitnessHistory = append(fitnessHistory, calculate(bestSoFar))
	}
	return bestSoFar, fitnessHistory, nil
}

type Arg struct {
	x, y float64
}

const highRange = 100.0

func makeNewEntry() float64 {
	return highRange * rand.Float64()
}

func makeNewQuadEntry(newX, newY float64) Arg {
	return Arg{
		x: newX,
		y: newY,
	}
}

func calculate(entry Arg) float64 {
	//booth (1;3) 0
	//a := entry.x + 2*entry.y - 7
	//b := 2*entry.x + entry.y - 5
	//return a*a + b*b

	//camel (0;0) 0
	// return 2*entry.x*entry.x - 1.05*math.Pow(entry.x, 4) + math.Pow(entry.x, 6)/6 + entry.x*entry.y + entry.y*entry.y

	//bill's (3;0.5) 0
	return math.Pow(1.5-entry.x+entry.x*entry.y, 2) + math.Pow(2.25-entry.x+math.Pow(entry.x*entry.y, 2), 2) + math.Pow(2.625-entry.x+math.Pow(entry.x*entry.y, 3), 2)
}

type GA struct {
}

func (l GA) GenerateInitialPopulation(populationSize int) []Arg {

	initialPopulation := make([]Arg, 0, populationSize)
	for i := 0; i < populationSize; i++ {
		initialPopulation = append(initialPopulation, makeNewQuadEntry(makeNewEntry(), makeNewEntry()))
	}

	return initialPopulation
}
func (l GA) PerformCrossover(result1, result2 Arg, _ int) Arg {
	return makeNewQuadEntry(
		(result1.x+result2.x)/2,
		(result1.y+result2.y)/2,
	)
}
func (l GA) PerformMutation(_ Arg, _ int) Arg {
	return makeNewQuadEntry(makeNewEntry(), makeNewEntry())
}
func (l GA) Sort(population []Arg) {
	sort.Slice(population, func(i, j int) bool {
		return calculate(population[i]) > calculate(population[j])
	})
}

func argMain() {
	settings := GeneticAlgorithmSettings{
		PopulationSize:           100,
		MutationRate:             102,
		CrossoverRate:            100,
		NumGenerations:           1000,
		KeepBestAcrossPopulation: true,
	}

	best, fitnessHistory, err := Run(GA{}, settings)
	if err != nil {
		println(err)
	} else {
		fmt.Printf("Best: x: %f  y: %f  F(x, y): %f\n", best.x, best.y, calculate(best))
	}

	img, err := createLineChart(fitnessHistory)
	if err != nil {
		fmt.Println(err)
	}
	printImage(img)
}
func main() {
	// Time := time.Now()
	argMain()
	// after=Time := time.Since(Time)
	// fmt.Printf("%d\n", afterTime)
}

func printImage(img image.Image) {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	fmt.Printf("\x1b]1337;File=inline=1:%s\a\n", imgBase64Str)
}

func createLineChart(data []float64) (image.Image, error) {
	// Create a new plot and set its dimensions
	p := plot.New()

	p.X.Label.Text = "Gens"
	p.Y.Label.Text = "Fitness"
	p.Add(plotter.NewGrid())

	// Create a new scatter plot with the data
	scatterData := make(plotter.XYs, len(data))
	for i, val := range data {
		scatterData[i].X = float64(i)
		scatterData[i].Y = val
	}
	s, err := plotter.NewScatter(scatterData)
	if err != nil {
		return nil, err
	}
	s.GlyphStyle.Color = color.RGBA{R: 255, A: 255}
	s.GlyphStyle.Radius = 2
	p.Add(s)

	// Draw the plot to a new image and return it
	canvas := vgimg.New(800, 400)
	p.Draw(draw.New(canvas))
	return canvas.Image(), nil
}
