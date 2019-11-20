**Statics**  
The purpose of Statics is to make it dead simple to embed external files into a Go binary.  
By default, it takes all of the files in your ./include folder and embeds them as byte arrays in map called files in a separate .go file called files.go.    

**Install**  
`go get github.com/jerblack/statics`  

This will create a file called `statics[.exe]` in `GOPATH/bin`.

**Usage**  
Statics is intended to do the right thing by default so you can just run `statics` and assuming you have a folder called `./include ` with files inside, a new `./files.go` will be generated (or overwritten).  
However, the default behavior can be changed. Run `statics -h` to see the available options.  

```
By default, statics takes all of the files in your ./include folder and embeds them as byte arrays 
in a map called files in a separate .go file called files.go.

Usage:

  statics [-p=./include] [-out=files.go] [-pkg=main] [-map=files] [-k] 
  	[-x="file1 | file2 | file3"] [-i="file1 | file 2 | file3"] [-v]

Flags:

  -i string
        pipe-separated list of files in include path to include.
        Only files in include folder or subfolders with matching name will be included.
        Surround whole list with quotes like: "file1 | file2 | file3"
        Wildcard expressions are supported.
  -k    retain directory path in file names used as keys in file map.
        dirname/filename stays dirname/filename instead of just filename in the file map
  -map string
        name of the generated files map (default "files")
  -out file
        output go file (default "files.go")
  -p string
        dir path with files to embed (default "./include")
  -pkg package
        package name of the go file (default "main")
  -v    verbose
  -x string
        pipe-separated list of files in include path to exclude.
        Files in include folder or subfolders with matching name will be excluded.
        Surround whole list with quotes like: "file1 | file.* | img?/*png | file3"
        Wildcard expressions are supported.

Wildcards:

-x and -i both support wildcard expressions. 
Filenames and wilcards will be matched in any subfolder in the include path.
Matching follows the pattern defined in https://golang.org/pkg/path/filepath/#Match
pattern:
        { term }
term:
        '*'         matches any sequence of non-Separator characters
        '?'         matches any single non-Separator character
        '[' [ '^' ] { character-range } ']'
                    character class (must be non-empty)
        c           matches character c (c != '*', '?', '\\', '[')
        '\\' c      matches character c

character-range:
        c           matches character c (c != '\\', '-', ']')
        '\\' c      matches character c
        lo '-' hi   matches character c for lo <= c <= hi

```
Just be sure to re-run `statics` after modifying any of the files in your `./include` folder. My build script usually starts with something like `statics && go build`.

**Example**  
A typical generated `files.go` will look something like this.  
```go
package main

var files = map[string][]byte{

	"index.htm": []byte{108, 101, 116, ...},

	"favicon.png": []byte{137, 80, 78, ...},

	"code.js": []byte{60, 33, 68, 79, ...},

	"style.css": []byte{104, 116, 109, ...},
}

```

The example below shows how it can be used.  

```go
func main() {
    http.HandleFunc("/", serveFiles)
	log.Fatal(http.ListenAndServe(":80", nil))
}

func serveFiles(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	var fName string
	switch {
	case p == "/":
		fName = "index.html"
	case p == "/favicon.ico":
		fName = "favicon.png"
	case strings.HasPrefix(p, "/static"):
		fName = strings.TrimPrefix(p, "/static/")
	default:
		http.Error(w, http.StatusText(http.StatusNotFound),
			http.StatusNotFound)
		return
	}
	f, ok := files[fName]  // <- the files map is what is provided by statics
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	ext := filepath.Ext(fName)
	mimeType := mime.TypeByExtension(ext)
	w.Header().Set("Content-Type", mimeType)
	_, err := w.Write(f)
	if err != nil {
		fmt.Println(err.Error())
	}
}

```

**Credit**  
Statics was distilled down from https://github.com/leighmcculloch/embedfiles.   
