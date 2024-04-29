package backend

import (
	"bdev/bd"
	"fmt"
	"log"
)

type BdBackend struct {
	name string
	size int64
	bd   *bd.BD
}

func NewBdBackend(name string, size int64) *BdBackend {
	bb := &BdBackend{}
	bb.name = name
	bb.size = size
	b, err := bd.OpenBD(name)
	if err != nil {
		b, err = bd.CreateBD(name)
	}

	if err != nil {
		log.Printf("Cannot open the block device\n")
		return nil
	}

	bb.bd = b

	return bb
}

func (b *BdBackend) ReadAt(p []byte, off int64) (n int, err error) {
	n = len(p)
	bd.Assert((n % bd.BlkSize) == 0)
	bd.Assert((off % bd.BlkSize) == 0)
	k := n / bd.BlkSize
	for i:=0; i<k; i++ {
		o := bd.BlkSize * i
		lba := (off + int64(o)) / bd.BlkSize
		b, ok := b.bd.ReadAt(lba)
		if !ok {
			return 0, fmt.Errorf("unknow error")
		}
		copy(p[o:], b)
	}
	return n, nil
}

func (b *BdBackend) WriteAt(p []byte, off int64) (n int, err error) {
	n = len(p)
	bd.Assert((n % bd.BlkSize) == 0)
	bd.Assert((off % bd.BlkSize) == 0)
	k := n / bd.BlkSize
	for i:=0; i<k; i++ {
		o := bd.BlkSize * i
		lba := (off + int64(o)) / bd.BlkSize
		ok := b.bd.WriteAt(lba, p[o:(o+bd.BlkSize)])
		if !ok {
			return -1, fmt.Errorf("unknow error")
		}
	}

	return n, nil
}

func (b *BdBackend) Size() (int64, error) {
	return b.size, nil
}

func (b *BdBackend) Sync() error {
	b.bd.Sync()
	return nil
}

