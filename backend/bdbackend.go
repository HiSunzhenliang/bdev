package backend

import (
	"os"
	"sync"
)

type BdBackend struct {
	file *os.File
	lock sync.RWMutex
}

func NewBdBackend(file *os.File) *BdBackend {
	return &BdBackend{file, sync.RWMutex{}}
}

func (b *BdBackend) ReadAt(p []byte, off int64) (n int, err error) {
	b.lock.RLock()

	n, err = b.file.ReadAt(p, off)

	b.lock.RUnlock()

	return
}

func (b *BdBackend) WriteAt(p []byte, off int64) (n int, err error) {
	b.lock.Lock()

	n, err = b.file.WriteAt(p, off)

	b.lock.Unlock()

	return
}

func (b *BdBackend) Size() (int64, error) {
	stat, err := b.file.Stat()
	if err != nil {
		return -1, err
	}

	return stat.Size(), nil
}

func (b *BdBackend) Sync() error {
	return b.file.Sync()
}

