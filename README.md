
# AOB: 基于Append-Only底层存储接口的块设备

## 背景

这是一个用于学习存储技术练习项目。Append-Only语义的接口，在存储系统中很常见。
但是很多上层应用不能直接使用Append-Only语义的存储接口，一般要求update-in-place
接口。则就需要把append-only接口转换成update-in-place接口，或者说，基于append-only
接口实现update-in-place接口。

append-only接口指的是，只能在文件或者对象尾部添加数据，已经写入了文件或者对象的
数据，不能修改，只能把文件或者对象整个删除掉才能释放已经占用空间。append-only的
文件或者对象，支持随机读，也就是可以读取任何位置的数据。update-in-place接口，则
可以修改文件和对象中任何位置的数据。

## 底层append-only接口

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


