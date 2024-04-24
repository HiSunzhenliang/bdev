package bd

import (
	"sync"
)

//支持原地更新的设备文件
type BD struct {
	name string

	//当两个components执行compaction完毕之后，切换期间需要加锁
	mutex sync.Mutex

	//当前内存中可变的component
	mutable	*MemCpnt

	//当前内存中不可变的component
	immutable *MemCpnt

	//硬盘上不可变的component，cpnt文件命名采用"<name>_<level>_<seq>.cpnt"的
	//形式，其中<name>是bd设备的名字，<level>是这个cpnt文件的层级，<seq>是
	//合并次数的序号。LSM tree的每个层级，有一个persist文件，最底层的文件是
	//persist[0]，更高一层的是persist[1]，...，最顶层是persist[n]。当新的
	//immutable写入硬盘之后，就变成了persist[n+1]
	persist []*Cpnt
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
func (bd *BD) ReadAt(lba int64) (blk []byte, error) {
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


