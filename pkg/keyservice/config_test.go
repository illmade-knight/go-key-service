package keyservice_test

import (
	"testing"

	"github.com/illmade-knight/go-key-service/pkg/keyservice"
	"github.com/stretchr/testify/assert"
)

// TestConfigInstantiation is a simple smoke test to ensure the Config
// struct can be created and its fields can be set and read.
func TestConfigInstantiation(t *testing.T) {
	// Arrange
	expectedAddr := ":8081"

	// Act
	cfg := &keyservice.Config{
		HTTPListenAddr: expectedAddr,
	}

	// Assert
	assert.NotNil(t, cfg)
	assert.Equal(t, expectedAddr, cfg.HTTPListenAddr)
}
