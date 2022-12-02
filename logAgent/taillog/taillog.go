package taillog

import (
	"context"
	"fmt"
	"github.com/hpcloud/tail"
	"io"
	"logAgent/kafka"
)

type TailTask struct {
	path     string
	topic    string
	instance *tail.Tail

	// 用于退出 TailTask.run()
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewTailTask(path, topic string) (t *TailTask) {
	ctx, cancel := context.WithCancel(context.Background())
	t = &TailTask{
		path:       path,
		topic:      topic,
		ctx:        ctx,
		cancelFunc: cancel,
	}
	err := t.Init()
	if err != nil {
		fmt.Println("tail file failed, err:", err)
	}
	return
}

func (t *TailTask) Init() (err error) {
	config := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd},
		MustExist: false,
		Poll:      true,
	}
	t.instance, err = tail.TailFile(t.path, config)
	go t.run()
	return
}

func (t *TailTask) run() {
	for {
		select {
		case <-t.ctx.Done():
			fmt.Printf("tail task:[%s_%s] finish...\n", t.path, t.topic)
			return
		case line := <-t.instance.Lines:
			fmt.Printf("get log data from %s success, log: %v\n", t.path, line.Text)
			kafka.SendToChan(t.topic, line.Text)
		}
	}
}
