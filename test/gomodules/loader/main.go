package main

import (
	"fmt"
	"log"
	"os"
	"plugin"
)

func main() {
	for _, pstr := range os.Args[1:] {
		p, err := plugin.Open(pstr)
		if err != nil {
			log.Fatal(err)
		}
		test, err := p.Lookup("Test")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(test.(func() string)())
	}
}
