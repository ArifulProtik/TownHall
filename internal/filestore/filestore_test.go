package filestore

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ArifulProtik/TownHall/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dummyPNG = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15c4")

func TestStorage_Upload_LocalFallback(t *testing.T) {
	s := New("", t.TempDir())
	ctx := context.Background()

	_, err := s.Upload(ctx, "script.sh", []byte("echo hi"))
	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)

	_, err = s.Upload(ctx, "avatar.png", []byte("echo hi"))
	require.Error(t, err)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)

	res, err := s.Upload(ctx, "avatar.png", dummyPNG)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Contains(t, res.URL, "/uploads/")
	assert.Equal(t, "avatar.png", res.Name)
	assert.Equal(t, int64(len(dummyPNG)), res.Size)
	assert.True(t, strings.HasSuffix(res.Key, ".png"), "stored extension must come from sniffed type")

	info, err := os.Stat(filepath.Join(s.uploadsDir, res.Key))
	require.NoError(t, err)
	assert.Equal(t, int64(len(dummyPNG)), info.Size())
}

func TestStorage_Upload_TooLarge(t *testing.T) {
	s := New("", t.TempDir())

	big := make([]byte, MaxUploadBytes+1)
	copy(big, dummyPNG)
	_, err := s.Upload(context.Background(), "big.png", big)
	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)
}
