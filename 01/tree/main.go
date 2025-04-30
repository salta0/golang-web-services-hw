package main

import (
	"fmt"
	"io"
	"os"
	path_pkg "path"
)

type row struct {
	name        string
	rowType     string
	baseSep     string
	sep         string
	size        int64
	printSize   bool
	printAsLast bool
}

func newRow(entry os.DirEntry, printSize bool, baseSep string, printAsLast bool) (*row, error) {
	newRow := &row{}
	newRow.printSize = printSize
	newRow.name = entry.Name()
	newRow.baseSep = baseSep
	newRow.sep = baseSep

	newRow.rowType = "file"
	if entry.IsDir() {
		newRow.rowType = "dir"
		newRow.printSize = false
	}

	fileInfo, err := entry.Info()
	if err != nil {
		return &row{}, err
	}

	newRow.size = fileInfo.Size()
	newRow.printAsLast = printAsLast

	if newRow.printAsLast {
		newRow.sep += "└───"
	} else {
		newRow.sep += "├───"
	}

	return newRow, nil
}

func (r *row) nextSeparator() string {
	nextS := r.baseSep
	if r.printAsLast {
		nextS += "\t"
	} else {
		nextS += "│\t"
	}
	return nextS
}

func (r *row) formattedSize() string {
	if r.size == 0 {
		return "empty"
	}
	return fmt.Sprintf("%vb", r.size)
}

func (r *row) print(out io.Writer) error {
	if r.rowType == "file" && !r.printSize {
		return nil
	}

	var fmtdRow string
	if r.printSize {
		fmtdRow = fmt.Sprintf("%s%s (%v)\n", r.sep, r.name, r.formattedSize())
	} else {
		fmtdRow = fmt.Sprintf("%s%s\n", r.sep, r.name)
	}

	_, err := out.Write([]byte(fmtdRow))
	if err != nil {
		return err
	}

	return nil
}

func fetchLastFileAndDir(files []os.DirEntry) (os.DirEntry, os.DirEntry) {
	lstFile := files[len(files)-1]
	var lstDir os.DirEntry
	if lstFile.IsDir() {
		lstDir = lstFile
	} else {
		for i := len(files) - 1; true; i-- {
			if files[i].IsDir() {
				lstDir = files[i]
				break
			}
			if i == 0 {
				break
			}
		}
	}
	return lstFile, lstDir
}

func printDirTree(out io.Writer, path string, printFiles bool, sep string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	lstFile, lstDir := fetchLastFileAndDir(files)

	for i := 0; i < len(files); i++ {
		printAsLast := (lstDir == files[i] && !printFiles) || (lstFile == files[i] && printFiles)

		r, err := newRow(files[i], printFiles, sep, printAsLast)
		if err != nil {
			return err
		}

		err = r.print(out)
		if err != nil {
			return err
		}

		if r.rowType == "dir" {
			printDirTree(out, path_pkg.Join(path, files[i].Name()), printFiles, r.nextSeparator())
		}
	}

	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	err := printDirTree(out, path, printFiles, "")
	if err != nil {
		return err
	}
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}

	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
