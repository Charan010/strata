package engine

import (
	"sync"
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

	/* Whenever data records needs to be inserted into bitcask file, this mutex is acquired for single record append
	or when a transaction is commited.
	*/
	writeMu sync.Mutex
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
	e.writeMu.Lock()
	defer e.writeMu.Unlock()

	_, err := e.putLocked(key, value, false)
	return err
}

func (e *Engine) Delete(key string) error {
	e.writeMu.Lock()
	defer e.writeMu.Unlock()

	_, err := e.putLocked(key, "", true)
	return err
}

func (e *Engine) putLocked(key, value string, deleted bool) (int64, error) {

	record := bitcask.Record{
		Key:       key,
		Value:     value,
		Timestamp: time.Now().UnixMilli(),
		Deleted:   deleted,
	}

	offset, size, err := e.dataFile.Append(record)
	if err != nil {
		return 0, err
	}

	e.keydir.Put(key, keydir.Entry{
		Offset:    offset,
		Size:      size,
		Timestamp: record.Timestamp,
		Deleted:   deleted,
	})

	return record.Timestamp, nil
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

func (e *Engine) getWithVersion(key string) (value string, version int64, ok bool) {

	entry, found := e.keydir.Get(key)
	if !found || entry.Deleted {
		return "", 0, false
	}

	record, err := e.dataFile.ReadAt(entry.Offset)
	if err != nil || record.Deleted {
		return "", 0, false
	}

	return record.Value, entry.Timestamp, true
}

func (e *Engine) currentVersion(key string) int64 {
	entry, ok := e.keydir.Get(key)
	if !ok || entry.Deleted {
		return 0
	}
	return entry.Timestamp
}

func (e *Engine) Close() error {
	return e.dataFile.Close()
}
