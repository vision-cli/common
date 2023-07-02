package execute_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vision-cli/common/execute"
)

func TestCommandExists_WhenCommandExists_ReturnsTrue(t *testing.T) {
	e := execute.NewOsExecutor()
	assert.True(t, e.CommandExists("ls"))
}

func TestOutput_ForLsOfThisTest_ReturnsCorrectString(t *testing.T) {
	e := execute.NewOsExecutor()
	cmd := exec.Command("ls", "execute_test.go")
	out, err := e.Output(cmd, ".", "testing")
	require.NoError(t, err)
	assert.Equal(t, "execute_test.go\n", out)
}

func TestOutput_ForUnexecutable_ReturnsError(t *testing.T) {
	e := execute.NewOsExecutor()
	cmd := exec.Command("execute_test.go")
	_, err := e.Output(cmd, ".", "testing")
	require.Error(t, err)
}

func TestErrors_ForLsOfThisTest_ReturnsNoErrors(t *testing.T) {
	e := execute.NewOsExecutor()
	cmd := exec.Command("ls", "execute_test.go")
	err := e.Errors(cmd, ".", "testing")
	require.NoError(t, err)
}

func TestErrors_ForUnexecutable_ReturnsErrors(t *testing.T) {
	e := execute.NewOsExecutor()
	cmd := exec.Command("execute_test.go")
	err := e.Errors(cmd, ".", "testing")
	require.Error(t, err)
}
