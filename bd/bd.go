package bd

import (
	"os"
	"log"
)

//原地更新的设备文件
type BD struct {
	//TODO: 这里定义你自己的数据结构
}


//如果这个设备已经存在，则只能用Open打开
func Open(name string) *BD {
	bd := BD {}

	//TODO: 下面是你自己的实现代码

	return &bd
}

//对不存在的设备，用Create创建
func Create(name string) *BD {
	file := BD {}

	//TODO: 下面是你自己的实现代码

	return &file
}

//删掉已经存在的设备
func Remove(name string) {
	//TODO: 下面是你自己的实现代码
}

//从设备的某个位置读出512Bytes这么多数据，目前只支持一次读出512字节
//blk -- 是一个512Bytes的数据块
//lba -- Logic Block Address，是一个数据块的地址，第1个512B的lba地址为0,
//       第2个512B的lba地址为1, ...
func (bd *BD) ReadAt(blk []byte, lba int64) error {
	//TODO: 下面是你自己的实现代码
	return nil
}

//这是一个原地更新的接口，把512字节数据写到lba这个位置。如果这个位置以前就存在
//数据，那么覆盖掉这些数据。
//blk -- 是一个512Bytes的数据块
//lba -- Logic Block Address，是一个数据块的地址，第1个512B的lba地址为0,
//       第2个512B的lba地址为1, ...
func (bd *BD) WriteAt(blk []byte, lba int64) error {
	//TODO: 下面是你自己的实现代码
	return nil
}


