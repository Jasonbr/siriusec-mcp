package unit

import (
	"siriusec-mcp/pkg/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSuccessResponse(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	resp := models.NewSuccessResponse(data)

	assert.Equal(t, string(models.ResultCodeSuccess), resp.Code)
	assert.Equal(t, data, resp.Data)
	assert.Empty(t, resp.Message)
}

func TestNewErrorResponse(t *testing.T) {
	message := "test error"
	resp := models.NewErrorResponse(message)

	assert.Equal(t, string(models.ResultCodeError), resp.Code)
	assert.Equal(t, message, resp.Message)
	assert.Nil(t, resp.Data)
}

func TestMCPResponseIsSuccess(t *testing.T) {
	successResp := models.MCPResponse{Code: string(models.ResultCodeSuccess)}
	errorResp := models.MCPResponse{Code: string(models.ResultCodeError)}

	assert.True(t, successResp.IsSuccess())
	assert.False(t, errorResp.IsSuccess())
}

func TestIsPermissionError(t *testing.T) {
	tests := []struct {
		message  string
		expected bool
	}{
		{"权限不足", true},
		{"permission denied", true},
		{"unauthorized access", true},
		{"access denied", true},
		{"forbidden", true},
		{"RAM user not authorized", true},
		{"需要授权", true},
		{"normal error", false},
		{"connection timeout", false},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := models.IsPermissionError(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnhancePermissionErrorMessage(t *testing.T) {
	original := "权限不足"
	enhanced := models.EnhancePermissionErrorMessage(original)

	assert.Contains(t, enhanced, original)
	assert.Contains(t, enhanced, "权限问题")
	assert.Contains(t, enhanced, "AccessKey")
}
