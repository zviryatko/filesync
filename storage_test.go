package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

var key = "momsbestpicture"

func TestPathTransformFunc(t *testing.T) {
	pathKey, err := CASPathTransformFunc(key)
	assert.Nil(t, err)
	assert.Equal(t, PathKey{
		PathName: "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff",
		Filename: "6804429f74181a63c50c3d81d733a12f14a353ff",
	}, pathKey)
}

func TestStore(t *testing.T) {
	s := NewStore(StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	})
	data := []byte("hello world")
	l, err := s.writeStream(key, bytes.NewReader(data))
	assert.Nil(t, err)
	assert.Equal(t, int64(len(data)), l)
	r, err := s.Read(key)
	assert.Nil(t, err)
	b, _ := io.ReadAll(r)
	assert.Equal(t, data, b)

	h, err := s.Has(key)
	assert.True(t, h)
	assert.Nil(t, err)

	assert.Nil(t, s.Delete(key))

	h1, err := s.Has(key)
	assert.False(t, h1)
	assert.Nil(t, err)
}
