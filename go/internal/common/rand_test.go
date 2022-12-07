package common_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nanobus/iota/go/internal/common"
)

func TestRandIntn(t *testing.T) {
	n := common.RandIntn(10)
	assert.True(t, n >= 0 && n < 10)
}
