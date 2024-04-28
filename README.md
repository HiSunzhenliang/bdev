
# BDEV: 基于Append-Only存储实现的块设备

## 背景

这是一个用于学习存储技术练习项目。Append-Only语义的接口，在存储系统中很常见。
但是很多上层应用不能直接使用Append-Only语义的存储接口，一般要求update-in-place
接口。则就需要把append-only接口转换成update-in-place接口，或者说，基于append-only
接口实现update-in-place接口。

Append-only接口指的是，只能在文件或者对象尾部添加数据，已经写入了文件或者对象的
数据，不能修改，只能把文件或者对象整个删除掉才能释放已经占用空间。append-only的
文件或者对象，支持随机读，也就是可以读取任何位置的数据。update-in-place接口，则
可以修改文件和对象中任何位置的数据。

Flash芯片的特性就跟Append-only语义很接近。Flash芯片只能以block为单位进行擦除，
通常一个block是64KB、128KB、256KB，甚至1MB。每次擦除之后，Flash芯片的这个block
可以被改写一次，改写的单位是page，通常是4KB或者8KB。一旦被改写，则不能再次被
改写，再次改写之前，必须先对整个block执行擦除操作。擦除操作性能开销很大。
基于flash的这些特性，可以很容易实现append-only接口。但是要实现update-in-place
接口，则没那么容易。SSD盘内部的FTL软件，一个主要功能就是实现append-only和
update-in-place之间的转换。

最近存储行业提出了ZNS接口的SSD盘，这个ZNS接口就是append-only语义的接口。
由于上层应用更多使用update-in-place接口，例如：POSIX文件接口、SCSI/NVMe等块设备
接口，或者KV数据库接口，等等。因此就需要，在ZNS接口上要实现update-in-place接口
语义的接口，块设备接口、POSIX文件接口、或者KV接口。

另外，很多分布式存储系统的存储层接口，也是Append-Only语义的。例如，阿里云的Pangu，
微软的WAS，谷歌的GFS等等，也需要基于Append-Only语义接口要实现块设备接口，文件
接口等。例如，阿里云的块存储服务EBS是基于Pangu的Append-only接口实现的。

对于存储行业研发人员来说，基于Append-only语义存储接口，如何实现update-in-place
接口，已经是必备知识了。

## BDEV

如下图所示，BDEV是一个基于append-only接口实现的nbd server。通过Linux内核的nbd
client可以链接到这个nbd server上，就可以看到Linux主机上增加了一个新的块设备
/dev/nbd0，然后可以在这个块设备上创建文件系统，挂载这个块设备。

BDEV中nbd-server协议层代码，直接用[go-nbd](https://github.com/pojntfx/go-nbd)的
实现。块设备本身Write-in-place的语义，bdev/bd这个目录中的源码实现的。几个子目录
中源码功能如下：

* 主目录 - main.go, nbd-server.go，这两个是程序入口，也是nbd-server的主循环。
* aof - 这里面是一个append-only语义的文件接口
* bd - 这里面实现了一个简化的LSM Tree，也就是基于append-only接口实现了块设备所
  需的update-in-place接口。
* backend - 这里面是nbd-server需要的存储后端接口。这里面的代码没有实际功能，
  只是负责bd里面的update-in-place接口跟nbd-server需要的接口粘合起来。

![bdev](pics/bdev.png)


## 底层append-only接口(AOF)

用golang描述的底层append-only接口如下

```golang
//append-only对象的定义
type AOF struct {
        f *os.File
}


//对已经存在的AOF文件，用open打开
func Open(name string) (a *AOF, err error)


//对不存在的AOF文件，用Create创建
func Create(name string) (a *AOF, err error)

//返回AOF的属性
func Stat(name string)(fi os.FileInfo, err error)

//AOF文件属性包括下面一些信息
type FileInfo interface {
        Name() string       // base name of the file
        Size() int64        // length in bytes for regular files;
        Mode() FileMode     // file mode bits
        ModTime() time.Time // modification time
        IsDir() bool        // abbreviation for Mode().IsDir()
        Sys() any           // underlying data source (can return nil)
}

//删掉已经存在的AOF文件
func Remove(name string)

//从AOF文件的某个位置读出len(b)这么多数据
func (f *AOF) ReadAt(b []byte, off int64) (n int, err error)

//AOF文件的追加写的接口，返回写入的位置
func (f *AOF) Append(b []byte) int64

//关闭AOF文件
func (f *AOF) Close() error

```

## BDEV快速操作入门


(结束)

