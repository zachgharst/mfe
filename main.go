package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		help()
		return
	}

	dir := os.Args[1]
	files, err := getAllFiles(dir)
	if err != nil {
		fmt.Println(err)
		return
	}

	tempFile, err := createTempFile(files)
	defer os.Remove(tempFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	openEditorForModifications(tempFile)

	modifications, err := calculateModifications(files, tempFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	makeModifications(files, modifications)
}

func help() {
	fmt.Printf("%s: missing directory operand\n", os.Args[0])
	fmt.Printf("Usage: %s <directory>\n", os.Args[0])
}

// openEditorForModifications opens the file in the editor specified by envvar
// $EDITOR (fallback: vim)
func openEditorForModifications(file string) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// getAllFilesRecursively returns a list of all files in the directory and its
// subdirectories; each is prepended with the default command "move"
func getAllFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			files = append(files, "move "+path)
		}

		return nil
	})

	return files, err
}

// createTempFile creates a temporary file with the specified lines as content
func createTempFile(lines []string) (string, error) {
	tempFile, err := os.CreateTemp("", "mfe-")
	if err != nil {
		return "", err
	}

	_, err = tempFile.Write([]byte(strings.Join(lines, "\n")))
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

// calculateModifications checks for problems in the modifications file and
// returns the lines
func calculateModifications(files []string, tempFile string) ([]string, error) {
	modifications, err := os.ReadFile(tempFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(modifications), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) != len(files) {
		fmt.Println(len(lines), len(files))
		fmt.Println("Do not add or remove lines from the file")
		return nil, errors.New("Do not add or remove lines from the file")
	}

	return lines, nil
}

// makeModifications compares the diff; then, it moves or deletes the files as
// specified
func makeModifications(before []string, after []string) {
	for i := range after {
		if after[i] != before[i] {
			_, beforeFile, _ := strings.Cut(before[i], " ")
			afterVerb, afterFile, _ := strings.Cut(after[i], " ")

			if afterVerb == "move" || afterVerb == "m" {
				fmt.Printf("Moving %s to %s\n", beforeFile, afterFile)
				os.MkdirAll(filepath.Dir(afterFile), 0755)
				os.Rename(beforeFile, afterFile)
			} else if afterVerb == "delete" || afterVerb == "d" {
				fmt.Printf("Deleting %s\n", afterFile)
				os.Remove(afterFile)
			} else {
				fmt.Printf("Unknown command %s for %s\n", afterVerb, afterFile)
			}
		}
	}
}
