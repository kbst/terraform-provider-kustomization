package kustomize

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filesys"
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
					resource.TestCheckOutput("check1", "{\"apiVersion\":\"v1\",\"kind\":\"Namespace\",\"metadata\":{\"name\":\"test-overlay1\"}}\n"),
					resource.TestCheckOutput("check2", "{\"apiVersion\":\"v1\",\"kind\":\"Namespace\",\"metadata\":{\"name\":\"test-overlay2\"}}\n"),
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
	value = data.kustomization_overlay.test1.manifests["~G_v1_Namespace|~X|test-overlay1"]
}

output "check2" {
	value = data.kustomization_overlay.test2.manifests["~G_v1_Namespace|~X|test-overlay2"]
}

`
}

//
//
// Unit tests
func TestOverlayFileSystemCreate(t *testing.T) {
	mfs := filesys.MakeFsInMemory()
	dfs := filesys.MakeFsOnDisk()

	ofs := makeOverlayFS(mfs, dfs)

	_, err := ofs.Create(KFILENAME)
	defer ofs.RemoveAll(KFILENAME)
	assert.Equal(t, nil, err, nil)

	existsOnDisk := dfs.Exists(KFILENAME)
	assert.Equal(t, false, existsOnDisk, nil)

	existsInMem := mfs.Exists(KFILENAME)
	assert.Equal(t, true, existsInMem, nil)
}

func TestOverlayFileSystemMkdir(t *testing.T) {
	name := "test-mkdir"
	mfs := filesys.MakeFsInMemory()
	dfs := filesys.MakeFsOnDisk()

	ofs := makeOverlayFS(mfs, dfs)

	err := ofs.Mkdir(name)
	defer ofs.RemoveAll(name)
	assert.Equal(t, nil, err, nil)

	existsOnDisk := dfs.Exists(name)
	assert.Equal(t, false, existsOnDisk, nil)

	existsInMem := mfs.Exists(name)
	assert.Equal(t, true, existsInMem, nil)
}

func TestOverlayFileSystemMkdirAll(t *testing.T) {
	name := "test-mkdirall/test"
	mfs := filesys.MakeFsInMemory()
	dfs := filesys.MakeFsOnDisk()

	ofs := makeOverlayFS(mfs, dfs)

	err := ofs.MkdirAll(name)
	defer ofs.RemoveAll(name)
	assert.Equal(t, nil, err, nil)

	existsOnDisk := dfs.Exists(name)
	assert.Equal(t, false, existsOnDisk, nil)

	existsInMem := mfs.Exists(name)
	assert.Equal(t, true, existsInMem, nil)
}

func TestOverlayFileSystemRemoveAll(t *testing.T) {
	name := "test-mkdirall/test"
	mfs := filesys.MakeFsInMemory()
	dfs := filesys.MakeFsOnDisk()

	ofs := makeOverlayFS(mfs, dfs)

	ofs.MkdirAll(name)
	err := ofs.RemoveAll(name)
	assert.Equal(t, nil, err, nil)

	existsOnDisk := dfs.Exists(name)
	assert.Equal(t, false, existsOnDisk, nil)

	existsInMem := mfs.Exists(name)
	assert.Equal(t, false, existsInMem, nil)
}

func TestOverlayFileSystemOpen(t *testing.T) {
	mfs := filesys.MakeFsInMemory()
	dfs := filesys.MakeFsOnDisk()

	ofs := makeOverlayFS(mfs, dfs)

	ofs.Create(KFILENAME)
	defer ofs.RemoveAll(KFILENAME)

	_, err := ofs.Open(KFILENAME)
	assert.Equal(t, nil, err, nil)

	_, err = mfs.Open(KFILENAME)
	assert.Equal(t, nil, err, nil)

	_, err = dfs.Open(KFILENAME)
	assert.NotEqual(t, nil, err, nil)
}

func TestOverlayFileSystemCleanedAbs(t *testing.T) {
	mfs := filesys.MakeFsInMemory()
	dfs := filesys.MakeFsOnDisk()

	ofs := makeOverlayFS(mfs, dfs)

	ofs.Create(KFILENAME)
	defer ofs.RemoveAll(KFILENAME)

	dcd, dn, derr := dfs.CleanedAbs(".")
	assert.Equal(t, "", dn, nil)
	assert.Equal(t, nil, derr, nil)

	ocd, on, oerr := ofs.CleanedAbs(KFILENAME)
	assert.Equal(t, dcd, ocd, nil)
	assert.Equal(t, KFILENAME, on, nil)
	assert.Equal(t, nil, oerr, nil)

	mcd, mn, merr := mfs.CleanedAbs(KFILENAME)
	assert.Equal(t, filesys.ConfirmedDir("/"), mcd, nil)
	assert.Equal(t, KFILENAME, mn, nil)
	assert.Equal(t, nil, merr, nil)

	_, _, ederr := dfs.CleanedAbs(KFILENAME)
	assert.NotEqual(t, nil, ederr, nil)
}

func TestOverlayFileSystemExists(t *testing.T) {
	mfs := filesys.MakeFsInMemory()
	dfs := filesys.MakeFsOnDisk()

	ofs := makeOverlayFS(mfs, dfs)

	ofs.Create(KFILENAME)
	defer ofs.RemoveAll(KFILENAME)

	c := ofs.Exists(KFILENAME)
	assert.Equal(t, true, c, nil)
}

func TestOverlayFileSystemGlob(t *testing.T) {
	mfs := filesys.MakeFsInMemory()
	dfs := filesys.MakeFsOnDisk()

	ofs := makeOverlayFS(mfs, dfs)

	ofs.Create(KFILENAME)
	defer ofs.RemoveAll(KFILENAME)

	r, err := ofs.Glob(KFILENAME)
	assert.Equal(t, []string(nil), r, nil)
	assert.Equal(t, nil, err, nil)
}

func TestOverlayFileSystemIsDir(t *testing.T) {
	name := "test_kustomizations"
	mfs := filesys.MakeFsInMemory()
	dfs := filesys.MakeFsOnDisk()

	ofs := makeOverlayFS(mfs, dfs)

	mc := mfs.IsDir(name)
	assert.Equal(t, false, mc, nil)

	dc := dfs.IsDir(name)
	assert.Equal(t, true, dc, nil)

	oc := ofs.IsDir(name)
	assert.Equal(t, dc, oc, nil)
}

func TestOverlayFileSystemReadFile(t *testing.T) {
	mfs := filesys.MakeFsInMemory()
	dfs := filesys.MakeFsOnDisk()

	ofs := makeOverlayFS(mfs, dfs)
	ofs.Create(KFILENAME)
	defer ofs.RemoveAll(KFILENAME)

	od, oerr := ofs.ReadFile(KFILENAME)
	assert.Equal(t, []byte{}, od, nil)
	assert.Equal(t, nil, oerr, nil)

	md, merr := mfs.ReadFile(KFILENAME)
	assert.Equal(t, []byte{}, md, nil)
	assert.Equal(t, nil, merr, nil)

	dd, derr := dfs.ReadFile(KFILENAME)
	assert.Equal(t, []byte(nil), dd, nil)
	assert.NotEqual(t, nil, derr, nil)
}

func TestOverlayFileSystemWriteFile(t *testing.T) {
	mfs := filesys.MakeFsInMemory()
	dfs := filesys.MakeFsOnDisk()

	ed := []byte("test")

	ofs := makeOverlayFS(mfs, dfs)
	oerr := ofs.WriteFile(KFILENAME, ed)
	defer ofs.RemoveAll(KFILENAME)
	assert.Equal(t, nil, oerr, nil)

	od, oerr := ofs.ReadFile(KFILENAME)
	assert.Equal(t, ed, od, nil)
	assert.Equal(t, nil, oerr, nil)

	md, merr := mfs.ReadFile(KFILENAME)
	assert.Equal(t, ed, md, nil)
	assert.Equal(t, nil, merr, nil)

	dd, derr := dfs.ReadFile(KFILENAME)
	assert.Equal(t, []byte(nil), dd, nil)
	assert.NotEqual(t, nil, derr, nil)
}
