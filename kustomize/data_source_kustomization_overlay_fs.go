package kustomize

import (
	"fmt"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/filesys"
)

var KFILENAME string = "Kustomization"

var _ filesys.FileSystem = overlayFileSystem{}

type overlayFileSystem struct {
	upper filesys.FileSystem
	lower filesys.FileSystem
}

// When two kustmization_overlay data sources are defined in the same root module the shared
// file system prevents parallel execution.
// This filesys.FileSystem implementation solves this by handling the dynamic Kustomization in memory.
func makeOverlayFS(upper filesys.FileSystem, lower filesys.FileSystem) filesys.FileSystem {
	return overlayFileSystem{
		upper: upper,
		lower: lower,
	}
}

func (ofs overlayFileSystem) Create(name string) (filesys.File, error) {
	return ofs.upper.Create(name)
}

func (ofs overlayFileSystem) Mkdir(name string) error {
	return ofs.upper.Mkdir(name)
}

func (ofs overlayFileSystem) MkdirAll(name string) error {
	return ofs.upper.MkdirAll(name)
}

func (ofs overlayFileSystem) RemoveAll(name string) error {
	return ofs.upper.RemoveAll(name)
}

func (ofs overlayFileSystem) Open(name string) (filesys.File, error) {
	of, err := ofs.lower.Open(name)

	if err != nil {
		return ofs.upper.Open(name)
	}

	return of, err
}

func (ofs overlayFileSystem) CleanedAbs(path string) (filesys.ConfirmedDir, string, error) {
	cd, n, err := ofs.lower.CleanedAbs(path)

	if err != nil && strings.HasSuffix(path, KFILENAME) {
		cd, _, err = ofs.lower.CleanedAbs(".")
		return cd, KFILENAME, err
	}

	return cd, n, err
}

func (ofs overlayFileSystem) Exists(name string) bool {
	onDisk := ofs.lower.Exists(name)

	if onDisk == false {
		return ofs.upper.Exists(name)
	}

	return onDisk
}

func (ofs overlayFileSystem) Glob(pattern string) ([]string, error) {
	lfs, err := ofs.lower.Glob(pattern)
	if err != nil {
		return lfs, err
	}

	// Glob only errors if the pattern is invalid
	ufs, _ := ofs.lower.Glob(pattern)

	return append(lfs, ufs...), nil

}

func (ofs overlayFileSystem) IsDir(name string) bool {
	exl := ofs.lower.IsDir(name)

	if exl == false {
		return ofs.upper.IsDir(name)
	}

	return exl
}

func (ofs overlayFileSystem) ReadFile(name string) ([]byte, error) {
	d, err := ofs.lower.ReadFile(name)

	if err != nil && strings.HasSuffix(name, KFILENAME) {
		return ofs.upper.ReadFile(KFILENAME)
	}

	return d, err
}

func (ofs overlayFileSystem) WriteFile(name string, c []byte) error {
	if ofs.lower.Exists(name) {
		return fmt.Errorf("OverlayFS: %q already exists in lower FS.", name)
	}

	return ofs.upper.WriteFile(name, c)
}

func (ofs overlayFileSystem) Walk(path string, walkFn filepath.WalkFunc) error {
	return ofs.lower.Walk(path, walkFn)
}
