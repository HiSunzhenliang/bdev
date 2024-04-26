package bd

import (
	"bdev/aof"
	"sort"
	"log"
	"bytes"
	"encoding/binary"
)

const BlkSize	= 512

type LbaBuf struct {
	lba int64	//数据块的LBA地址
	buf []byte
}

//这是一个LSM tree的component类型，是在内存中的
type MemCpnt struct {
	name string
	lb []*LbaBuf
}

func CreateMemCpnt(name string) *MemCpnt {
	mc := &MemCpnt{}
	mc.name = name
	return mc
}

func (m *MemCpnt) Size() int {
	return len(m.lb)
}

func (m *MemCpnt) Find(lba int64) (idx int, found bool) {
	idx, found = sort.Find(len(m.lb), func(i int) int {
		return int(lba - m.lb[i].lba)
	})
	return idx, found
}

//从一个内存Component中读出一个blk
func (m *MemCpnt) ReadAt(lba int64) (blk []byte, ok bool) {
	i, f := m.Find(lba)
	if !f {
		return []byte{}, false
	}
	return m.lb[i].buf, true
}

//迭代读出所有blk
func (m *MemCpnt) ReadIter(iter *int) (lba int64, blk []byte, ok bool) {
	if *iter >= len(m.lb) {
		return int64(-1), []byte{}, false
	}
	i := *iter
	(*iter)++
	return m.lb[i].lba, m.lb[i].buf, true
}

//向一个内存component中写入一个blk
func (m *MemCpnt) WriteAt(lba int64, blk []byte) (ok bool){
	Assert(len(blk) == BlkSize)
	b := make([]byte, BlkSize)
	copy(b, blk)
	i, f := m.Find(lba)
	if f {
		//如果这个位置已经写过数据了，则覆盖掉这个数据
		m.lb[i].buf = b
	} else {
		//如果这个位置没有写过数据，则插入这个数据块
		n := &LbaBuf{lba, b}
		m.lb = append(m.lb[:i], append([]*LbaBuf{n}, m.lb[i:]...)...)
	}
	return true
}

//LBA和数据在aof文件中位置的对应关系
type LbaOff struct {
	Lba int64	//数据块的LBA地址
	Off int64	//数据块在aof文件中的偏移量
}

//这是一个LSM tree的component类型，是存储在硬盘上的
type Cpnt struct {
	a *aof.AOF
	lo []LbaOff
	ft CpntFooter
}

const NameLen = 32		//cpnt文件名的长度
const FooterLen = 128		//整个footer的长度
type CpntFooter struct {
	Name [NameLen]byte	//CPNT文件的名字
	NumBlk int64		//总的块数量
	IdxOff int64		//索引块的起始位置
	Level int32
	Seq int32
	Res [9]int64		//保留字段，暂时没有用
}

/* 下面这些定义，目的是用来对[]*Cpnt排序 */
type TreeCpnt []*Cpnt
func (tc TreeCpnt) Len() int           { return len(tc) }
func (tc TreeCpnt) Swap(i, j int)      { tc[i], tc[j] = tc[j], tc[i] }
func (tc TreeCpnt) Less(i, j int) bool {
	return tc[i].ft.Level < tc[j].ft.Level ||
		(tc[i].ft.Level == tc[j].ft.Level &&
		 tc[i].ft.Seq < tc[j].ft.Seq )
}

func (ft *CpntFooter)AssignName(name string) {
	b := []byte(name)
	n := NameLen
	if len(b) < n {
		n = len(b)
	}
	for i:=0; i<n; i++ {
		ft.Name[i] = b[i]
	}
}

func newCpnt(name string, level, seq int32) (*Cpnt, *CpntFooter) {
	c := &Cpnt{}
	ft := &c.ft
	ft.AssignName(name)
	ft.Level = level
	ft.Seq = seq
	a, err := aof.Create(name)
	if err != nil {
		log.Fatalf("create AOF file error\n")
	}
	c.a = a
	return c, ft
}


func (c *Cpnt) Size() int {
	return len(c.lo)
}

func writeFooter(c *Cpnt, ft *CpntFooter) {
	buf := new(bytes.Buffer)
	for i:=0; i<len(c.lo); i++ {
		err := binary.Write(buf, binary.LittleEndian, c.lo[i])
		Assert(err == nil)
	}
	b := buf.Bytes()
	if (len(b) % BlkSize) != 0 {
		n := ((len(b) + BlkSize - 1) / BlkSize) * BlkSize
		d := n - len(b)
		b = append(b, make([]byte, d)...)
	}
	off := c.a.Append(b)
	Assert(off == ft.IdxOff)

	buf = new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, ft)
	Assert(err == nil)

	b = buf.Bytes()
	Assert(len(b) == FooterLen)

	c.a.Append(b)
}

//把一个内存Component写入硬盘，变成一个持久化的Component
func CreateCpnt(name string, level, seq int32, m *MemCpnt) *Cpnt {

	c, ft := newCpnt(name, level, seq)
	ft.NumBlk = int64(len(m.lb))
	ft.IdxOff = ft.NumBlk * BlkSize

	for i:=0; i<len(m.lb); i++ {
		c.lo = append(c.lo, LbaOff{m.lb[i].lba, int64(i * BlkSize)})
	}

	for i:=0; i<len(m.lb); i++ {
		off := c.a.Append(m.lb[i].buf)
		Assert(off == c.lo[i].Off)
	}

	writeFooter(c, ft)

	return c
}

