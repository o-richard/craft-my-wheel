package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

const (
	EXIT = "exit"
	PWD  = "pwd"
	CD   = "cd"
	ECHO = "echo"
	TYPE = "type"
	CAT  = "cat"
)

var (
	PATHS    = strings.Split(os.Getenv("PATH"), ":")
	BUILTINS = []string{EXIT, PWD, CD, ECHO, TYPE}
)

func main() {
	for {
		fmt.Printf("$ ")
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("unable to read input,", err)
			os.Exit(1)
		}
		cleanedInput := []byte(input[:len(input)-1])
		command, arg := parseInput(cleanedInput)
		output := executeInput(command, arg)
		if output != "" {
			fmt.Println(output)
		}
	}
}

func parseInput(input []byte) (command, arg string) {
	var sep byte = ' '
	var sepPresence bool
	var firstArgEndIndex, secondArgStartIndex int
	for i := range input {
		if i == 0 {
			if input[i] == '\'' || input[i] == '"' {
				sep = input[i]
			}
			continue
		}
		if input[i] == sep {
			switch sep {
			case ' ':
				firstArgEndIndex = i
				secondArgStartIndex = firstArgEndIndex + 1
			default:
				firstArgEndIndex = i + 1
				secondArgStartIndex = firstArgEndIndex + 1
				if secondArgStartIndex > len(input) {
					secondArgStartIndex = len(input)
				}
			}
			sepPresence = true
			break
		}
	}
	if !sepPresence {
		firstArgEndIndex = len(input)
		secondArgStartIndex = firstArgEndIndex
	}
	return string(input[:firstArgEndIndex]), string(input[secondArgStartIndex:])
}

func buildString(s []byte, startIndex, lastIndex, spaceCheckIndex int, indicesToTrim []int) (original, sanitized string) {
	var output strings.Builder
	current := startIndex
	for _, index := range indicesToTrim {
		output.Write(s[current:index])
		current = index + 1
	}
	if current <= lastIndex {
		output.Write(s[current:lastIndex])
	}
	original = output.String()
	sanitized = original
	if spaceCheckIndex < len(s) && s[spaceCheckIndex] == ' ' {
		sanitized += " "
	}
	return original, sanitized
}

func parseArg(arg []byte) (args []string, stringifiedArgs string) {
	var stringArgs strings.Builder

	var startIndex int
	var currentdelim byte
	var indicesToTrim []int
	maxIndex := len(arg) - 1
	for i := range arg {
		nextIndex := i + 1
		prevIndex := i - 1
		prevPrevIndex := prevIndex - 1

		switch arg[i] {
		case '\'':
			switch currentdelim {
			case '\'':
				original, sanitized := buildString(arg, startIndex, i, i+1, nil)
				stringArgs.WriteString(sanitized)
				args = append(args, original)
				startIndex = i + 1
				currentdelim = 0
			case 0:
				startIndex = i + 1
				currentdelim = '\''
			}
			continue
		case '"':
			switch currentdelim {
			case '"':
				if arg[prevIndex] != '\\' || (prevPrevIndex >= 0 && arg[prevPrevIndex] == '\\') {
					original, sanitized := buildString(arg, startIndex, i, i+1, indicesToTrim)
					stringArgs.WriteString(sanitized)
					args = append(args, original)
					startIndex = i + 1
					currentdelim = 0
					indicesToTrim = nil
					continue
				}
			case 0:
				if prevIndex < 0 || arg[prevIndex] != '\\' || (prevPrevIndex >= 0 && arg[prevPrevIndex] == '\\') {
					startIndex = i + 1
					currentdelim = '"'
					continue
				}
			}
		case ' ':
			if currentdelim == 0 {
				startIndex = i + 1
			}
			continue
		case '\\':
			if currentdelim == '\'' {
				continue
			}
			noDelimSkip := nextIndex <= maxIndex && currentdelim == 0
			quotationDelimSkip := nextIndex <= maxIndex && currentdelim == '"' && (arg[nextIndex] == '\\' || arg[nextIndex] == '"')
			if (noDelimSkip || quotationDelimSkip) && (prevIndex < 0 || arg[prevIndex] != '\\') {
				indicesToTrim = append(indicesToTrim, i)
			}
		}
		if currentdelim == 0 && (nextIndex > maxIndex || arg[nextIndex] == ' ') {
			original, sanitized := buildString(arg, startIndex, nextIndex, nextIndex, indicesToTrim)
			stringArgs.WriteString(sanitized)
			args = append(args, original)
			startIndex = i + 1
			indicesToTrim = nil
		}
	}

	return args, stringArgs.String()
}

func absoluteBinpath(name string) string {
	for _, path := range PATHS {
		fullpath := filepath.Join(path, name)
		if _, err := os.Stat(fullpath); err == nil {
			return fullpath
		}
	}
	return ""
}

func executeInput(command, arg string) string {
	if command == EXIT {
		os.Exit(0)
	}
	if command == PWD {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Sprintf("unable to obtain current working directory, %v", err)
		}
		return dir
	}
	if command == CD {
		if arg == "~" {
			arg = os.Getenv("HOME")
		}
		if err := os.Chdir(arg); err != nil {
			return fmt.Sprintf("cd: %v: No such file or directory", arg)
		}
		return ""
	}
	if command == ECHO {
		_, output := parseArg([]byte(arg))
		return output
	}
	if command == TYPE {
		if slices.Contains(BUILTINS, arg) {
			return fmt.Sprintf("%v is a shell builtin", arg)
		}
		fullpath := absoluteBinpath(arg)
		if fullpath != "" {
			return fmt.Sprintf("%v is %v", arg, fullpath)
		}
		return fmt.Sprintf("%v: not found", arg)
	}
	_, commandType := parseArg([]byte(command))
	if fullpath := absoluteBinpath(commandType); fullpath != "" {
		args := []string{arg}
		if commandType == CAT {
			args, _ = parseArg([]byte(arg))
		}
		cmd := exec.Command(fullpath, args...)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Sprintf("unable to run external program, %v", err)
		}
		return strings.TrimRightFunc(string(output), unicode.IsSpace)
	}
	return fmt.Sprintf("%v: command not found", command)
}
