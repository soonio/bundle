# BUNDLE

> 一个练手项目，打包器。  
> 主要使用的技术chan，for + select case，范型
> 主要过程接收请求参数到 chan 队列中，然后当队列中数据每达到一定量的时候，合并多个调用handle

## 使用场景

  积累一定的任务数量，然后进行批量的处理，比如批量写入

## Install

```bash
go get -u github.com/soonio/bundle
```

## Usage

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "github.com/soonio/bundle"
)

// Form 表单参数
type Form struct {
    Name string
}

func main() {
    var b = bundle.New(
        func(ts []*Form) {
            var names []string
            for _, t := range ts {
                names = append(names, t.Name)
            }
            fmt.Println(strings.Join(names, "|"))
        },
        bundle.WithSize[*Form](30),
        bundle.WithTimeout[*Form](3*time.Second),
        bundle.WithPayloadSize[*Form](1000),
    )

    b.Start()
    defer b.Close()

    go func() {
        for i := 0; i < 100003; i++ {
            go func(i int) {
                b.Add(&Form{Name: fmt.Sprintf("%6d", i)})
            }(i)
            //if i%2 == 0 {
            //	time.Sleep(time.Millisecond)
            //}
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
}

```

## Licenses
    
  MIT

## 谢谢Jetbrain的Goland