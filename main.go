package main

import (
	"fmt"
	"sync"
	"time"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	//"math/rand"
	"container/heap"
	//"github.com/coraxster/PQueue"
)




func main()  {

	chan1in := make(chan *itemWithPriority)
	chan2out, _ := prioritise2 (chan1in)
	work(chan2out)

	sl := make([]*itemWithPriority, 0, 1000)
	for priority := 10; priority > 0 ; priority--  {
		for i:=1; i<100; i++ {
			item := itemWithPriority{
				i: i,
				item:"hi - pr:" + strconv.Itoa(priority) + " (" + strconv.Itoa(i) + ")",
				priority: priority,
			}
			sl = append(sl, &item)
		}
	}

	for _, item := range sl {
		chan1in <- item
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(make(chan os.Signal, 1), syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("exiting")
}



func work(ch chan *itemWithPriority){
	var m sync.Mutex
	doneCount := 0
	for i := 1; i <= 3; i++ {
		go func(i int) {
			for item := range ch{
				data := item.item.(string)
				m.Lock()
				fmt.Println(strconv.Itoa(doneCount) + ". " + data + " - worker " + strconv.Itoa(i))
				doneCount++
				m.Unlock()
				time.Sleep(time.Microsecond * 5 )
			}
		}(i)
	}
}


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

func prioritise2(in chan *itemWithPriority)(chan *itemWithPriority, error)  {
	out := make(chan *itemWithPriority)
	var m sync.Mutex
	pq := make(Queue, 0)
	heap.Init(&pq)

	go func() {
		for item := range in {
			m.Lock()
			heap.Push(&pq, item)
			m.Unlock()
		}
	}()

	go func() {
		for{
			if pq.Len() > 0 {
				m.Lock()
				item := heap.Pop(&pq).(*itemWithPriority)
				m.Unlock()
				out <- item
			}
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

// A Queue implements heap.Interface and holds Items.
type Queue []*itemWithPriority

func (pq Queue) Len() int { return len(pq) }

func (pq Queue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	if pq[i].priority == pq[j].priority {
		return pq[i].i < pq[j].i
	}
	return pq[i].priority < pq[j].priority
}

func (pq Queue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *Queue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*itemWithPriority)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *Queue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *Queue) update(item *itemWithPriority, value string, priority int) {
	item.item = value
	item.priority = priority
	heap.Fix(pq, item.index)
}