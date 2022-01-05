package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/david-mccullars/maze-ibm/parallelsearch"
)

const (
	MOVE        uint8 = 0
	SLIDE_RIGHT       = 1
	SLIDE_LEFT        = 2
	SLIDE_DOWN        = 3
	SLIDE_UP          = 4
)

type Command struct {
	operation uint8
	argument  uint8
}

func (self Command) String(columns uint8) string {
	switch self.operation {
	case MOVE:
		row := self.argument / columns
		column := self.argument % columns
		return fmt.Sprint("(", row, ",", column, ")")
	case SLIDE_RIGHT:
		return fmt.Sprint("R", self.argument)
	case SLIDE_LEFT:
		return fmt.Sprint("L", self.argument)
	case SLIDE_DOWN:
		return fmt.Sprint("D", self.argument)
	case SLIDE_UP:
		return fmt.Sprint("U", self.argument)
	default:
		return ""
	}
}

func ParseCommand(maze *Maze, command string) Command {
	switch strings.ToUpper(command)[0] {
	case 'R':
		a, err := strconv.Atoi(command[1:])
		if err != nil || a < 0 || a >= int(maze.Rows()) {
			log.Fatal("Invalid shift:", err)
		}
		return Command{SLIDE_RIGHT, uint8(a)}
	case 'L':
		a, err := strconv.Atoi(command[1:])
		if err != nil || a < 0 || a >= int(maze.Rows()) {
			log.Fatal("Invalid shift:", err)
		}
		return Command{SLIDE_LEFT, uint8(a)}
	case 'D':
		a, err := strconv.Atoi(command[1:])
		if err != nil || a < 0 || a >= int(maze.Columns()) {
			log.Fatal("Invalid shift:", err)
		}
		return Command{SLIDE_DOWN, uint8(a)}
	case 'U':
		a, err := strconv.Atoi(command[1:])
		if err != nil || a < 0 || a >= int(maze.Columns()) {
			log.Fatal("Invalid shift:", err)
		}
		return Command{SLIDE_UP, uint8(a)}
	case '(':
		args := strings.SplitN(command[1:len(command)-1], ",", 2)
		r, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal("Invalid movement:", err)
		}
		c, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatal("Invalid movement:", err)
		}
		if r < 0 || r >= int(maze.Rows()) || c < 0 || c >= int(maze.Columns()) {
			log.Fatal("Movement is out of boundaries:", command)
		}
		return Command{MOVE, uint8(r)*maze.Columns() + uint8(c)}
	default:
		log.Fatal("Can not parse command:", command)
		return Command{0, 0}
	}
}

// Sequence is a list of commands that have been run with the state of the maze arrived at by these
// commands
type Sequence struct {
	turnsRemaining uint8
	maze           *Maze
	location       uint8
	command        Command
	prev           *Sequence
}

func NewSequence(maze *Maze, turns uint8) *Sequence {
	return &Sequence{turns, maze, 0, Command{}, nil}
}

func (self *Sequence) Slide(command Command) *Sequence {
	return &Sequence{
		self.turnsRemaining - 1,
		self.maze.Slide(command),
		self.location,
		command,
		self,
	}
}

func (self *Sequence) MoveSame() *Sequence {
	return self.Move(Command{MOVE, self.location})
}

func (self *Sequence) Move(command Command) *Sequence {
	return &Sequence{
		self.turnsRemaining - 1,
		self.maze,
		command.argument,
		command,
		self,
	}
}

func (self *Sequence) CanMove(newLocation uint8) bool {
	for accessibleLocation := range self.maze.AccessibleLocations(self.location) {
		if accessibleLocation == newLocation {
			return true
		}
	}
	return false
}

func (self *Sequence) CanSlideHorizontal(row uint8) bool {
	return self.location/self.maze.Columns() != row
}

func (self *Sequence) CanSlideVertical(column uint8) bool {
	return self.location%self.maze.Columns() != column
}

func (self *Sequence) CanApply(command Command) bool {
	switch command.operation {
	case MOVE:
		return self.CanMove(command.argument)
	case SLIDE_RIGHT:
		fallthrough
	case SLIDE_LEFT:
		return self.CanSlideHorizontal(command.argument)
	case SLIDE_DOWN:
		fallthrough
	case SLIDE_UP:
		return self.CanSlideVertical(command.argument)
	default:
		return true
	}
}

