package main

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
)

func main(){
	g :=new(errgroup.Group)//创建一个group对象

	http.HandleFunc(`/`,func(rw http.ResponseWriter,r *http.Request){
		g.Go(func() error {//使用errgroup提供的go函数去开协程处理http请求
			/*处理请求*/
			return err
		})
	})

	if err := g.Wait(); err == nil {//wait函数返回第一个产生的错误
		fmt.Println("所有http请求处理完毕，安全退出")
	}
}
/*
助教老师，这周的内容对于我来说太硬核，我甚至连派goroutine去处理http请求的代码也模仿煎鱼助教的书写过一次，
我从网上看了几篇文章，然后看了一下errgroup例子，这里只是拙劣的模仿

我觉得就三步：
1、创建group对象
2、每当有http请求，就使用Go方法派goroutine去处理
3、在主协程中使用Wait方法做同步，直到所有的子协程处理完毕，或者有一个报错，那么wait返回，不再阻塞


第二项作业我更是linux signal也没有听过，所以实在没有办法在不到一周的时间补上基础，这里没法写。。。
*/