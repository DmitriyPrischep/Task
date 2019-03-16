package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
)

type FileSystemDriverMock struct{}

func (driver *FileSystemDriverMock) FileWithContext(filename string) (io.ReadCloser, error) {
	m := make(map[string]string)
	m["/test/testing"] =
		`An Introduction to Programming in Go
		A short, concise introduction to computer programming using the language Go. 
		Designed by Google, Go is a general purpose programming language with modern features, 
		clean syntax and a robust well-documented common library, making it an ideal language to learn as your first programming language.`

	m["/root/mailru"] =
		`A 4-week instructional series that covers the material in 
		An Introduction to Programming in Go as well as the basics of server-side 
		web development and Google App Engine.`

	m["/var/mytarget"] =
		`Your question may have been answered already. Before asking, 
		please search the group archive and check the Go Frequently Asked Questions page. 
		Also see the Go documentation index.  When asking a question, please describe 
		the system and the version of Go you are using.  Please say exactly what you did, 
		exactly what happened, and what you expected to happen instead.  Please avoid sending 
		screenshots if at all possible; send text, not a screenshot.`

	m["/usr/dmitry"] =
		`A sense of community flourishes when we come together in person. 
		As handles become names and avatars become faces, the smiles are real and true friendship can grow. 
		There is joy in the sharing of knowledge and celebrating the accomplishments of our friends, 
		colleagues, and neighbors. In our rapidly growing Go community this critical role is played by the Go user groups.`

	var err error
	var text string
	var exist bool
	if text, exist = m[filename]; !exist {
		err = errors.New("Text not found")
		return nil, err
	}
	r := ioutil.NopCloser(bytes.NewReader([]byte(text)))
	return r, err
}

type NetDriverMock struct {
	httpClient *http.Client
}

func (driver *NetDriverMock) NetFileWithContext(ctx context.Context, url string) (io.ReadCloser, error) {
	m := make(map[string]string)
	m["http://mysite.net"] =
		`An Introduction to Programming in Go
		A short, concise introduction to computer programming using the language Go. 
		Designed by Google, Go is a general purpose programming language with modern features, 
		clean syntax and a robust well-documented common library, making it an ideal language to learn as your first programming language.`

	m["http://yandex.ru/Golang"] =
		`A 4-week instructional series that covers the material in 
		An Introduction to Programming in Go as well as the basics of server-side 
		web development and Google App Engine.`

	m["ftp://work.in/mairu"] =
		`Your question may have been answered already. Before asking, 
		please search the group archive and check the Go Frequently Asked Questions page. 
		Also see the Go documentation index.  When asking a question, please describe 
		the system and the version of Go you are using.  Please say exactly what you did, 
		exactly what happened, and what you expected to happen instead.  Please avoid sending 
		screenshots if at all possible; send text, not a screenshot.`

	m["https://example.com"] =
		`A sense of community flourishes when we come together in person. 
		As handles become names and avatars become faces, the smiles are real and true friendship can grow. 
		There is joy in the sharing of knowledge and celebrating the accomplishments of our friends, 
		colleagues, and neighbors. In our rapidly growing Go community this critical role is played by the Go user groups.`

	var err error
	var text string
	var exist bool
	if text, exist = m[url]; !exist {
		err = errors.New("Text not found")
		return nil, err
	}
	r := ioutil.NopCloser(bytes.NewReader([]byte(text)))
	return r, err
}

var countTest = []struct {
	input    string
	expected WordsCountResult
}{
	{`/test/testing`, WordsCountResult{Input: `/test/testing`, RCount: 4}},
	{`/root/mailru`, WordsCountResult{Input: `/root/mailru`, RCount: 2}},
	{`/var/mytarget`, WordsCountResult{Input: `/var/mytarget`, RCount: 3}},
	{`/usr/dmitry`, WordsCountResult{Input: `/usr/dmitry`, RCount: 2}},
	{`http://mysite.net`, WordsCountResult{Input: `http://mysite.net`, RCount: 4}},
	{`http://yandex.ru/Golang`, WordsCountResult{Input: `http://yandex.ru/Golang`, RCount: 2}},
	{`ftp://work.in/mairu`, WordsCountResult{Input: `ftp://work.in/mairu`, RCount: 3}},
	{`https://example.com`, WordsCountResult{Input: `https://example.com`, RCount: 2}},
}

func TestNewWordsCounter(t *testing.T) {
	var fsDriver FileSystemDriverMock
	var netDriver NetDriverMock
	count := uint8(5)
	for _, test := range countTest {
		input, output, errors := NewWordsCounter("Go", &fsDriver, &netDriver, count)
		var wg sync.WaitGroup
		wg.Add(1)
		var result WordsCountResult
		go func() {
			defer wg.Done()
			for value := range output {
				result = value
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for value := range errors {
				fmt.Print(value.Error())
			}
		}()

		line := test.input
		inpTask := WordsCountInput{
			Input: line,
			Ctx:   context.Background(),
		}
		input <- inpTask

		close(input)
		wg.Wait()

		if result != test.expected {
			t.Errorf("NewWordsCounter(%s): expected %v, actual %v", test.input, test, result)
		}
	}
}
