package main

import (
	"fmt"
	"os"

	"github.com/AR1011/wazero/internal/gojs/testdata/argsenv"
	"github.com/AR1011/wazero/internal/gojs/testdata/crypto"
	"github.com/AR1011/wazero/internal/gojs/testdata/fs"
	"github.com/AR1011/wazero/internal/gojs/testdata/gc"
	"github.com/AR1011/wazero/internal/gojs/testdata/goroutine"
	"github.com/AR1011/wazero/internal/gojs/testdata/mem"
	"github.com/AR1011/wazero/internal/gojs/testdata/process"
	"github.com/AR1011/wazero/internal/gojs/testdata/stdio"
	"github.com/AR1011/wazero/internal/gojs/testdata/testfs"
	"github.com/AR1011/wazero/internal/gojs/testdata/time"
	"github.com/AR1011/wazero/internal/gojs/testdata/writefs"
)

// main includes a registry of all tests to reduce compilation time.
func main() {
	switch os.Args[1] {
	case "argsenv":
		argsenv.Main()
	case "crypto":
		crypto.Main()
	case "exit":
		os.Exit(255)
	case "fs":
		fs.Main()
	case "gc":
		gc.Main()
	case "goroutine":
		goroutine.Main()
	case "mem":
		mem.Main()
	case "process":
		process.Main()
	case "stdio":
		stdio.Main()
	case "testfs":
		testfs.Main()
	case "time":
		time.Main()
	case "writefs":
		writefs.Main()
	default:
		panic(fmt.Errorf("unsupported arg: %s", os.Args[1]))
	}
}
