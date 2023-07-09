package plugins_test

import (
	"fmt"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vision-cli/common/file"
	"github.com/vision-cli/common/mocks"
	"github.com/vision-cli/common/plugins"
)

func TestGoGetPlugins_WhenEnvSet_ReturnsGood(t *testing.T) {
	oldosgetenv := file.Osgetenv
	defer func() { file.Osgetenv = oldosgetenv }()
	file.Osgetenv = func(key string) string {
		return "/usr/local/go/bin"
	}

	oldreaddir := file.Osreaddir
	defer func() { file.Osreaddir = oldreaddir }()
	file.Osreaddir = mockReadDir

	e := mocks.NewMockExecutor()
	_, err := plugins.GetPlugins(&e)
	require.NoError(t, err)
}

func TestGoGetPlugins_CantReadDir_ReturnsError(t *testing.T) {
	old := file.Osgetenv
	defer func() { file.Osgetenv = old }()
	file.Osgetenv = func(key string) string {
		return "/usr/local/somethingelse"
	}
	e := mocks.NewMockExecutor()
	_, err := plugins.GetPlugins(&e)
	require.Error(t, err)
}

func TestGoGetPlugins_WhenEnvNotSet_CallsGoEnvPath(t *testing.T) {
	oldosgetenv := file.Osgetenv
	defer func() { file.Osgetenv = oldosgetenv }()
	file.Osgetenv = func(key string) string {
		return ""
	}

	oldreaddir := file.Osreaddir
	defer func() { file.Osreaddir = oldreaddir }()
	file.Osreaddir = mockReadDir

	e := mocks.NewMockExecutor()
	e.SetOutput("/usr/local/go\n")
	_, err := plugins.GetPlugins(&e)
	require.NoError(t, err)
}

func TestGoGetPlugins_ReturnsAllValidPlugins(t *testing.T) {
	oldgetenv := file.Osgetenv
	defer func() { file.Osgetenv = oldgetenv }()
	file.Osgetenv = func(key string) string {
		return "/usr/local/go/bin"
	}

	oldreaddir := file.Osreaddir
	defer func() { file.Osreaddir = oldreaddir }()
	file.Osreaddir = mockReadDir

	e := mocks.NewMockExecutor()
	result, err := plugins.GetPlugins(&e)
	require.NoError(t, err)
	assert.Equal(t, []plugins.Plugin{{"vision-plugin-myplugin-v2", "/usr/local/go/bin/vision-plugin-myplugin-v2", nil}}, result)
}

type MockDirEntry struct {
	name  string
	isDir bool
}

func (m MockDirEntry) Name() string {
	return m.name
}

func (m MockDirEntry) IsDir() bool {
	return m.isDir
}

func (m MockDirEntry) Type() fs.FileMode {
	return 0
}

func (m MockDirEntry) Info() (fs.FileInfo, error) {
	return nil, nil
}

var mockReadDir = func(name string) ([]fs.DirEntry, error) {
	if name != "/usr/local/go/bin" {
		return nil, fmt.Errorf("wrong path: %s", name)
	}
	return []fs.DirEntry{
		MockDirEntry{name: "vision-plugin-myplugin-v1", isDir: true},
		MockDirEntry{name: "vision-plugin-myplugin-v2", isDir: false},
		MockDirEntry{name: "vision-plugin-myplugin", isDir: false},
		MockDirEntry{name: "vision-plugin", isDir: false},
		MockDirEntry{name: "visions-plugin-myplugin-v1", isDir: false},
		MockDirEntry{name: "visions-plugin-myplugin-v1", isDir: false},
		MockDirEntry{name: "vision-plugin-myplugin-v1-extra", isDir: false},
	}, nil
}
