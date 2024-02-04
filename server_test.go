package tftp

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenFile_NotFound(t *testing.T) {
	filename := "foobarbazquxquux"
	srv := NewServer()
	a := assert.New(t)

	f, err := srv.openFile(filename, false)

	a.Nil(f)
	a.Implements((*packet)(nil), err)
	a.Equal(err.(*errorPacket).Code, errNotFound)
	a.ErrorContains(err, "not found")
}

func TestOpenFile_IsADirectory(t *testing.T) {
	a := assert.New(t)

	dir, err := os.MkdirTemp("", "tftp_***")
	a.Nil(err)
	defer os.RemoveAll(dir)

	srv := NewServer()
	f, err := srv.openFile(dir, false)

	a.Nil(f)
	a.Implements((*packet)(nil), err)
	a.Equal(err.(*errorPacket).Code, errUndefined)
	a.ErrorContains(err, "is a directory")
}

func TestOpenFile_ReadPermissionError(t *testing.T) {
	a := assert.New(t)

	f, err := os.CreateTemp("", "tftp_***")
	a.Nil(err)
	defer os.Remove(f.Name())

	// Remove read permissions from file
	a.Nil(os.Chmod(f.Name(), 000))

	srv := NewServer()
	of, err := srv.openFile(f.Name(), false)

	a.Nil(of)
	a.Implements((*packet)(nil), err)
	a.Equal(err.(*errorPacket).Code, errPermission)
	a.ErrorContains(err, "permission denied")
}

func TestOpenFile_AlreadyExists(t *testing.T) {
	a := assert.New(t)

	f, err := os.CreateTemp("", "tftp_***")
	a.Nil(err)
	defer os.Remove(f.Name())

	srv := NewServer()
	of, err := srv.openFile(f.Name(), true)

	a.Nil(of)
	a.Implements((*packet)(nil), err)
	a.Equal(err.(*errorPacket).Code, errAlreadyExists)
	a.ErrorContains(err, "file already exists")
}

func TestOpenFile_WritePermissionError(t *testing.T) {
	a := assert.New(t)

	dir, err := os.MkdirTemp("", "tftp_***")
	a.Nil(err)
	defer os.RemoveAll(dir)

	// Remove write permissions from directory
	a.Nil(os.Chmod(dir, 000))

	filename := path.Join(dir, "foobarquxquux")
	srv := NewServer()
	f, err := srv.openFile(filename, true)

	a.Nil(f)
	a.Implements((*packet)(nil), err)
	a.Equal(err.(*errorPacket).Code, errPermission)
	a.ErrorContains(err, "permission denied")
}
