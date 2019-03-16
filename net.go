package main

import (
	"context"
	"io"
	"net/http"
)

// Network интерфейс для работы с интернет страницами
type Network interface {
	// NetFileWithContext получить reader для страницы по адресу
	NetFileWithContext(ctx context.Context, url string) (io.ReadCloser, error)
}

// NetDriver интерфейс для работы с интернет страницами
type NetDriver struct {
	httpClient *http.Client
}

func (driver *NetDriver) NetFileWithContext(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.WithContext(ctx)
	resp, err := driver.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
