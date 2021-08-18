package kustomize

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var KFILENAME string = "Kustomization"

var _ filesys.FileSystem = overlayFileSystem{}

type overlayFileSystem struct {
	fs      filesys.FileSystem
	kfpReal string
	kfpVirt string
}

// When two kustmization_overlay data sources are defined in the same root module
// the shared file system prevents parallel execution.
// This filesys.FileSystem implementation solves this
// by handling the dynamic Kustomization in a temp directory.
func makeOverlayFS(fs filesys.FileSystem) (ofs filesys.FileSystem, tmp string, err error) {
	tmp, err = ioutil.TempDir("", "terraform-provider-kustomization-*")
	if err != nil {
		return ofs, tmp, err
	}
	kfpReal := filepath.Join(tmp, KFILENAME)

	cwd, err := os.Getwd()
	if err != nil {
		return ofs, tmp, err
	}
	kfpVirt := filepath.Join(cwd, KFILENAME)

	ofs = overlayFileSystem{
		fs:      fs,
		kfpReal: kfpReal,
		kfpVirt: kfpVirt,
	}

	return ofs, tmp, err
}

func (ofs overlayFileSystem) Create(name string) (filesys.File, error) {
	return ofs.fs.Create(name)
}

func (ofs overlayFileSystem) Mkdir(name string) error {
	return ofs.fs.Mkdir(name)
}

func (ofs overlayFileSystem) MkdirAll(name string) error {
	return ofs.fs.MkdirAll(name)
}

func (ofs overlayFileSystem) RemoveAll(name string) error {
	return ofs.fs.RemoveAll(name)
}

func (ofs overlayFileSystem) Open(name string) (filesys.File, error) {
	return ofs.fs.Open(name)
}

func (ofs overlayFileSystem) ReadDir(path string) ([]string, error) {
  return ofs.fs.ReadDir(path)
}

func (ofs overlayFileSystem) CleanedAbs(path string) (filesys.ConfirmedDir, string, error) {
	if path == ofs.kfpVirt {
		// if the path we're looking for is our virtual Kustomization file
		// fake a correct CleanedAbs response
		cd, _, err := ofs.fs.CleanedAbs(".")
		if err != nil {
			return "", "", err
		}
		return cd, KFILENAME, err
	}

	return ofs.fs.CleanedAbs(path)
}

func (ofs overlayFileSystem) Exists(name string) bool {
	ex := ofs.fs.Exists(name)

	if ex == false && name == ofs.kfpVirt {
		return ofs.fs.Exists(ofs.kfpReal)
	}

	return ex
}

func (ofs overlayFileSystem) Glob(pattern string) ([]string, error) {
	return ofs.fs.Glob(pattern)
}

func (ofs overlayFileSystem) IsDir(name string) bool {
	return ofs.fs.IsDir(name)
}

func (ofs overlayFileSystem) ReadFile(name string) ([]byte, error) {
	if name == ofs.kfpVirt {
		return ofs.fs.ReadFile(ofs.kfpReal)
	}

	return ofs.fs.ReadFile(name)
}

func (ofs overlayFileSystem) WriteFile(name string, c []byte) error {
	abs, err := filepath.Abs(name)
	if err != nil {
		return err
	}

	if abs == ofs.kfpVirt {
		return ofs.fs.WriteFile(ofs.kfpReal, c)
	}

	return ofs.fs.WriteFile(name, c)
}

func (ofs overlayFileSystem) Walk(path string, walkFn filepath.WalkFunc) error {
	return ofs.fs.Walk(path, walkFn)
}
