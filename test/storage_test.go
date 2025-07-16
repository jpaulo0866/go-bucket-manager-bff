package test

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockStrategy(t *testing.T) {
	strat := &mockStrategy{}
	files, err := strat.ListFiles("bucket")
	assert.NoError(t, err)
	assert.Equal(t, "test.log", files[0].FileName)

	rc, err := strat.DownloadFile("bucket", "test.log")
	assert.NoError(t, err)
	data, _ := io.ReadAll(rc)
	assert.Equal(t, "logdata", string(data))

	err = strat.UploadFile("bucket", "test.log", strings.NewReader("data"))
	assert.NoError(t, err)

	url, err := strat.PresignedURL("bucket", "test.log")
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com", url)
}
