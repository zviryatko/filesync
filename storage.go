package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// defaultRoot is the default root directory where the files are stored.
const defaultRoot = "network"

// PathTransformFunc transforms filename to internal path system.
type PathTransformFunc func(string) (PathKey, error)

// CASPathTransformFunc converts "path" to "p/a/t/h/path".
func CASPathTransformFunc(path string) (PathKey, error) {
	hash := sha1.Sum([]byte(path))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i+1)*blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashStr,
	}, nil
}

// PathKey is the internal representation of a file path.
type PathKey struct {
	PathName string
	Filename string
}

// StoreOpts contains the options for the Store.
type StoreOpts struct {
	// Root is the root directory where the files are stored.
	Root              string
	PathTransformFunc PathTransformFunc
}

// Store is a file storage system.
type Store struct {
	StoreOpts
}

// NewStore creates a new Store.
func NewStore(opts StoreOpts) *Store {
	if opts.Root == "" {
		opts.Root = defaultRoot
	}
	return &Store{opts}
}

// Has checks if the file exists in the store.
func (s *Store) Has(key string) (bool, error) {
	pathKey, err := s.PathTransformFunc(key)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(s.path(pathKey))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Delete deletes the file from the store.
func (s *Store) Delete(key string) error {
	pathKey, err := s.PathTransformFunc(key)
	if err != nil {
		return err
	}
	fullpath := s.path(pathKey)

	defer func() {
		log.Printf("deleted [%s] from disk", fullpath)
	}()

	_, err = os.Stat(fullpath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	// Remove the file.
	if err := os.Remove(fullpath); err != nil {
		return err
	}

	// Split the path into its components and remove them one by one if empty.
	// This is to avoid removing a parent directory that may contain other files.
	// For example, if the path is "a/b/c/d/e/f/g", we want to remove "g" first,
	// then "f", then "e", and so on.
	components := strings.Split(fullpath, "/")
	for i := len(components) - 2; i >= 1; i-- {
		dir := strings.Join(components[:i+1], "/")
		// Check if dir exists.
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			continue
		}
		// Check if dir is empty.
		f, err := os.Open(dir)
		if err != nil {
			return err
		}
		_, err = f.Readdirnames(1)
		_ = f.Close()
		if err != io.EOF {
			// dir is not empty - stop removing directories.
			break
		}

		if err := os.Remove(dir); err != nil {
			return err
		}
	}

	return nil
}

// Read reads the file from the store.
func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf, err
}

func (s *Store) path(key PathKey) string {
	return fmt.Sprintf("%s/%s/%s", s.Root, key.PathName, key.Filename)
}

func (s *Store) dir(key PathKey) string {
	return fmt.Sprintf("%s/%s", s.Root, key.PathName)
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey, err := s.PathTransformFunc(key)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(s.path(pathKey))
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey, err := s.PathTransformFunc(key)
	if err != nil {
		return err
	}
	pathAndFilename := s.path(pathKey)

	if err := os.MkdirAll(s.dir(pathKey), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(pathAndFilename)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("Wrote %d bytes to %s", n, pathAndFilename)

	return nil
}
