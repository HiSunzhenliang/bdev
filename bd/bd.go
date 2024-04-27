package bd

import (
	"sync"
	"io/ioutil"
	"log"
	"fmt"
	"sort"
	"strings"
	"time"
	"path/filepath"
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

	//硬盘上不可变的component，cpnt文件命名采用"<name>-<level>-<seq>.cpnt"的
	//形式，其中<name>是bd设备的名字，<level>是这个cpnt文件的层级，<seq>是
	//合并次数的序号。LSM tree的每个层级，有一个persist文件，最底层的文件是
	//persist[0]，更高一层的是persist[1]，...，最顶层是persist[n]。当新的
	//immutable写入硬盘之后，就变成了persist[n+1]
	persist []*Cpnt

	c chan int	//这个channel用来通知compaction开始工作
	seq int32
	closing bool
	wg sync.WaitGroup
}

func getCpntFiles(name string)([]string) {
	files := make([]string, 0)

	dir := filepath.Dir(name)
	base := filepath.Base(name)

	entries, err := ioutil.ReadDir(dir)
	Assert(err == nil)

	for _, entry := range entries {
		fname := entry.Name()
		if filepath.Ext(fname) != ".cpnt" {
			continue
		}
		if !strings.HasPrefix(fname, base) {
			continue
		}
		files = append(files, fname)
	}
	return files
}

func dumpBd(bd *BD) {
	for _, p := range bd.persist {
		a := p.ft.Name[:]
		log.Printf("%s\n", string(a))
	}
}

func (bd *BD)merge(p, q int) {
	ps := bd.persist[p].Size()
	qs := bd.persist[q].Size()
	//如果上一级x10比下一级大，则合并
	if ps * 10 > qs {
		bd.mutex.Unlock()
		c := MergeCpnt(bd.name, int32(q), bd.seq,
			bd.persist[p], bd.persist[q])
		bd.mutex.Lock()
		bd.seq++
		bd.persist[q] = c
		bd.persist = append(bd.persist[:p],
				bd.persist[p+1:]...)
	}
}

const WrLimit = 10
func (bd *BD)Compaction() {
	defer bd.wg.Done()
	for !bd.closing {
		select {
		case <-bd.c:
		case <-time.After(1*time.Second):
		}

		bd.mutex.Lock()
		//如果mutable里面大于10个block，则刷盘
		if (bd.mutable.Size() > WrLimit) || bd.closing {
			bd.immutable = bd.mutable
			bd.mutable = CreateMemCpnt(bd.name + "_mut")
			bd.mutex.Unlock()
			c := CreateCpnt(bd.name, int32(len(bd.persist)), bd.seq,
						bd.immutable)
			bd.mutex.Lock()
			bd.seq++
			bd.persist = append(bd.persist, c)
			bd.immutable = nil
		}

		p := len(bd.persist) - 1
		for p > 0 && !bd.closing {
			p := len(bd.persist) - 1
			for p > 0 {
				q := p - 1
				bd.merge(p, q)
				p--
			}
		}
		bd.mutex.Unlock()
	}
}

func NewBd(name string) *BD {
	bd := &BD {}
	bd.mutable = CreateMemCpnt(name + "_mut")
	bd.immutable = nil
	bd.name = name
	bd.seq = 0
	bd.c = make(chan int, 10)
	return bd
}

//如果这个设备已经存在，则只能用Open打开
func OpenBD(name string) (*BD, error) {

	cpntFiles := getCpntFiles(name)
	if len(cpntFiles)==0 {
		return nil, fmt.Errorf("cpnt file is not found")
	}

	bd := NewBd(name)
	for _, fname := range cpntFiles {
		c, err := OpenCpnt(fname)
		Assert(err == nil)
		bd.persist = append(bd.persist, c)
		if bd.seq < c.ft.Seq {
			bd.seq = c.ft.Seq + 1
		}
	}

	//对persist进行排序
	sort.Sort(TreeCpnt(bd.persist))

	go bd.Compaction()

	return bd, nil
}

//对不存在的设备，用Create创建
func CreateBD(name string) (*BD, error) {
	//如果设备文件已经存在，则只能用open打开，不能重新创建
	files := getCpntFiles(name)
	if len(files) > 0 {
		return nil, fmt.Errorf("BD has already existed")
	}

	bd := NewBd(name)

	go bd.Compaction()

	return bd, nil
}

func (bd *BD)Close() {
	bd.mutex.Lock()
	bd.wg.Add(1)
	bd.closing = true
	bd.c <- 2
	bd.mutex.Unlock()
	bd.wg.Wait()
}

//删掉已经存在的设备
func Remove(name string) {
	//TODO: 下面是你自己的实现代码
}

//从设备的某个位置读出512Bytes这么多数据，目前只支持一次读出512字节
//blk -- 是一个512Bytes的数据块
//lba -- Logic Block Address，是一个数据块的地址，第1个512B的lba地址为0,
//       第2个512B的lba地址为1, ...
func (bd *BD) ReadAt(lba int64) (blk []byte, ok bool) {
	bd.mutex.Lock()
	defer bd.mutex.Unlock()

	if b, ok := bd.mutable.ReadAt(lba); ok {
		return b, true
	}

	if bd.immutable != nil {
		if b, ok := bd.immutable.ReadAt(lba); ok {
			return b, true
		}
	}

	n := len(bd.persist)
	for i:=n-1; i>=0; i++ {
		if b, ok := bd.persist[i].ReadAt(lba); ok {
			return b, true
		}
	}

	return make([]byte, BlkSize), true
}

//这是一个原地更新的接口，把512字节数据写到lba这个位置。如果这个位置以前就存在
//数据，那么覆盖掉这些数据。
//blk -- 是一个512Bytes的数据块
//lba -- Logic Block Address，是一个数据块的地址，第1个512B的lba地址为0,
//       第2个512B的lba地址为1, ...
func (bd *BD) WriteAt(lba int64, blk []byte) (ok bool) {
	bd.mutex.Lock()
	defer bd.mutex.Unlock()
	ok = bd.mutable.WriteAt(lba, blk)
	if bd.mutable.Size() > WrLimit {
		bd.c <- 1
	}
	return ok
}


