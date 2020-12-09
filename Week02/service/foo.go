package service

import (
	"./dao"
	"database/sql"
	"errors"
)

var nr  = sql.ErrNoRows//定义一个sentinel error

func ServiceHandleErr()(interface{},error){
	data,err := dao.DaoCanNotHandleErr()//这里dao层处理不了错误的话，将这个错误抛上来了
	if errors.Is(err,nr){
		_ = data//舍弃data
		defaultdata := 0
		return defaultdata,nil //这里做降级处理
	}
	return data,nil
}
