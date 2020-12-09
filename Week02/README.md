# 第二周的内容：error
需要的前置知识：

pkg/errors这个包

官方的errors包

ppt中的reference

effective go

go语言圣经的：函数 这一章
---
使用errors.New() 来返回一个包级别的错误变量
```go
type error interface{
  Error() string
}

//一个error对象到底是什么呢：
type errorstring struct{
  s string
}//就是一个包含了错误信息的结构体

//我们使用New()函数来创建一个新型错误，具体是怎么实现的呢？
//注意这里是新型，既即使错误信息一样，但是名称不一样，同样是不同的错误

func New(text string)error{
  return & errorstring{text}//初始化一个新的结构体，将信息写进去，并取地址
}
```

在go中实现了多返回值和error接口，这是go处理异常的机制

---
panic 慎重使用！！！
一般来说，强依赖的部分没有实现，panic
配置文件明显错误，panic

不要在野生goroutine中写panic，创建出来的goroutine你一定要搞懂他的生命周期

强弱依赖是根据具体业务逻辑来决定的

业务判断语句不要用bool返回失败含义（因为有二义性），规规矩矩使用error

索引越界、栈溢出—>panic

**总结：go的错误处理机制，在写代码的时候要从发生的错误开始考虑；没有隐藏的控制流；error是一个值类型**
---
## error的类型

sentinel error(预定义的错误)

想要提供上下文检查的时候会破坏相等性检查

一般来说，可以在基础库中用，声明成包级别的变量，但是体量小的话还是算了，不然你写的标准文档都没人遵守
```go
var (
  ErrInvalidUnreadByte = errors.New("bufio: invalid use of UnreadByte")
)//在包级别定义大量的sentinel error
```

最糟糕的莫过于在两个包之间创建了依赖
---
error type：是实现了error接口的自定义类型
```go
type MyError struct{
  Msg string
  File string
  Line int
}

func (e *MyError)Error() string{//实现了error接口
  return fmt.Sprintf("%s:%d: %s",e.File,e.Line,e.Msg)
}

func test()error{
  return &MyError{"Something happened","server.go",42}
}

```

使用断言来判断是否是这个类型的错误，并且可以获得更多的上下文信息

但是使用断言（或者类型switch），会让自定义的error变为public。这种模型会导致和调用者产生强耦合，从而导致API变脆弱
---
opaque errors：不透明的错误

只返回错误，而不假设其内容（我们只知道有错误发生，或者nil）

但是我们如果想调查错误的性质，或者与外界交互。我们需要::断言错误实现了特定的行为，而不是断言错误是特定的类型或值::

```go
//比如在net包里面
package net

type Error interface{
  error
  Timeout() bool
  Temporary() bool
}

//那么我们在调用的时候，可以使用函数来对行为进行判断
if nerr,ok:=err.(net.Error);ok&&nerr.Temporary(){
  /**/
}

```
我们可以借鉴这种方式，将判断封装
```go
type temporary interface{
  Temporary() bool
}

func IsTemporary(err error)bool{
  te,ok:=err.(temporary)
  return ok && te.Temporary() 
}
```

---
## handling error

这里推荐一个很好的方法：
```go
type errWriter struct{
  io.Writer
  err error
}//将io.Writer封装进结构体中

func (e *errWriter)Write(buf []byte)(int,error){
  if e.err !=nil{
    return 0,e.err
  }
  var n int
  n,e.err := e.Writer.Write(buf)
  return n,nil
}
```

然后在主逻辑中使用结构体声明变量，在最后return变量的err，就可以得到第一个err

---
## 进化！！！：pkg/errors实现的wrap方法！！！
如果你关心值，那么必须对同时返回的err进行处理：

如果你要吞掉这个error，必须：
1. 将错误写进日志
2. 要保证这个值的完整性，要么返回空值，要么做降级处理
3. 之后不再报告该错误

如果不吞掉，千万不能写进日志，要wrap必要信息往上抛：

```go
return nil,err//这是常用的方式，不包含任何上下文信息

return nil,errors.Wrap(err,"openfile failed")//通过wrap函数添加一些信息

return config,errors.WithMessage(err,"could not read config")//只添加一点错误信息，不添加堆栈信息
```

