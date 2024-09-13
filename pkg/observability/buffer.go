package observability

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/in4it/wireguard-server/pkg/logging"
)

func (o *Observability) WriteBufferToStorage(n int64) error {
	o.ActiveBufferWriters.Add(1)
	defer o.ActiveBufferWriters.Done()
	// copy first to temporary buffer (storage might have latency)
	tempBuf := bytes.NewBuffer(make([]byte, n))
	_, err := io.CopyN(tempBuf, o.Buffer, n)
	o.LastFlushed = time.Now()
	if err != nil && err != io.EOF {
		return fmt.Errorf("write error from buffer to temporary buffer: %s", err)
	}

	file, err := o.Storage.OpenFileForWriting("data-" + time.Now().Format("2003-01-02T15:04:05") + "-" + strconv.FormatUint(o.FlushOverflowSequence.Add(1), 10))
	if err != nil {
		return fmt.Errorf("open file for writing error: %s", err)
	}
	_, err = io.Copy(file, tempBuf)
	if err != nil {
		return fmt.Errorf("file write error: %s", err)
	}
	return file.Close()
}

func (o *Observability) monitorBuffer() {
	for {
		time.Sleep(FLUSH_TIME_MAX_MINUTES * time.Minute)
		if time.Since(o.LastFlushed) >= (FLUSH_TIME_MAX_MINUTES * time.Minute) {
			if o.FlushOverflow.CompareAndSwap(false, true) {
				err := o.WriteBufferToStorage(int64(o.Buffer.Len()))
				o.FlushOverflow.Swap(true)
				if err != nil {
					logging.ErrorLog(fmt.Errorf("write log buffer to storage error: %s", err))
					continue
				}
			}
			o.LastFlushed = time.Now()
		}
	}
}

func (o *Observability) Ingest(data io.ReadCloser) error {
	defer data.Close()
	msgs, err := Decode(data)
	if err != nil {
		return fmt.Errorf("decode error: %s", err)
	}
	_, err = o.Buffer.Write(encodeMessage(msgs))
	if err != nil {
		return fmt.Errorf("write error: %s", err)
	}
	if o.Buffer.Len() >= o.MaxBufferSize {
		if o.FlushOverflow.CompareAndSwap(false, true) {
			go func() { // write to storage
				if n := o.Buffer.Len(); n >= o.MaxBufferSize {
					err := o.WriteBufferToStorage(int64(n))
					if err != nil {
						logging.ErrorLog(fmt.Errorf("write log buffer to storage error (buffer: %d): %s", o.Buffer.Len(), err))
					}
				}
				o.FlushOverflow.Swap(false)
			}()
		}
	}
	return nil
}

func (c *ConcurrentRWBuffer) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buffer.Write(p)
}
func (c *ConcurrentRWBuffer) Read(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buffer.Read(p)
}
func (c *ConcurrentRWBuffer) Len() int {
	return c.buffer.Len()
}
func (c *ConcurrentRWBuffer) Cap() int {
	return c.buffer.Cap()
}
