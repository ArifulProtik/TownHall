package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBroker_EmptyURLGivesMemory(t *testing.T) {
	b, err := NewBroker("")
	require.NoError(t, err)
	require.NotNil(t, b)
	_, ok := b.(*memoryBroker)
	assert.True(t, ok, "expected memory broker, got %T", b)
}

func TestNewBroker_BadURLReturnsError(t *testing.T) {
	_, err := NewBroker("redis://[::1]:namedport")
	require.Error(t, err)
}
