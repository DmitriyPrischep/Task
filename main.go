package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sync"
)

func main() {
	fs := FileSystemDriver{}
	total := uint64(0)
	net := NetDriver{
		httpClient: http.DefaultClient,
	}
	input, output, errors := NewWordsCounter("Go", &fs, &net, 5)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		chanels := []reflect.SelectCase{
			reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(output),
			},
			reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(errors),
			},
		}
		outOk := true
		errorsOk := true
		for outOk || errorsOk {
			i, v, ok := reflect.Select(chanels)
			if i == 0 {
				if !ok {
					outOk = false
					continue
				}
				val := v.Interface().(WordsCountResult)
				total += val.RCount
				fmt.Printf("%s: %d \n", val.Input, val.RCount)
			}
			if i == 1 {
				if !ok {
					errorsOk = false
					continue
				}
				val := v.Interface().(error)
				fmt.Printf("%s", val.Error())
			}
		}
	}()

	reader := bufio.NewScanner(os.Stdin)
	for reader.Scan() {
		line := reader.Text()
		inpTask := WordsCountInput{
			Input: line,
			Ctx:   context.Background(),
		}
		input <- inpTask
	}
	close(input)
	wg.Wait()
	fmt.Printf("Total: %d\n", total)
}
