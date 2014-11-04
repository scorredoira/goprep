package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
)

// goprep is a tool to comment or uncomment code based on #if and
// #endif directives.
// For example:
//
//       //#if debug
//       println("I'm debugging!")
//       //#endif
//
// $ goprep debug will uncomment println.
// $ goprep will comment it back
//
// It calls goimports to update the imported packages that may be
// needed by the uncommented code.

var args []string

const (
	preprocessorStart = "//#if "
	preprocessorEnd   = "//#endif"
	commentStart      = "//"
)

func main() {
	args = os.Args[1:]
	processWorkingDir()
}

func processWorkingDir() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	processDir(dir)
}

func processDir(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		err = processFile(f)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func processFile(file os.FileInfo) error {
	name := file.Name()
	if strings.HasSuffix(name, ".go") {
		buf, changed, err := processLines(name)
		if err != nil {
			return err
		}

		if changed {
			writeFile(file, buf)
		}
	}
	return nil
}

func processLines(path string) (*bytes.Buffer, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()

	var buffer bytes.Buffer
	var f = read
	var b string
	var changed bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		a := scanner.Text()
		b, f = f(a)
		buffer.WriteString(b)
		buffer.WriteByte('\n')
		if b != a {
			changed = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, false, err
	}

	return &buffer, changed, nil
}

// the function that processes lines returns the
// function for the next line to avoid a large
// switch to constantly recalculate the state.
type lineFunc func(line string) (string, lineFunc)

func read(line string) (string, lineFunc) {
	p := whitePreffix(line)
	if strings.HasPrefix(line[p:], preprocessorStart) {
		// uncomment if the argument has been invoked, else comment.
		if hasArgument(line[p+len(preprocessorStart):]) {
			return line, unComment
		}
		return line, comment
	}
	return line, read
}

func comment(line string) (string, lineFunc) {
	p := whitePreffix(line)
	if strings.HasPrefix(line[p:], preprocessorEnd) {
		return line, read
	}

	if !strings.HasPrefix(line[p:], commentStart) {
		line = line[:p] + commentStart + line[p:]
	}
	return line, comment
}

func unComment(line string) (string, lineFunc) {
	p := whitePreffix(line)
	trimmed := line[p:]
	if strings.HasPrefix(trimmed, preprocessorEnd) {
		return line, read
	}

	if strings.HasPrefix(trimmed, commentStart) {
		line = line[:p] + trimmed[2:]
	}
	return line, unComment
}

func writeFile(file os.FileInfo, buf *bytes.Buffer) {
	name := file.Name()
	ioutil.WriteFile(name, buf.Bytes(), file.Mode())

	// Use goimports to include any package needed by the
	// uncommented code
	syscall.Exec("goimports", []string{"-w", name}, os.Environ())
}

func whitePreffix(line string) int {
	for i, c := range line {
		if c != '\t' && c != ' ' {
			return i
		}
	}
	return len(line)
}

func hasArgument(arg string) bool {
	for _, l := range args {
		if arg == l {
			return true
		}
	}
	return false
}
