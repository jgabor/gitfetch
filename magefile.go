//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binName = "gitfetch"
	pkgPath = "./cmd/gitfetch"
)

func binPath() string {
	name := binName
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join("bin", name)
}

func installDir() string {
	if dir := os.Getenv("GOBIN"); dir != "" {
		return dir
	}
	if dir, err := sh.Output("go", "env", "GOBIN"); err == nil && dir != "" {
		return dir
	}
	gopath, err := sh.Output("go", "env", "GOPATH")
	if err != nil || gopath == "" {
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}
	return filepath.Join(gopath, "bin")
}

// Build compiles gitfetch into ./bin with size-optimized flags.
func Build() error {
	if err := os.MkdirAll("bin", 0o755); err != nil {
		return err
	}
	out := binPath()
	if err := os.Remove(out); err != nil && !os.IsNotExist(err) {
		return err
	}
	return sh.RunWith(map[string]string{"CGO_ENABLED": "0"},
		"go", "build",
		"-trimpath",
		"-ldflags=-s -w",
		"-o", out,
		pkgPath,
	)
}

// Install builds and installs gitfetch into GOBIN.
func Install() error {
	mg.Deps(Build)
	dest := installDir()
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	target := filepath.Join(dest, filepath.Base(binPath()))
	if err := sh.Copy(target, binPath()); err != nil {
		return err
	}
	fmt.Printf("installed %s\n", target)
	return nil
}

// Clean removes build artifacts.
func Clean() error {
	return os.RemoveAll("bin")
}


