package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const tmpl = `
package {{.Package}}

var {{.FilesVar}} = map[string][]byte{
{{range $name, $bytes := .Files}}
	"{{$name}}": []byte{ {{range $bytes}}{{.}},{{end}} },
{{end}}
}
`
var include string
type tmplData struct {
	Package      string
	Files        map[string][]byte
	FilesVar     string
}

func contains(term string, sl *[]string) bool {
	term, _ = filepath.Rel(include, term)
	for _, item := range *sl {
		if term == item {
			return true
		}
		m, _ := filepath.Match(item, term)
		if m {
			return true
		}
		m, _ = filepath.Match(item, filepath.Base(term))
		if m {
			return true
		}
	}
	return false
}

func parseFileList(ef string) *[]string {
	var arr []string
	if ef == "" {
		return &arr
	}
	for _, x := range strings.Split(ef, "|") {
		x = strings.TrimSpace(x)
		arr = append(arr, x)
	}
	return &arr
}

func main() {
	out := flag.String("out", "files.go", "output go `file`")
	pkg := flag.String("pkg", "main", "`package` name of the go file")
	include = filepath.Clean(*flag.String("p", "./include", "dir path with files to embed"))
	fileMap := flag.String("map", "files", "name of the generated files map")
	verbose := flag.Bool("v", false, "verbose")
	keepDirs := flag.Bool("k", false, "retain directory path in file names used as keys in file map.\ndirname/filename stays dirname/filename instead of just filename in the file map")
	excludeFiles := flag.String("x", "", "pipe-separated list of files in include path to exclude.\n" +
		"Files in include folder or subfolders with matching name will be excluded.\nSurround whole list with quotes like: \"file1 | file.* | img?/*png | file3\"\n" +
		"Wildcard expressions are supported.")
	includeFiles := flag.String("i", "", "pipe-separated list of files in include path to include.\n" +
		"Only files in include folder or subfolders with matching name will be included.\nSurround whole list with quotes like: \"file1 | file2 | file3\"\n" +
		"Wildcard expressions are supported.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "By default, statics takes all of the files in your ./include folder and embeds them as byte arrays in a map called files in a separate .go file called files.go.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\n")
		fmt.Fprintf(os.Stderr, "  statics [-p=./include] [-out=files.go] [-pkg=main] [-map=files] [-k] [-x=\"file1 | file2 | file3\"] [-i=\"file1 | file 2 | file3\"] [-v]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `Wildcards:

-x and -i both support wildcard expressions. Filenames and wilcards will be matched in any subfolder in the include path.
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

`)
	}
	flag.Parse()

	excludes := parseFileList(*excludeFiles)
	includes := parseFileList(*includeFiles)

	f, err := os.Create(*out)
	if err != nil {
		printErr("creating file", err)
		return
	}

	if *verbose {
		fmt.Println("excludes: ", strings.Join(*excludes, " | "))
		fmt.Println("includes: ", strings.Join(*includes, " | "))
		fmt.Println("include path: ", include)
	}

	files := map[string][]byte{}
	err = filepath.Walk(include, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walking: %s", err)
		}

		if info.IsDir() {
			return nil
		}

		if *verbose {
			fmt.Println("----")
			fmt.Println("path: ", path)
		}

		if contains(path, excludes) {
			if *verbose {
				fmt.Printf("skipping file in exclude list: %s\n", path)
			}
			return nil
		}

		if len(*includes) > 0 && !contains(path, includes) {
			if *verbose {
				fmt.Printf("skipping file not in include list: %s\n", path)
			}
			return nil
		}

		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading file: %s", err)
		}
		if *verbose {
			fmt.Printf("(%d bytes)\n", len(contents))
		}

		path = filepath.ToSlash(path)
		path = strings.TrimPrefix(path, fmt.Sprintf("%s/", include)[2:])
		if !*keepDirs {
			path = filepath.Base(path)
		}
		if *verbose {
			fmt.Printf("Added key to files map: %s\n", path)
		}

		files[path] = contents
		return nil
	})
	if err != nil {
		printErr("walking", err)
		return
	}

	t, err := template.New("").Parse(tmpl)
	if err != nil {
		printErr("parsing template", err)
		return
	}

	buf := bytes.Buffer{}
	err = t.Execute(&buf, &tmplData{Package: *pkg, Files: files, FilesVar: *fileMap})
	if err != nil {
		printErr("generating code", err)
		return
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		printErr("formatting code", err)
		return
	}

	_, _ = f.Write(formatted)
	err = f.Close()
	if err != nil {
		printErr("finalizing file", err)
		return
	}
}

func printErr(doing string, err error) {
	fmt.Fprintf(os.Stderr, "Error %s: %s\n", doing, err)
}