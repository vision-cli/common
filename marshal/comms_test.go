package marshal_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	api_v1 "github.com/vision-cli/api/v1"
	"github.com/vision-cli/common/marshal"
)

func TestUnmarshal_WithValidInputProvided_ReturnsObject(t *testing.T) {
	req := `{"Result":"result","Error":""}`
	result, err := marshal.Unmarshal[api_v1.PluginResponse](req)
	expected := api_v1.PluginResponse{
		Result: "result",
		Error:  "",
	}
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestUnmarshal_ReturnsErrorWhenInValidInputProvided(t *testing.T) {
	req := `{"Result":"result","Error":"",}` // extra comma
	_, err := marshal.Unmarshal[api_v1.PluginResponse](req)
	require.Error(t, err)
}

func TestMarshal_WithValidObject_ReturnsString(t *testing.T) {
	obj := api_v1.PluginResponse{
		Result: "result",
		Error:  "",
	}
	result, err := marshal.Marshal[api_v1.PluginResponse](obj)
	expected := `{"Result":"result","Error":""}`
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestMarshal_WithInValidObject_ReturnsError(t *testing.T) {
	_, err := marshal.Marshal[float64](math.Inf(1))
	assert.Equal(t, "json: unsupported value: +Inf", err.Error())
}
