package engine

import (
	"time"

	"github.com/Charan010/strata/internal/bitcask"
	"github.com/Charan010/strata/internal/keydir"
)

const (
	DataFilePath = "logs/master.log"
)

type Engine struct {
	keydir   *keydir.KeyDir
	dataFile *bitcask.DataFile
}

func New() (*Engine, error) {

	kd := keydir.New()

	df, err := bitcask.New(DataFilePath)
	if err != nil {
		return nil, err
	}

	err = df.Replay(func(record *bitcask.Record, offset int64, size uint32) error {

		kd.Put(record.Key, keydir.Entry{
			Offset:    offset,
			Size:      size,
			Timestamp: record.Timestamp,
			Deleted:   record.Deleted,
		})

		return nil
	})

	if err != nil {
		df.Close()
		return nil, err
	}

	return &Engine{
		keydir:   kd,
		dataFile: df,
	}, nil
}

func (e *Engine) Put(key, value string) error {

	record := bitcask.Record{
		Key:       key,
		Value:     value,
		Timestamp: time.Now().UnixMilli(),
		Deleted:   false,
	}

	offset, size, err := e.dataFile.Append(record)
	if err != nil {
		return err
	}

	e.keydir.Put(key, keydir.Entry{
		Offset:    offset,
		Size:      size,
		Timestamp: record.Timestamp,
	})

	return nil
}

func (e *Engine) Get(key string) (string, bool) {

	entry, ok := e.keydir.Get(key)
	if !ok || entry.Deleted {
		return "", false
	}

	record, err := e.dataFile.ReadAt(entry.Offset)
	if err != nil {
		return "", false
	}

	if record.Deleted {
		return "", false
	}

	return record.Value, true
}

func (e *Engine) Close() error {
	return e.dataFile.Close()
}
