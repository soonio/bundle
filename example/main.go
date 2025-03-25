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
		bundle.WithSize[*Form](10),
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
