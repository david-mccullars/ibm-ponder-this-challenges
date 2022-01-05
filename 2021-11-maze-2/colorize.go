package main

import (
	"fmt"
	"os"

	"github.com/gookit/color"
)

func colorize(colorName string, a ...interface{}) string {
	s := fmt.Sprint(a...)
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return color.Sprint("<", colorName, ">", s, "</>")
	}
	return s
}
