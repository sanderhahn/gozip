package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/sanderhahn/gozip"
)

func main() {
	var list, extract, create bool
	flag.BoolVar(&create, "c", false, "create zip (arguments: zipfile [files...])")
	flag.BoolVar(&list, "l", false, "list zip (arguments: zipfile)")
	flag.BoolVar(&extract, "x", false, "extract zip (arguments: zipfile [destination]")

	flag.Parse()

	args := flag.Args()
	argc := len(args)
	if list && argc == 1 {
		path := args[0]
		list, err := gozip.UnzipList(path)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range list {
			fmt.Printf("%s\n", f)
		}
	} else if extract && (argc == 1 || argc == 2) {
		path := args[0]
		dest := "."
		if argc == 2 {
			dest = args[1]
		}
		err := gozip.Unzip(path, dest)
		if err != nil {
			log.Fatal(err)
		}
	} else if create && argc > 1 {
		err := gozip.Zip(args[0], args[1:])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		flag.Usage()
	}
}
