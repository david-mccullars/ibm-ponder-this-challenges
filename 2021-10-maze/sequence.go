package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/david-mccullars/maze-ibm/parallelsearch"
)

const (
	MOVE        uint8 = 0
	SLIDE_RIGHT       = 1
	SLIDE_DOWN        = 2
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
	case SLIDE_DOWN:
		return fmt.Sprint("D", self.argument)
	default:
		return ""
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

func (self *Sequence) Append(newMaze *Maze, newLocation uint8, commandOp uint8, commandArg uint8) *Sequence {
	return &Sequence{
		self.turnsRemaining - 1,
		newMaze,
		newLocation,
		Command{commandOp, commandArg},
		self,
	}
}

func (self *Sequence) SlideRight(row uint8) *Sequence {
	newMaze, newLocation := self.maze.SlideRight(row, self.location)
	return self.Append(newMaze, newLocation, SLIDE_RIGHT, row)
}

func (self *Sequence) SlideDown(column uint8) *Sequence {
	newMaze, newLocation := self.maze.SlideDown(column, self.location)
	return self.Append(newMaze, newLocation, SLIDE_DOWN, column)
}

func (self *Sequence) MoveSame() *Sequence {
	return self.Append(self.maze, self.location, MOVE, self.location)
}

func (self *Sequence) Move(newLocation uint8) *Sequence {
	return self.Append(self.maze, newLocation, MOVE, newLocation)
}

func (self *Sequence) MoveIfAccessible(newLocation uint8) *Sequence {
	if !self.CanMove(newLocation) {
		log.Fatal("CAN NOT MOVE TO ", newLocation)
	}
	return self.Move(newLocation)
}

func (self *Sequence) CanMove(newLocation uint8) bool {
	for accessibleLocation := range self.maze.AccessibleLocations(self.location) {
		if accessibleLocation == newLocation {
			return true
		}
	}
	return false
}

func (self *Sequence) CommandString() string {
	return self.command.String(self.maze.Columns())
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
				return cmdArg == row
			case SLIDE_DOWN:
				return cmdArg == column
			default:
				return false
			}
		}
		prev.maze.Draw(prev.location, highlighter)
	}
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
					onNext(self.Move(accessibleLocation))
				}
			}
		}
		for row := uint8(0); row < self.maze.Rows(); row++ {
			// Canonicalize consecutive right slides (sorted by row)
			// This avoids duplicating redundant slides (e.g. R0R1 vs R1R0)
			if cmd.operation != SLIDE_RIGHT || cmd.argument <= row {
				if cmd.argument == 2 || cmd.argument == 4 || cmd.argument == 9 {
					onNext(self.SlideRight(row))
				}
			}
		}
		for column := uint8(0); column < self.maze.Columns(); column++ {
			// Canonicalize consecutive down slides (sorted by column)
			// This avoids duplicating redundant slides (e.g. C0C1 vs C1C0)
			if cmd.operation != SLIDE_DOWN || cmd.argument <= column {
				if cmd.argument == 0 || cmd.argument == 9 {
					onNext(self.SlideDown(column))
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
