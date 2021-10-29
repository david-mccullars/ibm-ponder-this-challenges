package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/david-mccullars/maze-ibm/parallelsearch"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////

// Scenario is a maze to be solved within the given number of turns
type Scenario struct {
	Turns       uint8
	Columns     uint8
	Rows        uint8
	MazePattern string
}

func (self *Scenario) startSequence() *Sequence {
	maze := NewMaze(self.MazePattern, self.Columns)
	return NewSequence(maze, self.Turns)
}

func copyFileIfNotExist(src string, dst string) {
	_, err := os.Stat(dst)
	if !os.IsNotExist(err) {
		return
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		log.Fatal(err)
	}

	from, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer from.Close()

	to, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, srcInfo.Mode())
	if err != nil {
		log.Fatal(err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		log.Fatal(err)
	}
}

func loadScenario() *Scenario {
	copyFileIfNotExist("example-scenario.json", "scenario.json")

	dat, err := os.ReadFile("scenario.json")
	if err != nil {
		log.Fatal(err)
	}

	scenario := Scenario{}
	err = json.Unmarshal(dat, &scenario)
	if err != nil {
		log.Fatal(err)
	}

	return &scenario
}

/////////////////////////////////////////////////////////////////////////////////////////////////////

func parseArgs() (*Maze, uint8) {
	if len(os.Args) != 4 {
		usage()
	}

	mazePattern := strings.ToLower(os.Args[1])
	mazePatternIsValid, _ := regexp.Match("^[[:xdigit:]]+$", []byte(mazePattern))
	if !mazePatternIsValid {
		fmt.Fprintf(os.Stderr, "Maze pattern must be hex string\n")
		usage()
	}

	dimensions := strings.SplitN(os.Args[2], "x", 2)
	if len(dimensions) != 2 {
		fmt.Fprintf(os.Stderr, "Dimensions must be two digits, e.g. 4x5\n")
		usage()
	}
	rows := parseUint8(dimensions[0])
	columns := parseUint8(dimensions[1])
	if int(rows*columns) != len(mazePattern) {
		fmt.Fprintf(os.Stderr, "Maze pattern is not of size %s\n", os.Args[2])
		usage()
	}

	turns := parseUint8(os.Args[3])

	return NewMaze(mazePattern, columns), turns
}

func parseUint8(s string) uint8 {
	i, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		usage()
	}
	return uint8(i)
}

func usage() {
	fmt.Fprintf(os.Stderr, "USAGE: maze-ibm [PATTERN] [DIMENSIONS] [TURNS]\n")
	os.Exit(1)
}

func main() {
	runtime.GOMAXPROCS(16)

	startMaze, turns := parseArgs()
	startSequence := NewSequence(startMaze, turns)

	if turns == 0 {
		startSequence.PrintSummary()
		os.Exit(0)
	}

	ps := parallelsearch.New(
		128,        // poolSize
		int(turns), // searchDepth
		8,          // searchLimit
	)
	ps.Start(startSequence)

	found := ps.WaitForFound()
	for _, s := range found {
		sequence := s.(*Sequence)
		sequence.PrintSummary()
		break
	}
}
