package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var variables = make(map[string]string)

type TaskBlock struct {
	parallel bool
	lines    []string
}

var tasks = make(map[string][]TaskBlock)

func parseGMake(content string) {
	lines := strings.Split(content, "\n")
	var currentTask string
	var currentBlock *TaskBlock

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "task ") && strings.HasSuffix(line, ":") {
			currentTask = strings.TrimSuffix(strings.TrimPrefix(line, "task "), ":")
			tasks[currentTask] = []TaskBlock{}
			currentBlock = nil
		} else if currentTask != "" {
			if line == "PARALLEL:" {
				currentBlock = &TaskBlock{parallel: true}
				tasks[currentTask] = append(tasks[currentTask], *currentBlock)
			} else {
				if currentBlock == nil || currentBlock.parallel {
					currentBlock = &TaskBlock{parallel: false}
					tasks[currentTask] = append(tasks[currentTask], *currentBlock)
				}
				tasks[currentTask][len(tasks[currentTask])-1].lines = append(tasks[currentTask][len(tasks[currentTask])-1].lines, line)
			}
		} else {
			parseLine(line, false)
		}
	}
}

func runTask(name string) {
	blocks, ok := tasks[name]
	if !ok {
		fmt.Printf("Task '%s' not found.\n", name)
		return
	}
	fmt.Printf("Running task: %s\n", name)
	for _, block := range blocks {
		if block.parallel {
			var wg sync.WaitGroup
			for _, line := range block.lines {
				wg.Add(1)
				go func(cmd string) {
					defer wg.Done()
					parseLine(cmd, true)
				}(line)
			}
			wg.Wait()
		} else {
			for _, line := range block.lines {
				parseLine(line, true)
			}
		}
	}
}

func parseLine(line string, fromTask bool) {
	line = strings.TrimSpace(line)

	switch {
	case strings.HasPrefix(line, "$"):
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line[1:], "=", 2)
			if len(parts) == 2 {
				varName := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				variables[varName] = value
				fmt.Printf("Set variable: %s = %s\n", varName, value)
			}
		} else {
			cmd := substituteVars(line)
			fmt.Println("DEBUG: Executing command:", cmd)
			executeCommand(cmd)
		}
	case strings.HasPrefix(line, "PRINT ="):
		text := strings.Trim(strings.TrimPrefix(line, "PRINT ="), "\"")
		fmt.Println(text)
	case strings.HasPrefix(line, "OUT:"):
		out := strings.TrimSpace(strings.TrimPrefix(line, "OUT:"))
		out = substituteVars(out)

		if _, err := os.Stat(out); os.IsNotExist(err) {
			if _, err := os.Stat(out + ".exe"); err == nil {
				out += ".exe"
			}
		}

		if _, err := os.Stat(out); err == nil {
			fmt.Println("Output target:", out)
		} else {
			fmt.Println("ERROR: Output file not found:", out)
		}
	case line == "STOP":
		fmt.Println("Execution stopped.")
		os.Exit(0)
	default:
		if fromTask {
			cmd := substituteVars(line)
			fmt.Println("DEBUG: Executing command:", cmd)
			executeCommand(cmd)
		} else {
			fmt.Println("Unknown command:", line)
		}
	}
}

func substituteVars(input string) string {
	for k, v := range variables {
		input = strings.ReplaceAll(input, "$"+k, v)
	}
	return input
}

func ensureOutputDir(path string) {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

func executeCommand(cmd string) {
	if cmd == "" {
		return
	}

	if strings.Contains(cmd, "go build") && strings.Contains(cmd, "-o") {
		parts := strings.Fields(cmd)
		for i, p := range parts {
			if p == "-o" && i+1 < len(parts) {
				ensureOutputDir(parts[i+1])
			}
		}
	}

	if strings.HasPrefix(cmd, "rm -rf") {
		parts := strings.Fields(cmd)
		if len(parts) >= 3 {
			target := substituteVars(parts[2])
			err := os.RemoveAll(target)
			if err != nil {
				fmt.Println("Error deleting:", err)
			} else {
				fmt.Println("Deleted:", target)
			}
		}
		return
	}

	var command *exec.Cmd
	if isWindows() {
		command = exec.Command("cmd", "/C", cmd)
	} else {
		command = exec.Command("sh", "-c", cmd)
	}

	output, err := command.CombinedOutput()
	fmt.Println("DEBUG: Output:\n" + string(output))
	if err != nil {
		fmt.Println("ERROR:", err)
	}
}

func isWindows() bool {
	return strings.Contains(strings.ToLower(os.Getenv("OS")), "windows")
}
