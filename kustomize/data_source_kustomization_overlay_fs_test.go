package kustomize

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

//
//
// Test namespace attr
func TestOverlayFileSystemTwoDataSources(t *testing.T) {

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testOverlayFileSystemTwoDataSourcesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("check1", "{\"apiVersion\":\"v1\",\"kind\":\"Namespace\",\"metadata\":{\"name\":\"test-overlay1\"}}"),
					resource.TestCheckOutput("check2", "{\"apiVersion\":\"v1\",\"kind\":\"Namespace\",\"metadata\":{\"name\":\"test-overlay2\"}}"),
				),
			},
		},
	})
}

func testOverlayFileSystemTwoDataSourcesConfig() string {
	return `
data "kustomization_overlay" "test1" {
	namespace = "test-overlay1"

	resources = [
		"test_kustomizations/basic/initial",
	]
}

data "kustomization_overlay" "test2" {
	namespace = "test-overlay2"

	resources = [
		"test_kustomizations/basic/initial",
	]
}

output "check1" {
	value = data.kustomization_overlay.test1.manifests["_/Namespace/_/test-overlay1"]
}

output "check2" {
	value = data.kustomization_overlay.test2.manifests["_/Namespace/_/test-overlay2"]
}

`
}

//
//
// Unit tests
func TestOverlayFileSystemCreate(t *testing.T) {
	tmp, err := ioutil.TempDir("", "test-terraform-provider-kustomization-*")
	defer os.RemoveAll(tmp)
	assert.Equal(t, nil, err, nil)
	name := filepath.Join(tmp, "test-file")

	dfs := filesys.MakeFsOnDisk()
	ofs, otmp, err := makeOverlayFS(dfs)
	defer os.RemoveAll(otmp)
	assert.Equal(t, nil, err, nil)

	_, err = ofs.Create(name)
	defer ofs.RemoveAll(name)
	assert.Equal(t, nil, err, nil)

	existsOnDisk := dfs.Exists(name)
	assert.Equal(t, true, existsOnDisk, nil)
}

func TestOverlayFileSystemMkdir(t *testing.T) {
	tmp, err := ioutil.TempDir("", "test-terraform-provider-kustomization-*")
	defer os.RemoveAll(tmp)
	assert.Equal(t, nil, err, nil)
	name := filepath.Join(tmp, "test-mkdir")

	dfs := filesys.MakeFsOnDisk()
	ofs, otmp, err := makeOverlayFS(dfs)
	defer os.RemoveAll(otmp)
	assert.Equal(t, nil, err, nil)

	err = ofs.Mkdir(name)
	assert.Equal(t, nil, err, nil)

	existsOnDisk := dfs.Exists(name)
	assert.Equal(t, true, existsOnDisk, nil)
}

func TestOverlayFileSystemMkdirAll(t *testing.T) {
	tmp, err := ioutil.TempDir("", "test-terraform-provider-kustomization-*")
	defer os.RemoveAll(tmp)
	assert.Equal(t, nil, err, nil)

	name := filepath.Join(tmp, "test-mkdirall")

	dfs := filesys.MakeFsOnDisk()
	ofs, otmp, err := makeOverlayFS(dfs)
	defer os.RemoveAll(otmp)
	assert.Equal(t, nil, err, nil)

	err = ofs.MkdirAll(name)
	assert.Equal(t, nil, err, nil)

	existsOnDisk := dfs.Exists(name)
	assert.Equal(t, true, existsOnDisk, nil)
}

func TestOverlayFileSystemRemoveAll(t *testing.T) {
	tmp, err := ioutil.TempDir("", "test-terraform-provider-kustomization-*")
	defer os.RemoveAll(tmp)
	assert.Equal(t, nil, err, nil)

	name := filepath.Join(tmp, "test-mkdirall/test")

	dfs := filesys.MakeFsOnDisk()
	ofs, otmp, err := makeOverlayFS(dfs)
	defer os.RemoveAll(otmp)
	assert.Equal(t, nil, err, nil)

	ofs.MkdirAll(name)
	err = ofs.RemoveAll(name)
	assert.Equal(t, nil, err, nil)

	existsOnDisk := dfs.Exists(name)
	assert.Equal(t, false, existsOnDisk, nil)
}

func TestOverlayFileSystemOpen(t *testing.T) {
	tmp, err := ioutil.TempDir("", "test-terraform-provider-kustomization-*")
	defer os.RemoveAll(tmp)
	assert.Equal(t, nil, err, nil)

	name := filepath.Join(tmp, "test")

	dfs := filesys.MakeFsOnDisk()
	ofs, otmp, err := makeOverlayFS(dfs)
	defer os.RemoveAll(otmp)
	assert.Equal(t, nil, err, nil)

	ofs.Create(name)
	defer ofs.RemoveAll(name)

	_, err = ofs.Open(name)
	assert.Equal(t, nil, err, nil)

	_, err = dfs.Open(name)
	assert.Equal(t, nil, err, nil)
}

