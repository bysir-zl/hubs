// 简单的协程池
package util

import "time"

type Pool struct {
	work    chan func()
	sem     chan struct{}
	workers []*worker
}

type worker struct {
	lastUseTime int64
	stop        chan struct{}
	task        func()
	work        chan func()
	done        chan struct{}
}

func (p *Pool) Schedule(task func()) error {
	select {
	case p.work <- task:
	case p.sem <- struct{}{}:
		worker := newWorker(task, p.work, p.sem)
		p.workers = append(p.workers, worker)
	}
	return nil
}

func (p *Pool) clear() {
	for range time.Tick(5 * time.Second) {
		now := time.Now().Unix()
		temp := p.workers[:0]
		for _, w := range p.workers {
			// 一分钟没使用就关闭
			if w.lastUseTime < now-60 {
				w.Stop()
			} else {
				temp = append(temp, w)
			}
		}
		p.workers = temp
	}

}

func (p *Pool) WorkGoCount() int {
	return len(p.sem)
}

// worker

func (w *worker) Run() {
	for {
		w.task()
		w.lastUseTime = time.Now().Unix()
		select {
		case w.task = <-w.work:
			continue
		case <-w.stop:
			<-w.done
			return
		}
	}
}

func (w *worker) Stop() {
	close(w.stop)
}

func newWorker(task func(), work chan func(), done chan struct{}) *worker {
	worker := &worker{
		stop:        make(chan struct{}),
		lastUseTime: time.Now().Unix(),
		task:        task,
		work:        work,
		done:        done,
	}
	go worker.Run()
	return worker
}

func NewPool(size int) *Pool {
	pool := &Pool{
		work:    make(chan func()),
		sem:     make(chan struct{}, size),
		workers: make([]*worker, 0),
	}
	go pool.clear()
	return pool
}

// 一个连接有3-4个Go, 所以这里大约能支持100w在线
var GoPool = NewPool(10000 * 3 * 100)
