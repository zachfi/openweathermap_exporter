package owm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	c := Config{}
	o, err := New(c)
	require.NotNil(t, o)
	require.NoError(t, err)

}
