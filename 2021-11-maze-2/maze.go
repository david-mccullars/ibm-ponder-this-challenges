package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ghetzel/go-stockutil/sliceutil"
)

type Maze struct {
	cells   []byte
	columns uint8
}

type Highlighter func(int, int) bool

func NewMaze(mazePattern string, columns uint8) *Maze {
	if len(mazePattern) > 255 {
		log.Fatal("Maze pattern is too large (must be < 256 characters)")
	} else if uint8(len(mazePattern))%columns != 0 {
		log.Fatal("Maze pattern has mismatched row sizes")
	}

	cells := make([]byte, len(mazePattern), len(mazePattern))
	for i, c := range strings.ToLower(mazePattern) {
		cells[i] = charToHex(c)
	}
	return &Maze{cells, columns}
}

func (self *Maze) Copy() *Maze {
	cells := make([]byte, self.TotalCells(), self.TotalCells())
	copy(cells, self.cells)
	return &Maze{cells, self.columns}
}

func (self *Maze) Columns() uint8 {
	return self.columns
}

func (self *Maze) Rows() uint8 {
	return self.TotalCells() / self.columns
}

func (self *Maze) TotalCells() uint8 {
	return uint8(len(self.cells))
}

func (self *Maze) Slide(command Command) *Maze {
	switch command.operation {
	case SLIDE_RIGHT:
		return self.SlideHorizontal(command.argument, true)
	case SLIDE_LEFT:
		return self.SlideHorizontal(command.argument, false)
	case SLIDE_DOWN:
		return self.SlideVertical(command.argument, true)
	case SLIDE_UP:
		return self.SlideVertical(command.argument, false)
	default:
		return self
	}
}

func (self *Maze) SlideHorizontal(row uint8, right bool) *Maze {
	if (right && row >= self.Rows()) || (!right && row == 0) {
		log.Fatal("Invalid row: ", row)
	}
	maze := self.Copy()

	idx1 := row * self.columns
	idx2 := idx1 + self.columns

	if right {
		maze.cells[idx1] = self.cells[idx2-1]
		copy(maze.cells[idx1+1:], self.cells[idx1:idx2-1])
	} else {
		copy(maze.cells[idx1:], self.cells[idx1+1:idx2-2])
	}

	return maze
}

func (self *Maze) SlideVertical(column uint8, down bool) *Maze {
	if column >= self.columns {
		log.Fatal("Invalid column: ", column)
	}
	maze := self.Copy()

	for r := uint8(1); r < self.Rows(); r++ {
		if down {
			maze.cells[r*self.columns+column] = self.cells[(r-1)*self.columns+column]
		} else {
			maze.cells[(r-1)*self.columns+column] = self.cells[r*self.columns+column]
		}
	}
	if down {
		maze.cells[column] = self.cells[(self.Rows()-1)*self.columns+column]
	} else {
		maze.cells[(self.Rows()-1)*self.columns+column] = self.cells[column]
	}

	return maze
}

func (self *Maze) AccessibleLocations(currentLocation uint8) <-chan uint8 {
	accessible := make(chan uint8)
	go func() {
		visited := new(big.Int)
		self.accessibleLocationsFrom(currentLocation, visited, accessible)
		close(accessible)
	}()
	return accessible
}

func (self *Maze) accessibleLocationsFrom(location uint8, visited *big.Int, accessible chan<- uint8) {
	if visited.Bit(int(location)) > 0 {
		return // Already visited
	}
	visited.SetBit(visited, int(location), 1)
	accessible <- location

	// If can go north
	if location > self.columns && self.cells[location]&8 > 0 && self.cells[location-self.columns]&2 > 0 {
		self.accessibleLocationsFrom(location-self.columns, visited, accessible)
	}
	// If can go east
	if (location+1)%self.columns != 0 && self.cells[location]&4 > 0 && self.cells[location+1]&1 > 0 {
		self.accessibleLocationsFrom(location+1, visited, accessible)
	}
	// If can go south
	if location+self.columns < self.TotalCells() && self.cells[location]&2 > 0 && self.cells[location+self.columns]&8 > 0 {
		self.accessibleLocationsFrom(location+self.columns, visited, accessible)
	}
	// If can go west
	if location%self.columns != 0 && self.cells[location]&1 > 0 && self.cells[location-1]&4 > 0 {
		self.accessibleLocationsFrom(location-1, visited, accessible)
	}
}

func (self *Maze) Draw(currentLocation uint8, highlighter Highlighter) {
	normalBlock := colorize("cyan", "██")
	//highlightedBlock := colorize("magenta", "██")
	highlightedBlock := colorize("magenta", "▓▓")
	me := colorize("yellow", "¥ ")

	var s1 strings.Builder
	var s2 strings.Builder
	var s3 strings.Builder

	for i := uint8(0); i < self.columns; i++ {
		fmt.Print("______")
	}
	fmt.Println("__")

	for row, rowData := range sliceutil.Chunks(self.cells, int(self.columns)) {
		s1.WriteRune('│')
		s2.WriteRune('│')
		s3.WriteRune('│')

		for column, cellData := range rowData {
			block := normalBlock
			if highlighter != nil && highlighter(row, column) {
				block = highlightedBlock
			}

			b := cellData.(byte)
			s1.WriteString(block)
			if b&8 == 0 {
				s1.WriteString(block)
			} else {
				s1.WriteString("  ")
			}
			s1.WriteString(block)

			if b&1 == 0 {
				s2.WriteString(block)
			} else {
				s2.WriteString("  ")
			}
			if uint8(row)*self.columns+uint8(column) == currentLocation {
				s2.WriteString(me)
			} else {
				s2.WriteString("  ")
			}
			if b&4 == 0 {
				s2.WriteString(block)
			} else {
				s2.WriteString("  ")
			}

			s3.WriteString(block)
			if b&2 == 0 {
				s3.WriteString(block)
			} else {
				s3.WriteString("  ")
			}
			s3.WriteString(block)
		}

		s1.WriteRune('│')
		s2.WriteRune('│')
		s3.WriteRune('│')
		fmt.Println(s1.String())
		fmt.Println(s2.String())
		fmt.Println(s3.String())
		s1.Reset()
		s2.Reset()
		s3.Reset()
	}

	for i := uint8(0); i < self.columns; i++ {
		fmt.Print("¯¯¯¯¯¯")
	}
	fmt.Println("¯¯")
}

func charToHex(c rune) byte {
	pattern := int(c)
	if pattern >= 48 && pattern <= 57 {
		return byte(pattern - 48)
	} else if pattern >= 97 && pattern <= 102 {
		return byte(pattern - 87)
	} else {
		log.Fatal("Invalid hex character: ", c)
		return 0
	}
}
