package main

import (
	"fmt"
	"github.com/bcampbell/steno/steno/kludge"
)

func main() {

	foo, err := kludge.DataPath()
	if err != nil {
		fmt.Printf("DataPath() failed: %s\n", err)
	} else {
		fmt.Println(foo)
	}

	foo, err = kludge.PerUserPath()
	if err != nil {
		fmt.Printf("PerUserPath() failed: %s\n", err)
	} else {
		fmt.Println(foo)
	}

}
