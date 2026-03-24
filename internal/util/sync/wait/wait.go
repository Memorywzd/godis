package wait

import (
	"sync"
	"time"
)

// 定义一个可以等待一定时间的 WaitGroup
type Wait struct {
	wg sync.WaitGroup
}

// 向WaitGroup计数器增加delta（可以为负数）
func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

// 向WaitGroup计数器减少1（相当于 Add(-1) ）
func (w *Wait) Done() {
	w.wg.Done()
}

// 等待WaitGroup计数器变为0
func (w *Wait) Wait() {
	w.wg.Wait()
}

// 等待WaitGroup计数器变为0，或者等待超时，超时返回 true
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	c := make(chan struct{}, 1)
	go func() {
		defer close(c)
		w.wg.Wait()
		c <- struct{}{}
	}()

	select {
	case <-c:
		return false // 正常完成，未超时
	case <-time.After(timeout):
		return true // 超时
	}
}
