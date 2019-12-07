**Statics**  
The purpose of Statics is to make it dead simple to embed external files into a Go binary.  
By default, it takes all of the files in your ./include folder and embeds them as byte arrays in map called files in a separate .go file called files.go.    

**Install**  
`go get github.com/jerblack/statics`  

This will create a file called `statics[.exe]` in `GOPATH/bin`.

**Usage**  
Statics is intended to do the right thing by default so you can just run `statics` and assuming you have a folder called `./include ` with files inside, a new `./files.go` will be generated (or overwritten).  
However, the default behavior can be changed. Run `statics -h` to see the available options.  
My build script usually starts with something like `statics && go build`.

**-h Help Output**  
```
Statics is a Go tool intended to let you to easily embed static files into your compiled Go binary, 
so all of your resource files can be read directly from the binary without needing to write them to disk first.

By default, statics takes all of the files in your "./include" folder and embeds them as byte arrays 
in a map called "files" in a separate .go file called "files.go" that is part of the "main" package. 
All of this can be customized.

The following items are customizable using the switches described below:

- Output file name
- Package name
- Name of the file map
- Set one or more folders with files to import
- Store the file names with their path hierarchy preserved, 
   or flatten the path and just store the file names.
- Exclude files or subfolders from the chosen import folders by path or filename. 
- Include only specific files from the chosen import folders, 
    which will make Statics include only the specified files.
- Use wildcards for both the exclude and include folder list.
- Set build tags to enable OS and architecture specific compilation.
- Set aliases for file names so you can store the file in the file map with a different name than the actual file name.
- For all arguments that accept multiple files or folders, you can either use a pipe-separated list 
   surrounded by quotes or just set the argument multiple times, once for each file.
      -arg "item1 | item2 | item3"   -or-   -arg item1 -arg item2 -arg item3 

Be sure to re-run 'statics' after adding or modifying any files in your './include' folder.  

Usage:

  statics [-p=./include] [-out=files.go] [-pkg=main] 
   [-map=files] [-bf="// +build !windows,!darwin"] [-a "filename | alias"] [-f]
   [-x="file1 | file[1-4].* | include/img?/*png | file3"] [-i="file1 | file[1-4].* | include/img?/*png | file3"] [-v]

Flags:  
-p      Import path[s]  
        Folder path or paths with files to import relative to current working directory.  
        Specify multiple import paths with either a pipe-separated list   
        or by specifying this argument multiple times.  
            -p "dir1 | dir2 | dir3"    -or-    -p dir1 -p dir2 -p dir3  
        Files stored in map will use path starting with specified import folder.  
        (default "-p ./include")  
-o      Output go file. If go extension is not specified, it will be added.  
        (default "-o files.go")  
-pkg    package name of the go file 
        (default "-pkg main")
-map    Name of the generated files map 
        (default "-map files")
-f      Flatten path, stripping folders and just using base file names as keys in the file map.
        File will be stored as files["filename"] instead of the default files["importfolder/dirname/filename"].
-a	    Store file in the file map with a name other than it's original filename.
        Call this argument multiple times to set multiple aliases.
        The parameter should look like "original name | alias"
        Aliased files will be stored in the same folder as the original file unless alias is a path
           -a "filename1 | alias1" 
            ./importfolder/filename1 --> files["importfolder/alias1"] 
           -a "filename1 | dir/alias1" 	
            ./importfolder/filename1 --> files["dir/alias1"] 
           -f -a "filename1 | alias1" 
            ./importfolder/filename1 --> files["alias1"]
           Explicitly setting an alias with a path ignores flatten, 
            allowing you to flatten everything but the aliased file.
           -f -a "filename1 | dir/alias1"
            ./importfolder/filename1 --> files["dir/alias1"] 
-bt     Specify build tags to put at the top of the .go file.
        Can be any of the following:
            Single line
                -bt "// +build !windows,!darwin"
            Two lines joined with \n newline character
                -bt "// +build !windows,!darwin\n// +build amd64"
            Pipe-separated list
                -bt "// +build !windows,!darwin | // +build amd64"
            Same argument called multiple times
                -bt "// +build !windows,!darwin" -bt "// +build amd64"
        Go requires an additional line break after your build tags, which will be inserted automatically.
        No validation is performed and anything you specify here will be inserted at the top of the file.
-x      Specify files or folders in the import paths to exclude. 
        Can be any of the following:
            File name
                -x file1
            Path to file name beginning from import folder
                -x "importfolder/dir1/file1"
            Pipe-separated list of files or paths in import folders to exclude.
                -x "file1 | file[1-4].* | include/img?/*png | file3"
            You can also specify this argument multiple times to exclude multiple files.
                -x file1 -x "file[1-4].*" -x "include/img?/*png" -x file3
        Specifying a filename without the path will match that file anywhere in the import folder hierarchy.
        Wildcard expressions are supported. Use wildcards to exclude a whole folder: "include/folder/*"
-i      Specify files in the import paths to include. If set, only the specified files will be included.
        Can be any of the following:
            File name
                -i file1
            Path to file name beginning from import folder
                -i "importfolder/dir1/file1"
            Pipe-separated list of files or paths in import folders to exclude.
                -i "file1 | file[1-4].* | include/img?/*png | file3"
            You can also specify this argument multiple times to exclude multiple files.
                -i file1 -i "file[1-4].*" -i "include/img?/*png" -i file3
        Specifying a filename without the path will match that file anywhere in the import folder hierarchy.
        Wildcard expressions are supported. Use wildcards to include a whole folder: "include/folder/*"
-v      Verbose

Wildcards:
-x and -i both support wildcard expressions. 
Filenames with wilcards will be matched in any subfolder in the include path.
If specifying a path with wildcards, the wildcards will not include anything beyond the current file or folder name.
 The entire path starting from the import folder must be accounted for.
    To include all the pngs in the import folder and it's subfolders:
        -i "*png"
    To include all the pngs in the subfolders ending with a number from 1 to 3 in an import folder called include:
        -i "include/img[1-3]/*png"

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
