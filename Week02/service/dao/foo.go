package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

func DaoCanNotHandleErr() (interface{}, error) {
	nr := sql.ErrNoRows                             //产生了一个sql错误
	sqldata := 1                                    //假如这是从sql中读取的数据
	_ = sqldata                                     //	必须将读取出来的值舍去
	return nil, fmt.Errorf("DAO层处理不了这个错误了： %w", nr) //如果dao层不能处理这个错误，一定要对值负责
}

func DaoHandleErr() (interface{}, error) {
	nr := sql.ErrNoRows //产生了一个sql错误
	sqldata := 1        //假如这是从sql中读取的数据
	_ = sqldata         //	必须将读取出来的值舍去

	if errors.Is(nr, sql.ErrNoRows) {
		defaultdata := 0
		log.Println("做降级处理，但是：", nr) //必须打日志
		return defaultdata, nil      //dao层打算降级处理，那么不要继续往上抛错误了
	}

	log.Println("返回空值，但是：", nr) //仍然吞掉，但是返回的值是nil
	return nil, nil


}

/*
助教好，我是认真学习了error这一章的，可以看看我做的笔记，但是我只能写成这样了，因为我发现我不懂：
1. 我觉得dao层、service层、api层什么的，不过就是看在哪一层需要向上屏蔽错误细节了，那么就在那一层吞掉错误
	但是我不知道哪一层需要向上屏蔽细节，所以我在这里将dao层能处理和不能处理的两种情况都写了。
2. 我学到这里，我觉得如果不使用pkg/errors包的话，只有fmt.Errorf()加上格式化参数%w能达到wrap类似的效果，
	但是没有堆栈信息！！！
	所以dao层如果向上抛错误的话，比如在service层,是不是要加上这个错误是从dao层抛上来的堆栈信息呢？
3.
*/