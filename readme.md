# Gozip

> NOTE: Please evaluate if [archive/zip](https://pkg.go.dev/archive/zip) is sufficient
> for your use case before using this library.

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
This makes it possible to distribute a single executable that carries its own
payload files and extracts them at runtime.

### Quick example with the gozip CLI

```bash
$ gozip
Usage of gozip:
  -c	create zip (arguments: zipfile [files...])
  -l	list zip (arguments: zipfile)
  -x	extract zip (arguments: zipfile [destination])

# make temporary copy of gozip
$ cp `which gozip` gozip

# add readme.md and LICENSE.txt as zip archive behind binary
$ gozip -c gozip readme.md LICENSE.txt

# list archive with the binary itself
$ ./gozip -l ./gozip
readme.md
LICENSE.txt
```

### Building your own self-extracting binary

A complete example lives in [`examples/selfextract/`](examples/selfextract/).
The key idea is that your program calls `gozip.Unzip(os.Executable(), dest)` to
extract files from itself:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sanderhahn/gozip"
)

func main() {
	self, _ := os.Executable()

	if !gozip.IsZip(self) {
		fmt.Println("No embedded files. Append them first:")
		fmt.Printf("  gozip -c %s <files...>\n", os.Args[0])
		os.Exit(1)
	}

	if err := gozip.Unzip(self, "extracted"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Files extracted to ./extracted")
}
```

Then build and pack:

```bash
go build -o myapp .
gozip -c myapp payload/   # append files to the binary
./myapp                   # extracts payload/ into ./extracted/
```

Run the full demo with:

```bash
cd examples/selfextract
bash demo.sh
```

### Using only the standard library

Since `gozip` appends a standard zip archive, you can also extract the embedded
files using only `archive/zip` from the Go standard library. The key is to use
`zip.NewReader` (which takes an `io.ReaderAt` and the total file size) instead
of `zip.OpenReader` — it searches backwards from the end of the file to find
the zip central directory, which works even when the zip is appended to a binary.

A complete example lives in [`examples/selfextract-stdlib/`](examples/selfextract-stdlib/):

```go
package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	self, _ := os.Executable()

	f, _ := os.Open(self)
	defer f.Close()
	info, _ := f.Stat()

	// zip.NewReader finds the archive appended to the binary
	r, err := zip.NewReader(f, info.Size())
	if err != nil {
		fmt.Println("No embedded files found.")
		os.Exit(1)
	}

	for _, zf := range r.File {
		outPath := filepath.Join("extracted", zf.Name)
		if zf.FileInfo().IsDir() {
			os.MkdirAll(outPath, zf.Mode())
			continue
		}
		os.MkdirAll(filepath.Dir(outPath), 0755)
		src, _ := zf.Open()
		dst, _ := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, zf.Mode())
		io.Copy(dst, src)
		src.Close()
		dst.Close()
	}
}
```

Run the stdlib demo with:

```bash
cd examples/selfextract-stdlib
bash demo.sh
```

## License

The source code uses the [MIT license](LICENSE.txt).

Contributors: Claude Opus 4.6, [eqawasm](https://github.com/eqawasm), [dixonwille](https://github.com/dixonwille)
