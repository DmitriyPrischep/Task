package main

import (
	"io"
	"os"
)

//FileSystem интерфейс для работы с файловой системой
type FileSystem interface {
	// Получить reader файла
	FileWithContext(filename string) (io.ReadCloser, error)
}

type FileSystemDriver struct{}

func (driver *FileSystemDriver) FileWithContext(filename string) (io.ReadCloser, error) {
	file, err := os.Open(filename)
	return file, err
}
