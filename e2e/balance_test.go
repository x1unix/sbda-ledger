package e2e

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBalance(t *testing.T) {
	require.NoError(t, TruncateData(), "failed to truncate data before test")

}
