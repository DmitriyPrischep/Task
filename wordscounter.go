package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
)

var urlRegexp = `^(http|ftp|https)://([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?$`
var fileNameRegexp = `^\/(.+\/)*(.+)$`

// WordsCountResult структура для описания ответа
type WordsCountResult struct {
	Input  string // имя полученное на вход
	RCount uint64 // количество вхождений искомой строки
}

// WordsCountInput структура для описания запроса
type WordsCountInput struct {
	Input string          // имя файла или url
	Ctx   context.Context // контекст задачи для отмены
}

// NewWordsCounter инициализирует воркеров и возвращает каналы для входа
// выхода и ошибок, пользователь должен закрыть канал входа самостоятельно
// каналы выхода и ошибок будут закрыты автоматически
func NewWordsCounter(word string, fs FileSystem, net Network, numWorkers uint8) (input chan WordsCountInput, output chan WordsCountResult, errors chan error) {
	input = make(chan WordsCountInput)
	output = make(chan WordsCountResult)
	errors = make(chan error)
	var wg sync.WaitGroup
	for i := uint8(0); i < numWorkers; i++ {
		wg.Add(1)
		go worker(word, fs, net, input, output, errors, &wg)
	}
	go func() {
		wg.Wait()
		close(output)
		close(errors)
	}()
	return
}

func countSubstring(reader io.Reader, ctx context.Context, substring string) (uint64, error) {
	scaner := bufio.NewScanner(reader)
	totalCount := uint64(0)
	for scaner.Scan() {
		str := scaner.Text()
		totalCount += uint64(strings.Count(str, substring))
		select {
		case <-ctx.Done():
			return totalCount, fmt.Errorf("Job Cancelled")
		default:
			continue
		}
	}
	return totalCount, nil
}

func process(word string, inp WordsCountInput, queryRe *regexp.Regexp, filenameRe *regexp.Regexp, fs FileSystem, net Network) (*WordsCountResult, error) {
	filereader := io.ReadCloser(nil)
	var err error
	if queryRe.MatchString(inp.Input) {
		filereader, err = net.NetFileWithContext(inp.Ctx, inp.Input)
		if err != nil {
			return nil, err
		}
		defer filereader.Close()
	} else {
		if filenameRe.MatchString(inp.Input) {
			filereader, err = fs.FileWithContext(inp.Input)
			if err != nil {
				return nil, err
			}
			defer filereader.Close()
		}
	}
	if filereader == nil {
		return nil, fmt.Errorf("Invalid filename or url %s \n", inp.Input)
	}
	count, err := countSubstring(filereader, inp.Ctx, word)
	if err != nil {
		return nil, err
	}
	result := WordsCountResult{
		Input:  inp.Input,
		RCount: count,
	}
	return &result, nil
}

func worker(word string, fs FileSystem, net Network, input chan WordsCountInput, output chan WordsCountResult, errors chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	regexpQuery := regexp.MustCompile(urlRegexp)
	regexpFile := regexp.MustCompile(fileNameRegexp)
	for inp := range input {
		res, err := process(word, inp, regexpQuery, regexpFile, fs, net)
		if err != nil {
			errors <- err
			continue
		}
		output <- *res
	}
}
