package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
)

const (
	defaultImportFolder = "./include"
	defaultFileName = "files.go"
	defaultPackageName = "main"
	defaultMapName = "files"
)

func main() {
	args.parse()
	importer.start()
}

func printHelp() {
	fmt.Printf(`Statics is a Go tool intended to let you to easily embed static files into your compiled Go binary, 
and all of your resource files can be read directly from the binary without needing to write them to disk first.

By default, statics takes all of the files in your "%[1]s" folder and embeds them as byte arrays 
in a map called "%[4]s" in a separate .go file called "%[2]s" that is part of the "%[3]s" package. 
All of this can be customized.

The following items are customizable using the switches described below:

	- Output file name
	- Package name
	- Name of the file map
	- Set one or more folders with files to import
	- Store the file names with their path hierarchy preserved , or to flatten the path and just store the file names.
	- Exclude files or subfolders from the chosen import folders by path or filename. 
	- Include only specific files from the chosen import folders, which will cause the tool to only 
      include the specified files.
	- Use wildcards for both the exclude and include folder list.
	- Set build tags to enable OS and architecture specific compilation.
	- Set aliases for file names so you can store the file in the file map with a different name than the actual file name.
	- For all arguments that accept multiple files or folders, you can either use a pipe-separated list 
      surrounded by quotes or just set the argument multiple times, once for each file.
		  -arg "item1 | item2 | item3"   -or-   -arg item1 -arg item2 -arg item3 

Usage:

  statics [-p=%[1]s] [-out=%[2]s] [-pkg=%[3]s] 
   [-map=%[4]s] [-bf="// +build !windows,!darwin"] [-a "filename | alias"] [-k]
   [-x="file1 | file.* | img?/*png | file3"] [-i="file1 | file.* | img?/*png | file3"] [-v]

Flags:
  -p string
        Folder path or paths with files to import relative to current working directory.
		Specify multiple import paths with either a pipe-separated list 
		or by specifying this argument multiple times.
			-p "dir1 | dir2 | dir3"    -or-    -p dir1 -p dir2 -p dir3
		Files stored in map will use path starting with specified import folder.
        (default "%[1]s")
  -o file
        Output go file. If go extension is not specified, it will be added.
        (default "%[2]s")
  -pkg package
        package name of the go file 
        (default "%[3]s")
  -map string
        Name of the generated files map 
        (default "%[4]s")
  -f    Flatten path, stripping folders and just using base file names as keys in the file map.
		File will be stored as files["filename"] instead of the default files["importfolder/dirname/filename"].
  -a	Store file in the file map with a name other than it's original filename.
        Call this argument multiple times to set multiple aliases.
        The parameter you pass to this argument will be a pipe-separated list with the first item
		being the filename or path, then pipe, then the alias.
		Aliased files will be stored in the same folder as the original file unless alias is a path
           -a "filename1 | alias1" 
			importfolder/filename1 --> files["importfolder/alias1"] 
		   -a "filename1 | dir/alias1" 	
            importfolder/filename1 --> files["dir/alias1"] 
           -f -a "filename1 | alias1" 
			importfolder/filename1 --> files["alias1"]
           Explicitly setting an alias with a path ignores flatten, allowing you to flatten everything but the aliased file.
           -f -a "filename1 | dir/alias1"
            importfolder/filename1 --> files["dir/alias1"] 
  -bf string
        Specify build flags to put at the top of the .go file.
		Can be any of the following:
			Single line
				-bf "// +build !windows,!darwin"
			Two lines joined with \n newline character
				-bf "// +build !windows,!darwin\n// +build amd64"
			Pipe-separated list
				-bf "// +build !windows,!darwin | // +build amd64"
			Same argument called multiple times
				-bf "// +build !windows,!darwin" -bf "// +build amd64"
        Additional line break is required after build flag and will be added automatically.
		No validation is performed and anything you specify here will be inserted at the top of the file.
  -x string
		Specify files or folders in the import paths to exclude. 
        Can be any of the following:
			File name
			Path to file name beginning from import folder
			Pipe-separated list of files or paths in import folders to exclude.
			You can also specify this argument multiple times to exclude multiple files.
        Files in import folders or subfolders with matching name will be excluded.
        Surround pipe-lists with quotes like: "file1 | file[1-4].* | include/img?/*png | file3"
        Wildcard expressions are supported. Use wildcards to exclude a whole folder: "include/folder/*"
  -i string
		Specify files in the include paths to include. If set, only the specified files will be included.
        Can be any of the following:
			File name
			Path to file name beginning from import folder
			Pipe-separated list of files or paths in import folder.
			You can also specify this argument multiple times to include multiple files.
        Only files in include folders or subfolders with matching name will be included.
        Surround pipe-lists with quotes like:"file1 | file[1-4].* | include/img?/*png | file3"
        Wildcard expressions are supported.
  -v    Verbose

Wildcards:
-x and -i both support wildcard expressions. Filenames and wilcards will be matched in any 
subfolder in the include path.
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
`, defaultImportFolder, defaultFileName, defaultPackageName, defaultMapName)
}

