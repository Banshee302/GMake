package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("GMake")

	// Load all lines from GMake
	lines, err := os.ReadFile("GMake")
	if err != nil {
		fmt.Println("Error opening GMake file:", err)
		os.Exit(1)
	}

	// Parse all tasks and global lines
	parseGMake(string(lines))

	// Run requested task (default: "build")
	taskName := "build"
	if len(os.Args) > 1 {
		taskName = os.Args[1]
	}

	runTask(taskName)
}

