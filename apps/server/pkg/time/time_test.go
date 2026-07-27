package time

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	assert.True(t, strings.HasSuffix(Now().String(), "UTC"))
}
