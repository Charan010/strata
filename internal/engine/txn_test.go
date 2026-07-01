package engine

import (
	"os"
	"sync"
	"testing"
)

func newTestEngine(t *testing.T) *Engine {

	t.Helper()

	if err := os.MkdirAll("logs", 0755); err != nil {
		t.Fatalf("MkdirAll(logs) error: %v", err)
	}

	e, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() {
		e.Close()
		os.Remove(DataFilePath)
	})
	return e

}

func TestTxnCommitVisibleAfterCommit(t *testing.T) {
	e := newTestEngine(t)

	txn := e.Begin()
	txn.Put("k", "v1")

	if _, ok := e.Get("k"); ok {
		t.Fatal("uncommitted write should not be visible outside the txn")
	}

	if err := txn.Commit(); err != nil {
		t.Fatalf("Commit() error: %v", err)
	}

	v, ok := e.Get("k")
	if !ok || v != "v1" {
		t.Fatalf("got %q, %v; want v1, true", v, ok)
	}
}

func TestTxnRollbackDiscardsWrites(t *testing.T) {
	e := newTestEngine(t)
	e.Put("k", "original")

	txn := e.Begin()
	txn.Put("k", "changed")

	if err := txn.Rollback(); err != nil {
		t.Fatalf("Rollback() error: %v", err)
	}

	v, ok := e.Get("k")
	if !ok || v != "original" {
		t.Fatalf("got %q, %v; want original, true", v, ok)
	}
}

func TestTxnConflictDetected(t *testing.T) {
	e := newTestEngine(t)

	e.Put("balance", "100")

	txnA := e.Begin()
	txnB := e.Begin()

	if _, ok := txnA.Get("balance"); !ok {
		t.Fatal("txnA should see balance")
	}
	if _, ok := txnB.Get("balance"); !ok {
		t.Fatal("txnB should see balance")
	}

	txnA.Put("balance", "70")  // 100 - 30
	txnB.Put("balance", "150") // 100 + 50

	if err := txnA.Commit(); err != nil {
		t.Fatalf("txnA.Commit() error: %v", err)
	}

	err := txnB.Commit()
	if err != ErrConflict {
		t.Fatalf("txnB.Commit() error = %v; want ErrConflict", err)
	}

	// balance should reflect only A's committed write.
	v, _ := e.Get("balance")
	if v != "70" {
		t.Fatalf("balance = %q; want 70 (B's conflicting write must not apply)", v)
	}

	// retry B against the fresh value.
	txnB2 := e.Begin()
	cur, _ := txnB2.Get("balance")
	if cur != "70" {
		t.Fatalf("retry should see 70, got %q", cur)
	}
	txnB2.Put("balance", "120") // 70 + 50
	if err := txnB2.Commit(); err != nil {
		t.Fatalf("retry Commit() error: %v", err)
	}

	v, _ = e.Get("balance")
	if v != "120" {
		t.Fatalf("balance = %q; want 120 after retry", v)
	}
}

func TestTxnReadYourOwnWrites(t *testing.T) {
	e := newTestEngine(t)

	txn := e.Begin()
	txn.Put("k", "v1")

	v, ok := txn.Get("k")

	if !ok || v != "v1" {
		t.Fatalf("got %q, %v; want v1, true", v, ok)
	}
	_ = txn.Rollback()
}

func TestTxnDoubleCommitReturnsErrTxnDone(t *testing.T) {
	e := newTestEngine(t)
	txn := e.Begin()
	txn.Put("k", "v")
	if err := txn.Commit(); err != nil {
		t.Fatalf("Commit() error: %v", err)
	}
	if err := txn.Commit(); err != ErrTxnDone {
		t.Fatalf("second Commit() error = %v; want ErrTxnDone", err)
	}
}

/*
	Runs two concurrent transactions where both commit different keys the ordering could be

T1, T2 or T2, T1 . But since, the keys are disjoint no conflict should occur.
*/
func TestTxnConcurrentDisjointKeysBothCommit(t *testing.T) {
	e := newTestEngine(t)

	var wg sync.WaitGroup
	errs := make(chan error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		txn := e.Begin()
		txn.Put("a", "1")
		errs <- txn.Commit()
	}()
	go func() {

		defer wg.Done()
		txn := e.Begin()
		txn.Put("b", "2")
		errs <- txn.Commit()
	}()
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected error on disjoint-key commit: %v", err)
		}
	}
	va, _ := e.Get("a")
	vb, _ := e.Get("b")
	if va != "1" || vb != "2" {
		t.Fatalf("got a=%q b=%q; want a=1 b=2", va, vb)
	}
}
