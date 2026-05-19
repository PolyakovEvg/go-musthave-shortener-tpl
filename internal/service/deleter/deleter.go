package service

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/logger"
	"errors"
	"sync"
	"time"
)

var ErrQueueFull = errors.New("delete queue is full")

type DeleteTask struct {
	UserID string
	IDs    []string
}

type Deleter struct {
	logger       *logger.Logger
	markFunc     func(userID string, shorts []string) error
	fanIn        chan DeleteTask
	maxBatchSize int
	batchTimeout time.Duration

	done chan struct{}
	wg   sync.WaitGroup
}

func NewDeleter(markFunc func(userID string, shorts []string) error, logger *logger.Logger) *Deleter {
	d := &Deleter{
		logger:       logger,
		markFunc:     markFunc,
		fanIn:        make(chan DeleteTask, 1024),
		maxBatchSize: 100,
		batchTimeout: 200 * time.Millisecond,
		done:         make(chan struct{}),
	}

	d.wg.Add(1)
	go d.batchWorker()
	return d
}

func (d *Deleter) Enqueue(task DeleteTask) error {
	select {
	case d.fanIn <- task:
		return nil
	default:
		return ErrQueueFull
	}
}

func (d *Deleter) batchWorker() {
	defer d.wg.Done()

	buffer := make([]DeleteTask, 0, d.maxBatchSize)
	timer := time.NewTimer(d.batchTimeout)
	defer timer.Stop()

	flush := func(tasks []DeleteTask) {
		if len(tasks) == 0 {
			return
		}
		group := make(map[string][]string)
		for _, t := range tasks {
			group[t.UserID] = append(group[t.UserID], t.IDs...)
		}
		for uid, ids := range group {
			err := d.markFunc(uid, ids)

			if err != nil {
				if d.logger != nil {
					d.logger.Zap.Errorw("failed to delete URLs",
						"userID", uid,
						"ids", ids,
						"error", err,
					)
				}
			} else {
				if d.logger != nil {
					d.logger.Zap.Infow("successfully deleted URLs",
						"userID", uid,
						"ids_count", len(ids),
					)
				}
			}
		}
	}

	for {
		timer.Reset(d.batchTimeout)
		select {
		case <-d.done:
			flush(buffer)
			return
		case t := <-d.fanIn:
			buffer = append(buffer, t)
			if len(buffer) >= d.maxBatchSize {
				flush(buffer)
				buffer = buffer[:0]
			}
		case <-timer.C:
			if len(buffer) > 0 {
				flush(buffer)
				buffer = buffer[:0]
			}
		}
	}
}

func (d *Deleter) Close() {
	close(d.done)
	d.wg.Wait()
}