const tmpl = `
package {{.Package}}

var {{.FilesVar}} = map[string][]byte{
{{range $name, $bytes := .Files}}
	"{{$name}}": []byte{ {{range $bytes}}{{.}},{{end}} },
{{end}}
}
`
type tmplData struct {
	Package      string
	Files        map[string][]byte
	FilesVar     string
}

var args = NewArgs()
func NewArgs() Args {
	var a Args
	a.GoFile 		= defaultFileName
	a.PackageName 	= defaultPackageName
	a.MapName		= defaultMapName
	a.NameAliases	= make(map[string]string)
	return a
}

type Args struct {
	BuildFlags      []string
	Excludes   		[]string
	Includes   		[]string
	ImportPaths     []string
	GoFile          string
	FlattenPath     bool
	MapName         string
	NameAliases     map[string]string
	PackageName     string
	Verbose         bool
	Debug 			bool
}

func (a *Args) parse() {
	in := os.Args[1:]

	get := func (param string, n int) string {
		if len(in) >= n + 1 {
			return in[n]
		}
		log.Fatalf("\"%s\" requires a parameter that was not provided", param)
		return ""
	}
	re := regexp.MustCompile(`^[-\\/]+`)
	for i := 0; i < len(in); i++ {
		arg := re.ReplaceAllString(in[i], "")
		switch arg {
		case "h", "?", "help":
			printHelp()
			os.Exit(0)
		case "v", "verbose":
			args.Verbose = true
		case "dbg":
			args.Debug = true
		case "o", "out":
			o := get(arg, i+1)
			if !strings.HasSuffix(o, ".go") {
				o = o + ".go"
			}
			a.GoFile = o
			i += 1
		case "pkg":
			a.PackageName = get(arg, i+1)
			i += 1
		case "map":
			a.MapName = get(arg, i+1)
			i += 1
		case "bf":
			bfs := a.parsePipeList(get(arg, i+1))
			a.BuildFlags = append(a.BuildFlags, bfs...)
			i += 1
		case "f":
			a.FlattenPath = true
		case "p":
			paths := a.parsePipeList(get(arg, i+1))
			paths = a.cleanPaths(paths)
			a.ImportPaths = append(a.ImportPaths, paths...)
			i += 1
		case "x":
			files := a.parsePipeList(get(arg, i+1))
			files = a.cleanPaths(files)
			a.Excludes = append(a.Excludes, files...)
			i += 1
		case "i":
			files := a.parsePipeList(get(arg, i+1))
			files = a.cleanPaths(files)
			a.Includes = append(a.Includes, files...)
			i += 1
		case "a":
			alias := get(arg, i+1)
			if strings.Count(alias, "|") != 1 {
				log.Fatal("Invalid parameter for alias. Should look like -a \"<original> | <alias>\"")
			}
			pair := a.parsePipeList(alias)
			pair = a.cleanPaths(pair)
			a.NameAliases[pair[0]] = pair[1]
			i += 1
		}
	}

	a.verifyImportPaths()
}

func (a *Args) verifyImportPaths() {
	if len(a.ImportPaths) == 0 {
		a.ImportPaths = []string{defaultImportFolder}
	}
	var exit bool
	for _, imp := range a.ImportPaths {
		p, err := os.Stat(imp)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("Import path does not exist: %s", imp)
				exit = true
			}
		} else {
			if !p.IsDir() {
				fmt.Printf("Import path is not a folder: %s", imp)
				exit = true
			}
		}
	}
	if exit {
		os.Exit(1)
	}
}

func (a *Args) cleanPaths(paths []string) []string {
	var arr []string
	for _, p := range paths {
		p = filepath.ToSlash(p)
		if p != "" {
			arr = append(arr, p)
		}
	}
	return arr
}

