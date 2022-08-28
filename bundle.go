package bundle

import (
	"sync"
	"time"
)

const (
	defaultSize    = 20
	defaultTimeout = 10 * time.Second
	defaultPCap    = 1000
)

//操作	    空值(nil)	非空已关闭	非空未关闭
//关闭		panic		panic		成功关闭
//发送数据	永久阻塞		panic		阻塞或成功发送
//接收数据	永久阻塞		永不阻塞		阻塞或者成功接收

func New[T any](handler func([]T), opt ...Apply[T]) *bundle[T] {
	b := &bundle[T]{
		size:    defaultSize,
		timeout: defaultTimeout,
		close:   make(chan struct{}),
		handler: handler,
	}
	for _, apply := range opt {
		apply(b)
	}

	b.timer = time.NewTimer(b.timeout)
	if b.payloads == nil {
		b.payloads = make(chan T, defaultPCap)
	}

	b.do = make(chan struct{}, cap(b.payloads)/b.size)

	return b
}

type bundle[T any] struct {
	count    int           // 计数是否达到阈值
	size     int           // 打包阈值
	payloads chan T        // 载荷
	timeout  time.Duration // 超时时间
	timer    *time.Timer   // 计时器
	close    chan struct{} // 关闭信号
	do       chan struct{} // 打包信号
	lock     sync.Mutex    // 计数锁
	handler  func([]T)     // 打包后处理回调
}

// Add 添加一个载荷
func (b *bundle[T]) Add(payload T) {
	b.payloads <- payload
	b.lock.Lock()
	defer b.lock.Unlock()
	b.count = (b.count + 1) % b.size
	if b.count+1 == b.size {
		b.do <- struct{}{}
	}
}

// Start 启动服务
func (b *bundle[T]) Start() {
	go b.working()
}

func (b *bundle[T]) working() {
	defer func() { // 完成关闭并通知
		b.close <- struct{}{}
	}()
	defer func() { // 关闭前把剩余未打包的进行打包操作
		for len(b.payloads) > 0 {
			b.pack()
		}
	}()
	for {
		select {
		case <-b.do: // 收到打包信号
			b.pack()
			b.timer.Reset(b.timeout)
		case <-b.timer.C: // 收到超时信号
			b.pack()
			b.timer.Reset(b.timeout)
		case <-b.close: // 收到关闭信号
			return
		}
	}
}

// 执行分组打包
func (b *bundle[T]) pack() {
	l := len(b.payloads)
	if l > 0 {
		var size = l
		if size > b.size {
			size = b.size
		}
		var ts = make([]T, size)
		for i := 0; i < size; i++ {
			ts[i] = <-b.payloads
			l--
		}
		b.handler(ts)
	}
}

func (b *bundle[T]) Close() {
	close(b.payloads)     // 关闭payloads
	b.timer.Stop()        // 关闭计时器
	b.close <- struct{}{} // 发送关闭信号
	<-b.close             // 等待关闭完成
	close(b.close)        // 销毁关闭信号chan
	close(b.do)           // 销毁执行任务信号chan
}