func (self *Sequence) Apply(command Command) *Sequence {
	switch command.operation {
	case MOVE:
		return self.Move(command)
	case SLIDE_RIGHT:
		fallthrough
	case SLIDE_LEFT:
		fallthrough
	case SLIDE_DOWN:
		fallthrough
	case SLIDE_UP:
		return self.Slide(command)
	default:
		return self
	}
}

func (self *Sequence) CommandString() string {
	return self.command.String(self.maze.Columns())
}

func (self *Sequence) CommandFromString(text string) Command {
	return ParseCommand(self.maze, text)
}

func (self *Sequence) PrintSummary() {
	var s strings.Builder

	fmt.Println()
	fmt.Println(colorize("yellow", "################################################################################"))
	fmt.Println()
	stack := []*Sequence{}
	for prev := self; prev != nil; prev = prev.prev {
		stack = append([]*Sequence{prev}, stack...)
	}
	for i, prev := range stack {
		if i > 0 {
			s.WriteString(prev.CommandString())
			s.WriteString(" ")
			fmt.Println(">>>", prev.CommandString())
		}
		highlighter := func(row int, column int) bool {
			cmdArg := int(prev.command.argument)
			switch prev.command.operation {
			case MOVE:
				return cmdArg == row*int(prev.maze.Columns())+column
			case SLIDE_RIGHT:
				fallthrough
			case SLIDE_LEFT:
				return cmdArg == row
			case SLIDE_DOWN:
				fallthrough
			case SLIDE_UP:
				return cmdArg == column
			default:
				return false
			}
		}
		prev.maze.Draw(prev.location, highlighter)
	}
	fmt.Println("SOLUTION:", colorize("green", s.String()))
}

func (self *Sequence) Draw() {
	var s strings.Builder

	stack := []*Sequence{}
	for prev := self; prev != nil; prev = prev.prev {
		stack = append([]*Sequence{prev}, stack...)
	}
	for i, prev := range stack {
		if i > 0 {
			s.WriteString(prev.CommandString())
			s.WriteString(" ")
		}
	}

	highlighter := func(row int, column int) bool {
		cmdArg := int(self.command.argument)
		switch self.command.operation {
		case MOVE:
			return cmdArg == row*int(self.maze.Columns())+column
		case SLIDE_RIGHT:
			fallthrough
		case SLIDE_LEFT:
			return cmdArg == row
		case SLIDE_DOWN:
			fallthrough
		case SLIDE_UP:
			return cmdArg == column
		default:
			return false
		}
	}
	self.maze.Draw(self.location, highlighter)
	fmt.Println("SOLUTION:", colorize("green", s.String()))
}

// Search implements Searchable interface for continuing the search from this sequence into a
// subsequence sequence by taking an available (and legal) action
func (self *Sequence) Search(onNext func(parallelsearch.Searchable)) {
	if self.turnsRemaining > 0 {
		cmd := self.command
		if self.prev == nil || cmd.operation != MOVE {
			for accessibleLocation := range self.maze.AccessibleLocations(self.location) {
				if self.location != accessibleLocation {
					onNext(self.Move(Command{MOVE, accessibleLocation}))
				}
			}
		}
		for row := uint8(0); row < self.maze.Rows(); row++ {
			// Canonicalize consecutive right slides (sorted by row)
			// This avoids duplicating redundant slides (e.g. R0R1 vs R1R0)
			if cmd.operation != SLIDE_RIGHT || cmd.argument <= row {
				if cmd.argument == 2 || cmd.argument == 4 || cmd.argument == 9 {
					onNext(self.Slide(Command{SLIDE_RIGHT, row}))
				}
			}
		}
		for column := uint8(0); column < self.maze.Columns(); column++ {
			// Canonicalize consecutive down slides (sorted by column)
			// This avoids duplicating redundant slides (e.g. C0C1 vs C1C0)
			if cmd.operation != SLIDE_DOWN || cmd.argument <= column {
				if cmd.argument == 0 || cmd.argument == 9 {
					onNext(self.Slide(Command{SLIDE_LEFT, column}))
				}
			}
		}
	}
}

// IsFound implements Searchable interface to determine if the current sequence meets the goal
// we are looking for
func (self *Sequence) IsFound() bool {
	return self.location == self.maze.TotalCells()-1 // At exit
}

// Score implements Searchable interface and provides the ability to sort the discovered solutions
// to try and present the "best" solution first.
func (self *Sequence) Score() int {
	return int(self.turnsRemaining)
}
