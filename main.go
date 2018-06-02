package main

import (
	"fmt"
	"sync"
	"time"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"github.com/coraxster/PriorityQueue"
)

type Job struct {
	str string
}

func main()  {
	fmt.Println("Start!")
	InChs := make([]chan interface{}, 0)
	for priority := 0; priority <= 10 ; priority++  {
		InChs = append(InChs, make(chan interface{}, 5))
	}

	//chan2out, _ := SimplePrioritizeChans(InChs...) // pretty simple but blocks in-channels
	chan2out, _ := PriorityQueue.Prioritize(InChs...) // doesn't block in-channels

	work(chan2out)

	for priority := 10; priority > 0 ; priority--  {
		go func(priority int){
			for i:=1; i<100; i++ {
				InChs[priority] <- "hi - pr:" + strconv.Itoa(priority) + " (" + strconv.Itoa(i) + ")"
			}
		}(priority)

	}


	sigs := make(chan os.Signal, 1)
	signal.Notify(make(chan os.Signal, 1), syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("exiting")
}




func work(ch chan interface{}){
	var m sync.Mutex
	doneCount := 0
	for i := 1; i <= 3; i++ {
		go func(i int) {
			for item := range ch{
				data := item.(string)
				m.Lock()
				fmt.Println(strconv.Itoa(doneCount) + ". " + data + " - worker " + strconv.Itoa(i))
				doneCount++
				m.Unlock()
				time.Sleep(time.Millisecond * 500 )
			}
		}(i)
	}
}

func SimplePrioritizeChans(ins... chan interface{}) (chan interface{}, error) {
	out := make(chan interface{})
	go func() {
		for {
		switcher: for _, ch := range ins {
			select {
			case item := <- ch:
				out <- item
				break switcher
			default:

			}
		}
		}
	}()
	return out, nil
}