wrap包是怎么实现的呢？
```go

func Wrap(err error, message string) error {

  if err == nil {

    return nil

  }

  err = &withMessage{

    cause: err,

    msg:   message,

  }

  return &withStack{

    err,

    callers(),//调用callers函数返回寄存器里面的堆栈信息（这是运行时的东西，好好去学学）

  }

}

//这是errors源码，优点就在于error接口类型的优越性
```

当我们需要吞掉错误的时候（写log，或者直接打印），errors包也实现了标准打印的方法！！！（errors yysd）

先来看看如果要处理错误的话是怎么打印的
```go
fmt.Printf("original error: %T %v\n",errors.Cause(err),errors.Cause(err))//这里%T打印的是类型，%v打印的是内容，都是打印的根因的堆栈信息（使用Cause调用根因）

fmt.Printf("stack trace:\n%+v\n",err)//实现了%+v打印所有堆栈信息
```

这是怎么实现的呢：
```go

func (w *withStack) Format(s fmt.State, verb rune) {

  switch verb {

  case 'v':

    if s.Flag('+') {

      fmt.Fprintf(s, "%+v", w.Cause())

      w.stack.Format(s, verb)

      return

    }

    fallthrough

  case 's':

    io.WriteString(s, w.Error())

  case 'q':

    fmt.Fprintf(s, "%q", w.Error())

  }

}

////////////////////

func (w *withMessage) Format(s fmt.State, verb rune) {

  switch verb {

  case 'v':

    if s.Flag('+') {

      fmt.Fprintf(s, "%+v\n", w.Cause())

      io.WriteString(s, w.msg)

      return

    }

    fallthrough

  case 's', 'q':

    io.WriteString(s, w.Error())

  }

}
```

pkg/errors这个库也提供了New和Errorf函数，兼容标准库的操作

cause函数返回的根因就可以直接和sentinel error进行比较了

---
## go 1.13之后的版本，将pkg/errors库这里面的思想吸收，加入了自己的errors库（pkg/errors库的contributor也在go team里面）

1.13之前的版本，如果不用调用pkg/errors这个库，是怎么处理错误的：
```go
//1. 不透明处理
if err != nil{
}

//2. 必须要对错误类型进行判断（与包级sentinel error）
if err == foo.ErrNotFound{//会创建依赖
}

//3. 实现了error interface的自定义error struct，通过断言获取更丰富的上下文
type NotFoundError struct{
  Name string
}

func (e *NotFoundError)Error()string{return e.Name + ": not found"}

if e,ok:=err.(*NotfoundError);ok{
  //e.Name wasn't found
}

//4. 使用Errorf返回新的错误信息（会使等值判断失效）
if err !=nil{
  return nil,fmt.Errorf("decompress %v: %v", name,err)
}

//5. 方法3的拓展：将根因也封装进去（其实和现在的思路很像了）
type NewError struct{
  Name string
  Olderr error
}

//实现Error方法

//断言获取根因
if e,ok:=err.(*NewError);ok&&e.Olderr == ErrPermission{
}
```

1.13以后将这些思路和pkg/errors里的思路结合提供了新的方式：

1. errors 包提供了新的函数
2. fmt.Errorf有了新的格式化输出参数

```go
type QueryError struct{
  Name string
  Err error
}

//Unwrap方法返回错误包含的错误
func (e *QueryError)Unwrap()error{return e.Err}

//Is函数检查现在这个错误的根因是不是指定错误,应该说是它包裹的某一层是不是指定错误
if errors.Is(err,ErrNotfound){
}


// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
//
var e *QueryError//新的错误类型

if errors.As(err, &e){
  //err 是一个*QueryError，并且e被设置成了err错误链中存在的那个*QueryError类型的值！
}
```
具体Is和As函数的实现代码不在这里贴出来，但是要注意的是，这两个函数中都使用了递归去找，直到找到错误链中有这种类型的错误

```go
//fmt.Errorf有新的格式化参数帮助生成带上下文，并且不影响等值判断的error类型进行返回

fmt.Errorf("decompress : %w",ErrPermission)

//举例
myerr := errors.New("新的错误类型")
fmt.Println(errors.Is(fmt.Errorf("额外信息 %w",myerr),myerr))
//true
```

---
## 进阶内容：自定义你的Is和As函数，加自己的代码逻辑进去
这一部分过段时间再补吧！

---
```go
nr := sql.ErrNoRows//产生了一个sql错误

return fmt.Errorf("Dao 层处理不了这个错误 ： %w"，nr)


```
