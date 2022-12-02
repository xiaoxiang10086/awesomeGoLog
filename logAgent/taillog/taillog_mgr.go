package taillog

import (
	"fmt"
	"logAgent/etcd"
	"time"
)

var taskMrg *TailLogMgr

type TailLogMgr struct {
	logEntry    []*etcd.LogEntry
	taskMap     map[string]*TailTask
	newConfChan chan []*etcd.LogEntry
}

func Init(logEntryConf []*etcd.LogEntry) {
	taskMrg = &TailLogMgr{
		logEntry:    logEntryConf,
		taskMap:     make(map[string]*TailTask, 16),
		newConfChan: make(chan []*etcd.LogEntry), // 无缓冲区的通道
	}
	for _, logEntry := range logEntryConf {
		tailObj := NewTailTask(logEntry.Path, logEntry.Topic)
		mk := fmt.Sprintf("%s_%s", logEntry.Path, logEntry.Topic)
		taskMrg.taskMap[mk] = tailObj
	}
	go taskMrg.run()
}

func (t *TailLogMgr) run() {
	for {
		select {
		case newConf := <-t.newConfChan:
			// 1. 配置新增
			for _, conf := range newConf {
				mk := fmt.Sprintf("%s_%s", conf.Path, conf.Topic)
				_, ok := t.taskMap[mk]
				if ok {
					continue
				} else {
					tailObj := NewTailTask(conf.Path, conf.Topic)
					t.taskMap[mk] = tailObj
				}
			}

			for _, c1 := range t.logEntry {
				isDelete := true
				for _, c2 := range newConf {
					if c2.Path == c1.Path && c2.Topic == c1.Topic {
						isDelete = false
						continue
					}
				}

				// 2. 配置删除
				if isDelete {
					mk := fmt.Sprintf("%s_%s", c1.Path, c1.Topic)
					t.taskMap[mk].cancelFunc()
				}
			}

			fmt.Println("new config is arrived！", newConf)
		default:
			time.Sleep(time.Second)
		}
	}
}

// NewConfChan 向外暴露 taskMgr 的 newConfChan
func NewConfChan() chan<- []*etcd.LogEntry {
	return taskMrg.newConfChan
}
