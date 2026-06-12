package http

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/assignment"
)

func TestReadSB3UploadRejectsOversizedBody(t *testing.T) {
	_, err := readSB3Upload(bytes.NewReader(bytes.Repeat([]byte("a"), 6)), 5)

	require.ErrorIs(t, err, assignment.ErrSB3TooLarge)
}

func TestReadSB3UploadReturnsBodyWithinLimit(t *testing.T) {
	rawSB3, err := readSB3Upload(bytes.NewReader([]byte("abc")), 5)

	require.NoError(t, err)
	require.Equal(t, []byte("abc"), rawSB3)
}
