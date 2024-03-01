package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
    if (len(os.Args) <  2) {
        fmt.Println("mass_file_editor: missing directory operand")
        fmt.Println("Usage: mass_file_editor <directory>")
        return
    }

    dir := os.Args[1]
    files, err := getAllFiles(dir)
    if err != nil {
        fmt.Println(err)
        return
    }
    os.WriteFile("files", []byte(strings.Join(files, "\n")), 0644)

    EDITOR := os.Getenv("EDITOR")
    if EDITOR == "" {
        EDITOR = "vim"
    }

    editor := exec.Command(EDITOR, "files")
    editor.Stdin = os.Stdin
    editor.Stdout = os.Stdout

    editor.Run()

    modifications, err := os.ReadFile("files")
    if err != nil {
        fmt.Println(err)
        return
    }

    lines := strings.Split(string(modifications), "\n")
    if lines[len(lines) - 1] == "" {
        lines = lines[:len(lines) - 1]
    }

    if len(lines) != len(files) {
        fmt.Println(len(lines), len(files))
        fmt.Println("Do not add or remove lines from the file")
        return
    }   

    for i := range lines {
        if lines[i] != files[i] {
            _, beforeFile, _ := strings.Cut(files[i], " ")
            afterVerb, afterFile, _ := strings.Cut(lines[i], " ")

            if afterVerb == "move" ||afterVerb == "m" {
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

func getAllFiles(dir string) ([]string, error) {
    var files []string

    err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        if !d.IsDir() {
            files = append(files, "move " + path)
        }

        return nil
    })

    return files, err
}

