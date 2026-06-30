package bitcask

import (
	"os"
)

type DataFile struct {
	file *os.File
	path string
}

func New(path string) (*DataFile, error) {

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}

	return &DataFile{
		file: file,
		path: path,
	}, nil
}

func (d *DataFile) Close() error {
	return d.file.Close()
}
