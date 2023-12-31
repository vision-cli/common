package plugins

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vision-cli/common/execute"
	"github.com/vision-cli/common/file"
	"github.com/vision-cli/common/tmpl"
)

const (
	goBinEnvVar      = "GOBIN"
	visionSeparator  = "-"
	visionFirstWord  = "vision"
	visionSecondWord = "plugin"
)

type Plugin struct {
	Name            string
	PluginPath      string
	InternalCommand func(input string, e execute.Executor, t tmpl.TmplWriter) string
}

var InternalPlugins = []Plugin{}

func GetPlugins(executor execute.Executor) ([]Plugin, error) {
	var plugins []Plugin
	pluginPath, err := goBinPath(executor)
	if err != nil {
		return plugins, err
	}
	pluginFiles, err := file.ReadDir(pluginPath)
	if err != nil {
		return plugins, fmt.Errorf("cannot read plugin directory %s: %s", pluginPath, err.Error())
	}
	for _, pluginFile := range pluginFiles {
		if !pluginFile.IsDir() && fileIsVisionPlugin(pluginFile.Name()) && !fileIsInternalPlugin(pluginFile.Name()) {
			plugins = append(plugins, Plugin{
				Name:            pluginFile.Name(),
				PluginPath:      filepath.Join(pluginPath, pluginFile.Name()),
				InternalCommand: nil,
			})
		}
	}

	plugins = append(plugins, InternalPlugins...)

	return plugins, nil
}

func goBinPath(executor execute.Executor) (string, error) {
	goBinPath := file.GetEnv(goBinEnvVar)
	if goBinPath == "" {
		goPath, err := executor.Output(exec.Command("go", "env", "GOPATH"), ".", "getting GOPATH")
		if err != nil {
			return "", err
		}
		goBinPath = string(goPath)[:len(goPath)-1] + "/bin"
	}
	return goBinPath, nil
}

func fileIsVisionPlugin(filename string) bool {
	c := strings.Split(filename, visionSeparator)
	if len(c) != 4 || c[0] != visionFirstWord || c[1] != visionSecondWord {
		return false
	}
	return true
}

func fileIsInternalPlugin(filename string) bool {
	for _, internalPlugin := range InternalPlugins {
		if filename == internalPlugin.Name {
			return true
		}
	}
	return false
}