func (a *Args) parsePipeList(ef string) []string {
	var arr []string
	if ef == "" {
		return arr
	}
	for _, x := range strings.Split(ef, "|") {
		x = strings.TrimSpace(x)
		arr = append(arr, x)
	}
	return arr
}

var importer Importer
type Importer struct {
	files map[string][]byte
}

func (i *Importer) start() {
	if args.Verbose {
		fmt.Println("[Parameters]")
		fmt.Println("Go file:", args.GoFile)
		fmt.Println("Map name:", args.MapName)
		fmt.Println("Package name:", args.PackageName)
		fmt.Println("Aliases:", args.NameAliases)
		fmt.Println("Excludes:", strings.Join(args.Excludes, " | "))
		fmt.Println("Includes:", strings.Join(args.Includes, " | "))
		fmt.Println("Import paths:", 	strings.Join(args.ImportPaths, " | "))
		fmt.Println("Build flags:", 	strings.Join(args.BuildFlags, " | "))
		fmt.Println("Flatten path:", args.FlattenPath)
	}

	i.files = make(map[string][]byte)

	for _, p := range args.ImportPaths {
		i.walkPath(p)
	}

	i.createFile()
}

func (i *Importer) createFile() {
	f, err := os.Create(args.GoFile)
	chk("creating file", err)
	if args.Verbose {
		fmt.Println("----")
	}
	if len(args.BuildFlags) > 0 {
		for _, bf := range args.BuildFlags {
			if args.Verbose {
				fmt.Println("Writing build flag: ", bf)
			}
			_, err = fmt.Fprintln(f, bf)
			chk("writing build flag", err)
		}
		_, err := fmt.Fprintln(f, "")
		chk("writing build flag", err)
		if args.Verbose {
			fmt.Println("----")
		}
	}

	t, err := template.New("").Parse(tmpl)
	chk("parsing template", err)

	buf := bytes.Buffer{}
	err = t.Execute(&buf, &tmplData{Package: args.PackageName, Files: i.files, FilesVar: args.MapName})
	chk("generating code", err)


	formatted, err := format.Source(buf.Bytes())
	chk("formatting code", err)

	_, _ = f.Write(formatted)
	err = f.Close()
	chk("finalizing file", err)


	fmt.Println("Wrote new .go file:", args.GoFile)
}

func (i *Importer) getAlias(name string) (string, bool, bool) {
	dir := filepath.ToSlash(filepath.Dir(name))
	for _, n := range []string{name, filepath.Base(name)} {
		a, ok := args.NameAliases[n]
		if ok {
			if strings.Contains(a, "/") {
				return a, ok, true
			}

			if dir != "" && !args.FlattenPath {
				return fmt.Sprintf("%s/%s", dir, a), ok, false
			}
			return a, ok, false
		}
	}
	return name, false, false
}

func (i *Importer) walkPath(importPath string) {
	var files []string
	walkFn := func(path string, f os.FileInfo, err error) error {
		if err != nil { return err }
		if f.IsDir() { return nil }
		path = filepath.ToSlash(path)

		if len(args.Includes) > 0 {
			for _, inc := range args.Includes {
				if match, _ := filepath.Match(inc, path); !match {
					return nil
				}
			}
		}

		for _, exc := range args.Excludes {
			if match, _ := filepath.Match(exc, path); match {
				return nil
			}
		}

		files = append(files, path)
		return nil
	}

	err := filepath.Walk(importPath, walkFn)
	chk("walking", err)
	i.importFiles(files)
}

func (i *Importer) importFiles(files []string) {
	for _, path := range files {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("reading file: %s", err)
		}
		if args.Verbose {
			fmt.Printf("(%d bytes)\n", len(contents))
		}


		path, found, aliasIsPath := i.getAlias(path)
		if args.Verbose && found {
			fmt.Printf("Using alias: %s\n", path)
		}

		if args.FlattenPath && !aliasIsPath {
			path = filepath.Base(path)
		}


		_, ok := i.files[path]
		if ok {
			log.Fatalf("File name conflict. Duplicate file detected: %s\n", path)
		}

		if args.Verbose {
			fmt.Printf("Adding key to %s map: %s\n", args.MapName, path)
		}
		i.files[path] = contents
	}
}

func chk(doing string, err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error %s: %s\n", doing, err)
	}
}

func dbg(d ...interface{}) {
	if args.Debug {
		caller := ""
		pc, _, _, ok := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		if ok && details != nil {
			caller = details.Name()
		}

		fmt.Printf(">>> %s.", caller)
		fmt.Println(d...)
	}
}
