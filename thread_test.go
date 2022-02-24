package main

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	"log"
	"sync"
	"testing"
	"time"
)

const (
	runTime = 100
)

func demosFunc() {
	time.Sleep(10 * time.Millisecond)
	fmt.Println("Hello World!")
}

func demoFunc() {
	time.Sleep(time.Duration(1) * time.Millisecond)
}

//antsDefaultCommon 使用默认普通pool
//其实就是使用了普通的pool，为了方便直接使用，在内部已经new了一个普通的pool，
//相当于下面那个新建的过程给你写好了，容量大小和过期时间都用默认的，详细信息可以看源码，里面剥一层就可以看到
//例子三
func antsDefaultCommon() {
	var wg sync.WaitGroup //这里使用等待是为了看出结果，阻塞主线程，防止直接停止，如果在web项目中，就不需要

	defer ants.Release() //退出工作，相当于使用后关闭
	log.Println("start ants work")
	for i := 0; i < runTime; i++ {
		wg.Add(1)
		ants.Submit(func() { //提交函数，将逻辑函数提交至work中执行，这里写入自己的逻辑
			log.Println(i, ":hello")
			time.Sleep(time.Millisecond * 10)
			wg.Done()
		})
	}
	wg.Wait()
	log.Println("stop ants work")
}
func antsCommon() {
	p, _ := ants.NewPool(5, ants.WithExpiryDuration(time.Second)) //新建一个pool对象，其他同上
	defer p.Release()
	for j := 0; j < runTime; j++ {
		_ = p.Submit(func() {
			log.Println(":hello")
			time.Sleep(time.Millisecond * 10)
		})
	}

}

func TestAntsCommon(t *testing.T) {
	antsMarkFuncPut()
}

func myFunc(i interface{}) {
	fmt.Println(i)
}
func antsMarkFuncPut() {
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(10, func(i interface{}) {
		myFunc(i)
		wg.Done()
	})
	defer p.Release()
	for i := 0; i < runTime; i++ {
		wg.Add(1)
		_ = p.Invoke(int32(i))
	}
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", p.Running())
}
