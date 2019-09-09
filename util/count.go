package util

import "sync/atomic"

type Count int32

func (c *Count) Inc() int32 {
	//java 的是拷贝到工作内存+1，再放回去
	//这里是直接原子修改公用变量
	return atomic.AddInt32((*int32)(c), 1)
}

func (c *Count) IncN(num int32) int32 {
	//java 的是拷贝到工作内存+1，再放回去
	//这里是直接原子修改公用变量
	return atomic.AddInt32((*int32)(c), num)
}

func (c *Count) Dec() int32 {
	//java 的是拷贝到工作内存+1，再放回去
	//这里是直接原子修改公用变量
	return atomic.AddInt32((*int32)(c), -1)
}

func (c *Count) Get() int32 {
	return atomic.LoadInt32((*int32)(c))
}
