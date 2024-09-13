package observability

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/in4it/wireguard-server/pkg/logging"
)

func (o *Observability) WriteBufferToStorage(n int64) error {
	o.BufferMu.Lock()
	defer o.BufferMu.Unlock()
	file, err := o.Storage.OpenFileForWriting("data-" + time.Now().Format("2003-01-02T15:04:05") + "-" + strconv.FormatUint(o.FlushOverflowSequence.Add(1), 10))
	if err != nil {
		return fmt.Errorf("open file for writing error: %s", err)
	}
	_, err = io.CopyN(file, &o.Buffer, n)
	if err != nil {
		return fmt.Errorf("file write error: %s", err)
	}
	o.LastFlushed = time.Now()
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
	fmt.Printf("Buffer size: %d\n", o.Buffer.Len())
	if o.Buffer.Len() >= MAX_BUFFER_SIZE {
		if o.FlushOverflow.CompareAndSwap(false, true) {
			go func() { // write to storage
				if n := o.Buffer.Len(); n >= MAX_BUFFER_SIZE {
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
