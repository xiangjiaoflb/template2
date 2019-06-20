package utils

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Memory 内存缓存
type Memory struct {
	kv sync.Map // string map time.Time

	size int32 // 数据最多存在多少个

	count int32 //map大小

	cleandata clean

	deletechan chan interface{}

	status int32
}

type clean struct {
	cleanTime time.Duration
	closechan chan bool
	wg        sync.WaitGroup
}

type survivalData struct {
	value interface{}
	time  time.Time
}

// NewMemory 创建内存缓存
// size 数据最多存在多少个
func NewMemory(size int32, arg ...interface{}) *Memory {
	var mm Memory
	mm.deletechan = make(chan interface{}, 100)
	mm.size = size
	mm.cleandata.cleanTime = time.Minute * 10
	if len(arg) != 0 {
		ctime, ok := arg[0].(time.Duration)
		if ok {
			mm.cleandata.cleanTime = ctime
		}
	}

	mm.cleandata.closechan = make(chan bool)

	go mm.clean()
	go mm.delete()

	return &mm
}

type close int

//
func (mm *Memory) Close() {
	oldstatus := atomic.SwapInt32(&mm.status, 1)
	if oldstatus == 1 {
		return
	}
	mm.cleandata.closechan <- true
	mm.deletechan <- close(0)
}

//Store 设置key的过期时间
//survivalTime 存活时间 单位 秒
func (mm *Memory) Store(key interface{}, v interface{}, survivalTime time.Duration) error {
	//先拿一个凳子
	count := atomic.AddInt32(&mm.count, 1)

	value := survivalData{
		value: v,
		time:  time.Now().Add(survivalTime),
	}
	_, load := mm.kv.LoadOrStore(key, value)

	//看自己之前的位置还在不
	if load {
		//位置还在，把刚刚拿的凳子让出来
		atomic.AddInt32(&mm.count, -1)
		//坐在之前的位置上
		mm.kv.Store(key, value)
		return nil
	}

	//运行到这表示之前的位置不在了，那就看凳子还摆的下不
	if count > mm.size {
		//摆不下了则出去
		mm.kv.Delete(key)
		return fmt.Errorf("登录者达到最大量了")
	}

	return nil
}

//Load 获取key是否存在
func (mm *Memory) Load(key interface{}) (v interface{}, ok bool) {
	value, ok := mm.kv.Load(key)
	if !ok {
		return nil, false
	}

	data, ok := value.(survivalData)
	if !ok || !time.Now().Before(data.time) {
		//删除缓存
		mm.Delete(key)
		return nil, false
	}

	return data.value, true
}

//Delete 获取key是否存在
func (mm *Memory) Delete(key interface{}) {
	mm.deletechan <- key
}

//定时清理map
func (mm *Memory) clean() {
	//定时器
	tc := time.NewTicker(mm.cleandata.cleanTime)
	for {
		select {
		case <-mm.cleandata.closechan:
			return
		case nowtime := <-tc.C:
			mm.kv.Range(func(key interface{}, value interface{}) bool {
				data := value.(survivalData)
				if nowtime.After(data.time) {
					//删除缓存
					mm.Delete(key)
				}
				return true
			})
		}
	}
}

//保证删除
func (mm *Memory) delete() {
	for {
		select {
		case inter := <-mm.deletechan:
			if _, ok := inter.(close); ok {
				return
			}

			//判断是否存在
			if _, ok := mm.kv.Load(inter); ok {
				//存在则删除并修改总量
				//删除缓存
				mm.kv.Delete(inter)
				atomic.AddInt32(&mm.count, -1)
			}
		}
	}
}
