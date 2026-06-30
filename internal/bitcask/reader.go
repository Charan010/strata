package bitcask

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
)

func (d *DataFile) Replay(fn func(*Record, int64, uint32) error) error {

	file, err := os.Open(d.path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var offset int64 = 0

	for scanner.Scan() {

		line := scanner.Bytes()

		var record Record

		if err := json.Unmarshal(line, &record); err != nil {
			return err
		}

		size := uint32(len(line) + 1)

		if err := fn(&record, offset, size); err != nil {
			return err
		}

		offset += int64(size)
	}

	return scanner.Err()
}

func (d *DataFile) ReadAt(offset int64) (*Record, error) {

	if _, err := d.file.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(d.file)

	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var record Record

	if err := json.Unmarshal(line, &record); err != nil {
		return nil, err
	}

	return &record, nil
}
