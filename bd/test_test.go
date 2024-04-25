package bd

import (
	"testing"
	_ "os"
	"fmt"
	_ "encoding/binary"
)

func TestMemCpnt3A(t *testing.T) {
	fmt.Printf("MemCpnt3A\n")
	m := CreateMemCpnt("cpnt-1-2.mem")
	Assert(m != nil)

	b, ok := m.ReadAt(10)
	if ok {
		fmt.Printf("b = %v", b)
		t.Fatalf("unexpect ok\n")
	}

	c := make([]byte, BlkSize)
	c[0] = 'a'
	c[1] = 'b'
	m.WriteAt(10, c)

	d, ok1 := m.ReadAt(10)
	Assert(ok1)
	Assert(d[0] == c[0])
	Assert(d[1] == c[1])

	c[0] = 'd'
	m.WriteAt(10, c)
	d, ok1 = m.ReadAt(10)
	Assert(d[0] == c[0])
	Assert(d[1] == c[1])

	m.WriteAt(10, c)
	m.WriteAt(12, c)
	m.WriteAt(13, c)
	m.WriteAt(5, c)
	m.WriteAt(8, c)
	m.WriteAt(7, c)
	m.WriteAt(10, c)
	i := 0
	lst := make([]int64, 0)
	for {
		lba, _, ok3 := m.ReadIter(&i)
		if !ok3 {
			break
		}
		lst = append(lst, lba)
	}
	Assert(len(lst) == 6)
	Assert(lst[0] == 5)
	Assert(lst[1] == 7)
	Assert(lst[2] == 8)
	Assert(lst[3] == 10)
	Assert(lst[4] == 12)
	Assert(lst[5] == 13)
}

func getBuf(lba, v int64) []byte {
	b := make([]byte, BlkSize)
	b[0] = byte(v)
	return b
}


func TestCpnt3A(t *testing.T) {
	fmt.Printf("Cpnt3A\n")
	m1 := CreateMemCpnt("cpnt-1-2.mem")
	for i:=int64(0); i<10; i++ {
		b := getBuf(i, i)
		m1.WriteAt(i, b)
	}
	b := getBuf(20, 20)
	m1.WriteAt(20, b)
	c1 := CreateCpnt("test-1-1.cpnt", 0, 1, m1)
	b1, o1 := c1.ReadAt(1)
	Assert(o1)
	Assert(b1[0] == 1)

	m2 := CreateMemCpnt("cpnt-2-2.mem")
	for i:=int64(5); i<15; i++ {
		b := getBuf(i, 10+i)
		m2.WriteAt(i, b)
	}
	b = getBuf(20, 22)
	m2.WriteAt(20, b)
	c2 := CreateCpnt("test-2-2.cpnt", 1, 1, m2)
	b2, o2 := c2.ReadAt(5)
	Assert(o2)
	Assert(b2[0] == 15)

	c3 := MergeCpnt("test-1-3.cpnt", 0, 2, c2, c1)
	Assert(c3 != nil)

	for i:=int64(0); i<5; i++ {
		b, ok := c3.ReadAt(i)
		Assert(ok)
		Assert(b[0]==byte(i))
	}

	for i:=int64(5); i<15; i++ {
		b, ok := c3.ReadAt(i)
		Assert(ok)
		Assert(b[0]==byte(10+i))
	}

	c3.Close()
	c4, err := OpenCpnt("test-1-3.cpnt")
	Assert(err == nil)


	for i:=int64(0); i<5; i++ {
		b, ok := c4.ReadAt(i)
		Assert(ok)
		Assert(b[0]==byte(i))
	}

	for i:=int64(5); i<15; i++ {
		b, ok := c4.ReadAt(i)
		Assert(ok)
		Assert(b[0]==byte(10+i))
	}

}

func TestOpenBD3A(t *testing.T) {

	_, _ = OpenBD("./test")
}

