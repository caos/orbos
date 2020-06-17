package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/caos/orbos/internal/helpers"
)

func main() {

	args := os.Args[1:]

	location := ""
	if args[0] == "--http" {
		location = args[1]
		args = args[2:]
	}

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

	if location == "" {
		messages, err := executeChecks(args)
		if err != nil {
			panic(err)
		}
		fmt.Printf(messages)
		os.Exit(0)
	}

	pathStart := strings.Index(location, "/")
	listen := location[0:pathStart]
	path := location[pathStart:]
	http.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
		messages, err := executeChecks(args)
		if err != nil {
			writer.WriteHeader(http.StatusServiceUnavailable)
		}
		writer.Write([]byte(messages))
		request.Body.Close()
	})
	fmt.Printf("Serving healthchecks at %s", location)
	if err := http.ListenAndServe(listen, nil); err != nil {
		panic(err)
	}
}

func executeChecks(args []string) (string, error) {
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

	return messages, err
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
