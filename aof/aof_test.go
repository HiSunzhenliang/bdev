package aof

import (
	"testing"
	"os"
	"fmt"
	_ "encoding/binary"
)

func TestOpen3A(t *testing.T) {
	fmt.Printf("TestOpen3A\n")
	fn := "t1.aof"
	_, err := os.Stat(fn)
	if err != nil {
		_, err = os.Create(fn)
		if err != nil {
			t.Fatal(err)
		}
	}
	a, err1 := Open(fn)
	if err1 != nil {
		t.Fatal(err1)
	}
	a.Close()
	os.Remove(fn)
}

func RemoveIfExist(t *testing.T, fn string) {
	_, err := os.Stat(fn)
	if err == nil {
		err = os.Remove(fn)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestCreate3A(t *testing.T) {
	fmt.Printf("TestCreate3A\n")
	fn := "t2.aof"
	RemoveIfExist(t, fn)
	a, err1 := Create(fn)
	if err1 != nil {
		t.Fatal(err1)
	}
	a.Close()
	os.Remove(fn)
}

func CreateAndAppend(t *testing.T, fn string, b []byte) *AOF {
	RemoveIfExist(t, fn)
	a, err1 := Create(fn)
	if err1 != nil {
		t.Fatal(err1)
	}
	a.Append(b)
	return a
}

func TestAppend3A(t *testing.T) {
	fmt.Printf("TestAppend3A\n")
	fn := "t3.aof"
	b := make([]byte, 512)
	a := CreateAndAppend(t, fn, b)
	off := a.Append(b)
	if off != 512 {
		t.Fatalf("offset error")
	}
	a.Close()
	Remove(fn)
}

func TestRead3A(t *testing.T) {
	fmt.Printf("TestRead3A\n")
	fn := "t4.aof"
	b := make([]byte, 512)
	b[0] = 'h'
	b[1] = 's'
	b[2] = 'y'
	a := CreateAndAppend(t, fn, b)
	b1 := make([]byte, 512)
	a.ReadAt(b1, 0)
	if b1[0]!='h' || b1[1] != 's' || b1[2] != 'y' {
		t.Fatalf("error")
	}
	a.Close()
	Remove(fn)
}

