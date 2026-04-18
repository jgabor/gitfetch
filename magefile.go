//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
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
	// Remove any prior binary so `go build` always writes a fresh one.
	// Go's linker skips the write when the target already contains a
	// matching Go BuildID — which a UPX-packed binary still carries,
	// so subsequent builds would leave the packed file untouched and
	// UPX would then fail with AlreadyPackedException.
	if err := os.Remove(out); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := sh.RunWith(map[string]string{"CGO_ENABLED": "0"},
		"go", "build",
		"-trimpath",
		"-ldflags=-s -w",
		"-o", out,
		pkgPath,
	); err != nil {
		return err
	}
	return compress(out)
}

// Install builds, compresses, and installs gitfetch into GOBIN.
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

// compress runs UPX --lzma on the built binary. Opt-in via GITFETCH_COMPRESS=1
// because UPX-LZMA decompression adds ~58 ms to every exec — unacceptable for
// a fetch-style CLI that must feel instant in .bashrc. The 3 MB size saving is
// not worth the 60× startup cost; users who care about binary size can opt in.
func compress(path string) error {
	if os.Getenv("GITFETCH_COMPRESS") != "1" {
		return nil
	}
	upx, err := exec.LookPath("upx")
	if err != nil {
		fmt.Println("upx not found; skipping compression")
		return nil
	}
	fmt.Printf("compressing %s with upx\n", path)
	return sh.Run(upx, "--best", "--overlay=strip", "--lzma", path)
}
