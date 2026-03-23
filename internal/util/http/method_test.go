package http

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetMethod(t *testing.T) {
	assert.True(t, IsValid("get"), "")
	assert.True(t, IsValid("GET"), "")
}

func TestPostMethod(t *testing.T) {
	assert.True(t, IsValid("post"), "")
	assert.True(t, IsValid("POST"), "")
}

func TestDeleteMethod(t *testing.T) {
	assert.True(t, IsValid("delete"), "")
	assert.True(t, IsValid("DELETE"), "")
}

func TestPutMethod(t *testing.T) {
	assert.True(t, IsValid("put"), "")
	assert.True(t, IsValid("PUT"), "")
}

func TestHeadMethod(t *testing.T) {
	assert.True(t, IsValid("head"), "")
	assert.True(t, IsValid("HEAD"), "")
}

func TestConnectMethod(t *testing.T) {
	assert.True(t, IsValid("connect"), "")
	assert.True(t, IsValid("CONNECT"), "")
}

func TestTraceMethod(t *testing.T) {
	assert.True(t, IsValid("trace"), "")
	assert.True(t, IsValid("TRACE"), "")
}

func TestOptionsMethod(t *testing.T) {
	assert.True(t, IsValid("options"), "")
	assert.True(t, IsValid("OPTIONS"), "")
}

func TestPatchMethod(t *testing.T) {
	assert.True(t, IsValid("patch"), "")
	assert.True(t, IsValid("PATCH"), "")
}

func TestFailedMethod(t *testing.T) {
	assert.False(t, IsValid("failed"), "")
	assert.False(t, IsValid("FAILED"), "")
}
