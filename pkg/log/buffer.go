package log

import (
	"bytes"
	"fmt"
	"sync"
	"time"
)

const (
	defaultBatchSize    = 1024 * 4
	defaultFlushTimeout = time.Second
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, defaultBatchSize))
	},
}

// BufferedWriteSyncer 带缓冲的写入器
type BufferedWriteSyncer struct {
	sync.Mutex
	underlying *DateWriteSyncer
	buf        *bytes.Buffer
	timer      *time.Timer
	size       int
}

// NewBufferedWriteSyncer 创建带缓冲的写入器
func NewBufferedWriteSyncer(ws *DateWriteSyncer) *BufferedWriteSyncer {
	b := &BufferedWriteSyncer{
		underlying: ws,
		buf:        bufferPool.Get().(*bytes.Buffer),
	}
	b.timer = time.AfterFunc(defaultFlushTimeout, b.Flush)
	return b
}

func (b *BufferedWriteSyncer) Write(p []byte) (n int, err error) {
	b.Lock()
	defer b.Unlock()

	n, err = b.buf.Write(p)
	if err != nil {
		return n, err
	}
	b.size += n

	if b.size >= defaultBatchSize {
		return n, b.FlushLocked()
	}
	return n, nil
}

// Flush 刷新缓冲区到文件
func (b *BufferedWriteSyncer) Flush() {
	b.Lock()
	defer b.Unlock()
	err := b.FlushLocked()
	if err != nil {
		// 使用标准库log记录清理错误，避免循环依赖
		fmt.Printf("flush log failed: %v\n", err)
	}
}

// FlushLocked 在已获得锁的情况下刷新缓冲区
func (b *BufferedWriteSyncer) FlushLocked() error {
	if b.buf.Len() == 0 {
		return nil
	}

	_, err := b.underlying.Write(b.buf.Bytes())
	b.buf.Reset()
	b.size = 0
	b.timer.Reset(defaultFlushTimeout)

	return err
}

// Sync 实现 zapcore.WriteSyncer 接口
func (b *BufferedWriteSyncer) Sync() error {
	b.Lock()
	defer b.Unlock()

	if err := b.FlushLocked(); err != nil {
		return err
	}
	return b.underlying.Sync()
}

// Close 关闭写入器
func (b *BufferedWriteSyncer) Close() error {
	b.Lock()
	defer b.Unlock()

	b.timer.Stop()
	if err := b.FlushLocked(); err != nil {
		return err
	}
	bufferPool.Put(b.buf)
	return nil
}