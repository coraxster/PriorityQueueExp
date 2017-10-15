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
		InChs = append(InChs, make(chan interface{}))
	}

	chan2out, _ := PriorityQueue.Prioritize(InChs...)
	work(chan2out)

	for priority := 10; priority > 0 ; priority--  {
		for i:=1; i<100; i++ {
			InChs[priority] <- "hi - pr:" + strconv.Itoa(priority) + " (" + strconv.Itoa(i) + ")"
		}
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







//Legacy impl
func prioritise(in chan *itemWithPriority) (chan *itemWithPriority, error) {
	var queues [][]*itemWithPriority
	var m sync.Mutex
	out := make(chan *itemWithPriority)

	go func() {
		var dataOut *itemWithPriority
		for{
			found := false
			m.Lock()
			for priority, queue := range queues {
				if len(queue) > 0 {
					dataOut, queues[priority] = queue[0], queue[1:]
					found = true
					break
				}
			}
			m.Unlock()
			if found {
				out <- dataOut
			}
		}
	}()

	go func() {
		for {
			dataIn := <- in
			m.Lock()
			if dataIn.priority >= len(queues) {
				newQueues := make([][]*itemWithPriority, dataIn.priority+1, dataIn.priority * 2)
				copy(newQueues, queues)
				queues = newQueues
			}
			//c := cap(queues[dataIn.priority])
			//l := len(queues[dataIn.priority])
			//if c < l + 1 {
			//	newSlice := make([]itemWithPriority, l, c * 2)
			//	copy(newSlice, queues[dataIn.priority])
			//	queues[dataIn.priority] = newSlice
			//}
			queues[dataIn.priority] = append(queues[dataIn.priority], dataIn)
			m.Unlock()
		}
	}()

	return out, nil
}

type itemWithPriority struct {
	i int
	index int
	priority int
	item     interface{}
}
