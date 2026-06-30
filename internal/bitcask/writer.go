package bitcask

import (
	"encoding/json"
	"io"
)

func (d *DataFile) Append(record Record) (int64, uint32, error) {

	offset, err := d.file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, 0, err
	}

	data, err := json.Marshal(record)
	if err != nil {
		return 0, 0, err
	}

	data = append(data, '\n')

	n, err := d.file.Write(data)
	if err != nil {
		return 0, 0, err
	}

	if err := d.file.Sync(); err != nil {
		return 0, 0, err
	}

	return offset, uint32(n), nil
}
