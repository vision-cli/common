package comms_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	api_v1 "github.com/vision-cli/api/v1"
	"github.com/vision-cli/common/comms"
	"github.com/vision-cli/common/mocks"
)

func TestCall_WhenRequestIsNil_ReturnsError(t *testing.T) {
	_, err := comms.Call[int]("plugin", nil, nil)
	require.Error(t, err)
}

func TestCall_WhenExecutorIsNil_ReturnsError(t *testing.T) {
	_, err := comms.Call[int]("plugin", &pluginRequest, nil)
	require.Error(t, err)
}

func TestCall_WhenReturnIsValid_ReturnsStruct(t *testing.T) {
	e := mocks.NewMockExecutor()
	e.SetOutput(`{"Msg":"hello"}`)
	result, err := comms.Call[TestMsg]("plugin", &pluginRequest, &e)
	require.NoError(t, err)
	require.Equal(t, "hello", result.Msg)
}

func TestCall_WhenReturnIsInvalid_ReturnsError(t *testing.T) {
	t.Skip()
	e := mocks.NewMockExecutor()
	e.SetOutput(`{}`)
	result, err := comms.Call[TestMsg]("plugin", &pluginRequest, &e)
	require.NoError(t, err)
	require.Equal(t, "hello", result.Msg)
}

func TestCall_WhenExecutorFails_ReturnsError(t *testing.T) {
	e := mocks.NewMockExecutor()
	e.SetOutputErr(fmt.Errorf("error"))
	_, err := comms.Call[int]("plugin", &pluginRequest, &e)
	require.Error(t, err)
}

var pluginRequest = api_v1.PluginRequest{
	Command: "run",
	Args:    []string{},
}

type TestMsg struct {
	Msg string
}
