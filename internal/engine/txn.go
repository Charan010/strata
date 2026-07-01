package engine

import "errors"

var ErrConflict = errors.New("engine: transaction conflict, retry")
var ErrTxnDone = errors.New("engine: transaction already finished")

type bufferedOp struct {
	value   string
	deleted bool
}

type Txn struct {
	engine *Engine
	buffer map[string]*bufferedOp
	order  []string
	reads  map[string]int64

	done bool
}

func (engine *Engine) Begin() *Txn {
	return &Txn{
		engine: engine,
		buffer: make(map[string]*bufferedOp),
		reads:  make(map[string]int64),
	}
}

func (txn *Txn) Put(key, value string) {

	if _, exists := txn.buffer[key]; !exists {
		txn.order = append(txn.order, key)
	}

	txn.buffer[key] = &bufferedOp{value: value}
}

func (t *Txn) Delete(key string) {
	if _, exists := t.buffer[key]; !exists {
		t.order = append(t.order, key)
	}
	t.buffer[key] = &bufferedOp{deleted: true}
}

func (t *Txn) Get(key string) (string, bool) {

	if op, buffered := t.buffer[key]; buffered {
		if op.deleted {
			return "", false
		}
		return op.value, true
	}

	value, version, ok := t.engine.getWithVersion(key)

	if _, alreadyTracked := t.reads[key]; !alreadyTracked {
		t.reads[key] = version
	}

	return value, ok
}

func (t *Txn) Rollback() error {
	if t.done {
		return ErrTxnDone
	}
	t.buffer = nil
	t.order = nil
	t.reads = nil
	t.done = true
	return nil
}

func (t *Txn) Commit() error {
	if t.done {
		return ErrTxnDone
	}
	defer func() { t.done = true }()

	t.engine.writeMu.Lock()
	defer t.engine.writeMu.Unlock()

	for key, sawVersion := range t.reads {
		if t.engine.currentVersion(key) != sawVersion {
			return ErrConflict
		}
	}

	for _, key := range t.order {
		op := t.buffer[key]
		if _, err := t.engine.putLocked(key, op.value, op.deleted); err != nil {
			return err
		}
	}

	return nil
}
