//go:generate goderive .

package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/caos/orbiter/internal/core/helpers"
)

type resultTuple func() (interface{}, error)

func main() {

	args := os.Args[1:]

	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	if len(args) < 1 {
		panic("No arguments")
	}

	messages := ""
	var err error
	var wg sync.WaitGroup

	for argument := range toChan(args) {
		wg.Add(1)
		go func(arg string) {
			msg, goErr := check(arg)
			err = helpers.Concat(err, goErr)
			messages += msg + "\n"
			wg.Done()
		}(argument)
	}

	wg.Wait()

	if err != nil {
		panic(err)
	}

	fmt.Printf(messages)
}

func toChan(args []string) <-chan string {
	ch := make(chan string)
	go func() {
		for _, arg := range args {
			ch <- arg
		}
		close(ch)
	}()
	return ch
}
