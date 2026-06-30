package main

import (
	"fmt"

	"github.com/Charan010/strata/internal/engine"
)

func main() {

	db, err := engine.New()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Put("charan", "123")
	v, ok := db.Get("charan")
	fmt.Println(v, ok)
}
