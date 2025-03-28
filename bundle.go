package bundle

import (
	"sync"
	"time"
)

const (
	defaultSize    = 20   // 一个包的大小
	defaultTimeout = 10   // 超时时间
	defaultPCap    = 1000 // 存储数据的队列
)

func New[T any](handler func([]T), opt ...Apply[T]) *Bundle[T] {
	b := &Bundle[T]{
		size:    defaultSize,
		timeout: defaultTimeout * time.Second,
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

type Bundle[T any] struct {
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
func (b *Bundle[T]) Add(payload T) {
	b.payloads <- payload

	b.lock.Lock()
	defer b.lock.Unlock()

	nc := b.count + 1
	b.count = nc % b.size
	if nc == b.size {
		b.do <- struct{}{}
	}
}

// Start 启动服务
func (b *Bundle[T]) Start() {
	go b.working()
}

func (b *Bundle[T]) working() {
	for {
		select {
		case <-b.do: // 信号触发打包
			b.pack()
			b.timer.Reset(b.timeout) // 重置超时器
		case <-b.timer.C: // 超时触发打包
			b.pack()
			b.timer.Reset(b.timeout)
		case <-b.close: // 收到关闭信号
			return
		}
	}
}

// 执行分组打包
func (b *Bundle[T]) pack() {
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

func (b *Bundle[T]) Close() {
	close(b.payloads) // 关闭payloads
	b.timer.Stop()    // 关闭计时器
	close(b.close)    // 销毁关闭信号chan
	close(b.do)       // 销毁执行任务信号chan

	for len(b.payloads) > 0 {
		b.pack()
	}
}
