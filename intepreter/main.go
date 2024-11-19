package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/o-richard/intepreter/marble"
)

func main() {
	var filepath string
	flag.StringVar(&filepath, "filepath", "", "the path of the file to open")
	flag.Parse()
	if filepath == "" {
		flag.Usage()
		os.Exit(1)
	}

	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("unable to open file, ", err)
		os.Exit(1)
	}
	defer file.Close()

	input, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("unable to read file, ", err)
		return
	}

	l := marble.NewLexer(input)
	p := marble.NewParser(l)
	program := p.ParseProgram()
	if errors := p.Errors(); errors != nil {
		fmt.Println("parsing errors, ", errors)
		return
	}
	evaluated := marble.Eval(program, marble.NewEnvironment())
	var actuatlOutput string
	if evaluated != nil {
		actuatlOutput = evaluated.String()
	}
	fmt.Println(actuatlOutput)
}
