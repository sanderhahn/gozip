# Gozip

The `gozip` library can be used to add, list and extract zipped content into a
zip file or behind an executable binary. The use case for adding zip files
behind a binary is to distribute one executable that can automatically extract
required files.

```
go get -v -u github.com/sanderhahn/gozip/cmd/gozip
```

The api consist of the `Zip`, `UnzipList` and `Unzip` functions:

```go
import "github.com/sanderhahn/gozip"

// zip files/directories into file.zip (file.zip can also be an executable)
err := gozip.Zip("file.zip", []string{"content.txt", ...})

// list the zip file contents
list, err := gozip.UnzipList("file.zip")
for _, f := range list {
        fmt.Printf("%s\n", f)
}

// unzip the zip file into destination
err := gozip.Unzip("file.zip", "destination")
```

## Self Extracting Binary

The zip functions also work when the actual zip content starts behind a binary.
For example its possible to append the `readme.md` into the `gozip` command.

```bash
$ gozip
Usage of gozip:
  -c	create zip (arguments: zipfile [files...])
  -l	list zip (arguments: zipfile)
  -x	extract zip (arguments: zipfile [destination]

# make temporary copy of gozip
$ cp `which gozip` gozip

# add readme.md and UNLICENSE.txt as zip archive behind binary
$ gozip -c gozip readme.md UNLICENSE.txt

# list archive with the binary itself
$ ./gozip -l ./gozip
readme.md
UNLICENSE.txt
```

## License

The source code is [public domain](UNLICENSE.txt), feel free to reuse.

Contributors: [@dixonwille](https://github.com/dixonwille)