func TestOverlayFileSystemCleanedAbs(t *testing.T) {
	dfs := filesys.MakeFsOnDisk()
	ofs, otmp, err := makeOverlayFS(dfs)
	defer os.RemoveAll(otmp)
	assert.Equal(t, nil, err, nil)

	ofs.WriteFile(KFILENAME, []byte{})
	defer ofs.RemoveAll(KFILENAME)

	cwd, cwderr := os.Getwd()
	assert.Equal(t, nil, cwderr, nil)
	kfilepath := filepath.Join(cwd, KFILENAME)

	dcd, dn, derr := dfs.CleanedAbs(kfilepath)
	assert.Equal(t, filesys.ConfirmedDir(""), dcd, nil)
	assert.Equal(t, "", dn, nil)
	assert.NotEqual(t, nil, derr, nil)

	// test the overlay fakes KFILENAME into the current working dir
	ocd, on, oerr := ofs.CleanedAbs(kfilepath)
	assert.Equal(t, filesys.ConfirmedDir(cwd), ocd, nil)
	assert.Equal(t, KFILENAME, on, nil)
	assert.Equal(t, nil, oerr, nil)

	// test neither kfilepath nor KFILENAME are on disk
	_, _, ederr := dfs.CleanedAbs(kfilepath)
	assert.NotEqual(t, nil, ederr, nil)

	_, _, ederr = dfs.CleanedAbs(KFILENAME)
	assert.NotEqual(t, nil, ederr, nil)
}

func TestOverlayFileSystemExists(t *testing.T) {
	tmp, err := ioutil.TempDir("", "test-terraform-provider-kustomization-*")
	defer os.RemoveAll(tmp)
	assert.Equal(t, nil, err, nil)
	name := filepath.Join(tmp, "test")

	dfs := filesys.MakeFsOnDisk()
	ofs, otmp, err := makeOverlayFS(dfs)
	defer os.RemoveAll(otmp)
	assert.Equal(t, nil, err, nil)

	ofs.Create(name)
	defer ofs.RemoveAll(name)

	c := ofs.Exists(name)
	assert.Equal(t, true, c, nil)
}

func TestOverlayFileSystemGlob(t *testing.T) {
	tmp, err := ioutil.TempDir("", "test-terraform-provider-kustomization-*")
	defer os.RemoveAll(tmp)
	assert.Equal(t, nil, err, nil)
	name := filepath.Join(tmp, "test")

	dfs := filesys.MakeFsOnDisk()
	ofs, otmp, err := makeOverlayFS(dfs)
	defer os.RemoveAll(otmp)
	assert.Equal(t, nil, err, nil)

	ofs.Create(name)
	defer ofs.RemoveAll(name)

	r, err := ofs.Glob(filepath.Join(tmp, "*"))
	assert.Equal(t, []string{name}, r, nil)
	assert.Equal(t, nil, err, nil)
}

func TestOverlayFileSystemIsDir(t *testing.T) {
	name := "test_kustomizations"

	dfs := filesys.MakeFsOnDisk()
	ofs, otmp, err := makeOverlayFS(dfs)
	defer os.RemoveAll(otmp)
	assert.Equal(t, nil, err, nil)

	dc := dfs.IsDir(name)
	assert.Equal(t, true, dc, nil)

	oc := ofs.IsDir(name)
	assert.Equal(t, dc, oc, nil)
}

func TestOverlayFileSystemReadFile(t *testing.T) {
	dfs := filesys.MakeFsOnDisk()
	ofs, otmp, err := makeOverlayFS(dfs)
	defer os.RemoveAll(otmp)
	assert.Equal(t, nil, err, nil)

	ofs.WriteFile(KFILENAME, []byte{})

	cwd, cwderr := os.Getwd()
	assert.Equal(t, nil, cwderr, nil)
	kfilepath := filepath.Join(cwd, KFILENAME)

	// overlay the file is read with cwd as path
	od, oerr := ofs.ReadFile(kfilepath)
	assert.Equal(t, []byte{}, od, nil)
	assert.Equal(t, nil, oerr, nil)

	// the file is not on disk in cwd
	dd, derr := dfs.ReadFile(kfilepath)
	assert.Equal(t, []byte(nil), dd, nil)
	assert.NotEqual(t, nil, derr, nil)

	// the file is on disk in overlay tmp
	dd, derr = dfs.ReadFile(filepath.Join(otmp, KFILENAME))
	assert.Equal(t, []byte{}, dd, nil)
	assert.Equal(t, nil, derr, nil)
}

func TestOverlayFileSystemWriteFile(t *testing.T) {
	dfs := filesys.MakeFsOnDisk()
	ofs, otmp, err := makeOverlayFS(dfs)
	defer os.RemoveAll(otmp)
	assert.Equal(t, nil, err, nil)

	// overlay write the file without path
	ed := []byte("test")
	oerr := ofs.WriteFile(KFILENAME, ed)
	assert.Equal(t, nil, oerr, nil)

	cwd, cwderr := os.Getwd()
	assert.Equal(t, nil, cwderr, nil)
	kfilepath := filepath.Join(cwd, KFILENAME)

	// overlay read the file with cwd path
	od, oerr := ofs.ReadFile(kfilepath)
	assert.Equal(t, ed, od, nil)
	assert.Equal(t, nil, oerr, nil)

	// on-disk file with cwd does not exist
	dd, derr := dfs.ReadFile(kfilepath)
	assert.Equal(t, []byte(nil), dd, nil)
	assert.NotEqual(t, nil, derr, nil)

	// on-disk file without cwd does not exist either
	dd, derr = dfs.ReadFile(KFILENAME)
	assert.Equal(t, []byte(nil), dd, nil)
	assert.NotEqual(t, nil, derr, nil)

	// on-disk file in overlay tmp does exist
	dd, derr = dfs.ReadFile(filepath.Join(otmp, KFILENAME))
	assert.Equal(t, ed, dd, nil)
	assert.Equal(t, nil, derr, nil)
}
