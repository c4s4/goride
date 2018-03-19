package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"sort"
	"flag"
)

var alpha = 0.8
var beta  = 2.2

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func distance(a, b, x, y int) int {
	return abs(x-a) + abs(y-b)
}

func parseLine(line string, length int) ([]int, error) {
	fields := strings.Split(line, " ")
	if len(fields) != length {
		return nil, fmt.Errorf("bad line length")
	}
	ints := make([]int, length)
	for i := 0; i < length; i++ {
		var err error
		ints[i], err = strconv.Atoi(fields[i])
		if err != nil {
			return nil, err
		}
	}
	return ints, nil
}

type City struct {
	Rows  int
	Cols  int
	Cars  int
	Rides int
	Bonus int
	Steps int
}

func NewCity(line string) (*City, error) {
	fields, err := parseLine(line, 6)
	if err != nil {
		return nil, err
	}
	city := City{
		Rows:  fields[0],
		Cols:  fields[1],
		Cars:  fields[2],
		Rides: fields[3],
		Bonus: fields[4],
		Steps: fields[5],
	}
	return &city, nil
}

type Ride struct {
	Index int
	A     int
	B     int
	X     int
	Y     int
	Start int
	End   int
	Len   int
	City  *City
}

func NewRide(index int, city *City, line string) (*Ride, error) {
	fields, err := parseLine(line, 6)
	if err != nil {
		return nil, err
	}
	ride := Ride{
		Index: index,
		A:     fields[0],
		B:     fields[1],
		X:     fields[2],
		Y:     fields[3],
		Start: fields[4],
		End:   fields[5],
		Len:   distance(fields[0], fields[1], fields[2], fields[3]),
		City:  city,
	}
	return &ride, nil
}

type Car struct {
	Index int
	Moves []*Move
	X     int
	Y     int
	T     int
}

func (c *Car) Add(move *Move) {
	c.Moves = append(c.Moves, move)
	c.X = move.X
	c.Y = move.Y
	c.T = move.End
}

func (c *Car) String() string {
	b := strings.Builder{}
	b.WriteString(strconv.Itoa(c.Index))
	b.WriteString(" ")
	for i := 0; i < len(c.Moves); i++ {
		b.WriteString(" ")
		b.WriteString(strconv.Itoa(c.Moves[i].Ride.Index))
	}
	return b.String()
}

type Move struct {
	Car   *Car
	Ride  *Ride
	A     int
	B     int
	X     int
	Y     int
	Start int
	End   int
	Score int
	Value float64
}

func NewMove(car *Car, ride *Ride) *Move {
	score := 0
	begin := max(car.T+distance(car.X, car.Y, ride.A, ride.B), ride.Start)
	end := begin + ride.Len
	if begin <= ride.Start {
		score += ride.City.Bonus
	}
	if end <= ride.End {
		score += ride.Len
	}
	value := alpha*float64(score)/float64(end-car.T) - beta*float64(end)/float64(ride.City.Steps)
	move := Move{
		Car:   car,
		Ride:  ride,
		A:     car.X,
		B:     car.Y,
		X:     ride.X,
		Y:     ride.Y,
		Start: car.T,
		End:   end,
		Score: score,
		Value: value,
	}
	return &move
}

func parse(source string) (*City, []*Ride, error) {
	lines := strings.Split(strings.TrimSpace(source), "\n")
	city, err := NewCity(lines[0])
	if err != nil {
		return nil, nil, err
	}
	rides := make([]*Ride, city.Rides)
	for i := 0; i < city.Rides; i++ {
		ride, err := NewRide(i, city, lines[i+1])
		if err != nil {
			return nil, nil, err
		}
		rides[i] = ride
	}
	return city, rides, nil
}

func assignRidesSort(city *City, rides []*Ride) []*Car {
	sort.Slice(rides, func(i, j int) bool {
		return rides[i].Start < rides[j].Start
	})
	cars := make([]*Car, city.Cars)
	for i := 0; i < city.Cars; i++ {
		cars[i] = &Car{Index: i}
	}
	index := 0
	for _, ride := range rides {
		car := cars[index]
		move := NewMove(car, ride)
		car.Add(move)
		index = (index + 1) % city.Cars
	}
	return cars
}

func assignRidesValue(city *City, rides []*Ride) []*Car {
	cars := make([]*Car, city.Cars)
	for i := 0; i < city.Cars; i++ {
		car := &Car{Index: i}
		cars[i] = car
		for car.T < city.Steps && len(rides) > 0 {
			var best *Move
			var index int
			for i, ride := range rides {
				move := NewMove(car, ride)
				if best == nil || best.Value < move.Value {
					best = move
					index = i
				}
			}
			rides[index] = rides[len(rides)-1]
			rides = rides[:len(rides)-1]
			car.Add(best)
		}
	}
	return cars
}

func writeFile(cars []*Car, file, output string) error {
	result := strings.Builder{}
	for _, car := range cars {
		result.WriteString(car.String())
		result.WriteString("\n")
	}
	path := filepath.Join(output, file[:len(file)-3]+".out")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(result.String())
	return nil
}

func computeScore(cars []*Car) int {
	score := 0
	for _, car := range cars {
		for _, move := range car.Moves {
			score += move.Score
		}
	}
	return score
}

func processFile(file, input, output string) (int, error) {
	fmt.Printf("%s:\n", file)
	start := time.Now()
	path := filepath.Join(input, file)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	source := strings.TrimSpace(string(bytes))
	city, rides, err := parse(source)
	if err != nil {
		return 0, err
	}
	cars := assignRidesValue(city, rides)
	duration := time.Since(start)
	fmt.Printf("  duration: %s\n", duration)
	score := computeScore(cars)
	fmt.Printf("  score: %d\n", score)
	if err := writeFile(cars, file, output); err != nil {
		return 0, err
	}
	return score, nil
}

func processDirectory(input, output string) error {
	files, err := ioutil.ReadDir(input)
	if err != nil {
		return err
	}
	score := 0
	report := strings.Builder{}
	for _, file := range files {
		s, err := processFile(file.Name(), input, output)
		if err != nil {
			return err
		}
		line := fmt.Sprintf("%-20v %d", file.Name()+":", s)
		report.WriteString(line)
		report.WriteRune('\n')
		score += s
	}
	fmt.Println("total:", score)
	line := fmt.Sprintf("%-20v %d", "total:", score)
	report.WriteString(line)
	report.WriteRune('\n')
	path := filepath.Join(output, "README")
	err = ioutil.WriteFile(path, []byte(report.String()), 0644)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Float64Var(&alpha, "alpha", alpha, "alpha constant")
	flag.Float64Var(&beta, "beta", beta, "beta constant")
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		fmt.Println("You must pass input and output directories on command line")
		os.Exit(1)
	}
	input := args[0]
	output := args[1]
	err := processDirectory(input, output)
	if err != nil {
		fmt.Println("ERROR: " + err.Error())
	}
}