//打开硬盘上的一个持久化的cpnt。一个CPNT只能被写一次，因此打开之后的CPNT不能再
//被改写，只能读。
func OpenCpnt(name string) (*Cpnt, error) {

	//获取aof文件的属性，目的是：1、判断文件是否存在，如果不存在，则
	//返回错误；2、获取文件size，后面读取footer的时候会用到size；
	st, err1 := aof.Stat(name)
	if err1 != nil {
		return nil, err1
	}
	size := st.Size()

	//打开AOF文件
	c := &Cpnt{}
	a, err := aof.Open(name)
	if err != nil {
		log.Fatalf("create AOF file error\n")
	}
	c.a = a

	//读出footer所在位置的bytes数据
	b := make([]byte, FooterLen)
	n, err2 := c.a.ReadAt(b, size - FooterLen)
	if err2 != nil {
		log.Printf("err2 = %v\n", err2)
	}
	Assert(err2 == nil)
	Assert(n == FooterLen)

	//把bytes数据转换为footer结构体
	ft := &c.ft
	r := bytes.NewReader(b)
	if err := binary.Read(r, binary.LittleEndian, ft); err != nil {
		log.Fatalf("binary.Read failed: %v\n", err)
	}

	//计算硬盘上index区域的字节数
	lo := &LbaOff{}
	loSize := int(ft.NumBlk) * binary.Size(lo)
	loBlk := (loSize + BlkSize - 1) / BlkSize
	loBytes := loBlk * BlkSize

	//从硬盘上读出index区域，把数据存储字节数组b
	b = make([]byte, loBytes)
	n, err2 = c.a.ReadAt(b, ft.IdxOff)
	Assert(err2 == nil)
	Assert(n == loBytes)

	//把字节数组b中的LbaOff数据，转换之后存入c.lo数组中
	r = bytes.NewReader(b)
	for i:=int64(0); i<ft.NumBlk; i++ {
		if err := binary.Read(r, binary.LittleEndian, lo); err != nil {
			log.Fatalf("binary.Read failed: %v\n", err)
		}
		c.lo = append(c.lo, *lo)
	}

	return c, nil
}

//删除硬盘上一个持久化的cpnt
func RemoveCpnt(name string) {
	aof.Remove(name)
}

//把两个硬盘上持久化的cpnt合并成一个。其中a是相对新的cpnt，b是相对旧的cpnt，
//如果a和b中有相同的，那么
func MergeCpnt(name string, level, seq int32, a *Cpnt, b *Cpnt) *Cpnt {
	c, ft := newCpnt(name, level, seq)
	ft.NumBlk = 0

	var lp *LbaOff
	var t *Cpnt

	/*
	 * 执行归并排序过程，把a和b中的blk写入c
	 */
	ap, bp := 0, 0	/* LbaOff中数组的新位置 */
	for {
		if ap >= len(a.lo) && bp >= len(b.lo) {
			/* 如果a和b都处理完了，则退出循环 */
			break
		} else if ap >= len(a.lo) && bp < len(b.lo) {
			/* 如果a完了，b还有，则只取b */
			lp = &b.lo[bp]
			bp++
			t = b
		} else if ap < len(a.lo) && bp >= len(b.lo) {
			/* 如果b完了，a还有，则只取a */
			lp = &a.lo[ap]
			ap++
			t = a
		} else if a.lo[ap].Lba == b.lo[bp].Lba {
			/* 如果a和b的LBA相等，因为a比较新，两者都弹出，
			 * 但是只取a */
			lp = &a.lo[ap]
			ap++
			bp++
			t = a
		} else if a.lo[ap].Lba < b.lo[bp].Lba {
			/* 如果a的lba更小，则取a */
			lp = &a.lo[ap]
			ap++
			t = a
		} else {
			/* 如果b的lba更小，则取b */
			lp = &b.lo[bp]
			bp++
			t = b
		}

		c.lo = append(c.lo, LbaOff{lp.Lba, int64(ft.NumBlk * BlkSize)})

		blk, ok := t.ReadAt(lp.Lba)
		Assert(ok)

		off := c.a.Append(blk)
		Assert(off == (BlkSize * ft.NumBlk))

		ft.NumBlk++
	}

	ft.IdxOff = ft.NumBlk * BlkSize
	writeFooter(c, ft)
	return c
}

//在LbaOff数组中查找lba，如果找到则返回index和true
func (c *Cpnt) Find(lba int64) (idx int, found bool) {
	idx, found = sort.Find(len(c.lo), func(i int) int {
		return int(lba - c.lo[i].Lba)
	})
	return idx, found
}

//读出lba位置的一个block
func (c *Cpnt) ReadAt(lba int64)(blk []byte, ok bool) {

	i, f := c.Find(lba)
	if !f {
		return []byte{}, false
	}

	b := make([]byte, BlkSize)
	n, err := c.a.ReadAt(b, c.lo[i].Off)
	Assert(n==BlkSize)
	Assert(err==nil)
	return b, true
}

func (c *Cpnt) Close() {
	c.a.Close()
}



