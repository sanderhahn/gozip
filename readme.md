# Gozip Library

The `gozip` library can be used to add, list and extract zipped content into a zip file or behind an executable binary. The use case for adding zip files behind a binary is to distribute one executable that can automatically extract required files.

```
go get -v -u github.com/sanderhahn/gozip
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

## License

The source code is [public domain](UNLICENSE.txt), feel free to reuse.

Contributors: [@dixonwille](https://github.com/dixonwille)
