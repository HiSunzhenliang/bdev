package aof

import (
	"os"
	"log"
)

//Append Only File
type AOF struct {
	f *os.File
}

//对已经存在的文件，用open打开
func Open(name string) (a *AOF, err error) {
	file := AOF {}
	f, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	file.f = f
	return &file, nil
}

//对不存在的文件，用Create创建
func Create(name string) (a *AOF, err error) {
	file := AOF {}
	f, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	file.f = f
	return &file, nil
}

/*
//返回文件属性包括下面一些信息
type FileInfo interface {
	Name() string       // base name of the file
	Size() int64        // length in bytes for regular files;
	Mode() FileMode     // file mode bits
	ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	Sys() any           // underlying data source (can return nil)
}
*/
func Stat(name string)(fi os.FileInfo, err error) {
	return os.Stat(name)
}

//删掉已经存在的文件
func Remove(name string) {
	err := os.Remove(name)
	if err != nil {
		log.Fatal(err)
	}
}

//从文件的某个位置读出len(b)这么多数据
func (f *AOF) ReadAt(b []byte, off int64) (n int, err error) {
	return f.f.ReadAt(b, off)
}

//这是一个追加写的接口，返回写入的位置
func (f *AOF) Append(b []byte) int64 {
	off, err := f.f.Seek(0, 2)
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.f.Write(b)
	if err != nil {
		log.Fatal(err)
	}
	return off
}

//从文件的某个位置读出len(b)这么多数据
func (f *AOF) Close() error {
	return f.f.Close()
}

