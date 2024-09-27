package observability

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/storage"
)

func (o *Observability) WriteBufferToStorage(n int64) error {
	o.ActiveBufferWriters.Add(1)
	defer o.ActiveBufferWriters.Done()
	o.WriteLock.Lock()
	defer o.WriteLock.Unlock()
	logging.DebugLog(fmt.Errorf("writing buffer to file. Buffer has: %d bytes", n))
	// copy first to temporary buffer (storage might have latency)
	tempBuf := bytes.NewBuffer(make([]byte, 0, n))
	_, err := io.CopyN(tempBuf, o.Buffer, n)
	if err != nil && err != io.EOF {
		return fmt.Errorf("write error from buffer to temporary buffer: %s", err)
	}
	prefix := o.Buffer.ReadPrefix(n)
	o.LastFlushed = time.Now()

	for _, bufferPosAndPrefix := range mergeBufferPosAndPrefix(prefix) {
		now := time.Now()
		filename := bufferPosAndPrefix.prefix + "/data-" + strconv.FormatInt(now.Unix(), 10) + "-" + strconv.FormatUint(o.FlushOverflowSequence.Add(1), 10)
		err = ensurePath(o.Storage, filename)
		if err != nil {
			return fmt.Errorf("ensure path error: %s", err)
		}
		file, err := o.Storage.OpenFileForWriting(filename)
		if err != nil {
			return fmt.Errorf("open file for writing error: %s", err)
		}
		_, err = io.CopyN(file, tempBuf, int64(bufferPosAndPrefix.offset))
		if err != nil && err != io.EOF {
			return fmt.Errorf("file write error: %s", err)
		}
		logging.DebugLog(fmt.Errorf("wrote file: %s", filename))
		err = file.Close()
		if err != nil {
			return fmt.Errorf("file close error: %s", err)
		}
	}
	return nil
}

func (o *Observability) monitorBuffer() {
	for {
		time.Sleep(FLUSH_TIME_MAX_MINUTES * time.Minute)
		if time.Since(o.LastFlushed) >= (FLUSH_TIME_MAX_MINUTES*time.Minute) && o.Buffer.Len() > 0 {
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
	logging.DebugLog(fmt.Errorf("messages ingested: %d", len(msgs)))
	if len(msgs) == 0 {
		return nil // no messages to ingest
	}
	_, err = o.Buffer.Write(encodeMessage(msgs), FloatToDate(msgs[0].Date).Format(DATE_PREFIX))
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

func (o *Observability) Flush() error {
	// wait until all data is flushed
	o.ActiveBufferWriters.Wait()

	// flush remaining data that hasn't been flushed
	if n := o.Buffer.Len(); n >= 0 {
		err := o.WriteBufferToStorage(int64(n))
		if err != nil {
			return fmt.Errorf("write log buffer to storage error (buffer: %d): %s", o.Buffer.Len(), err)
		}
	}
	return nil
}

func (c *ConcurrentRWBuffer) Write(p []byte, prefix string) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prefix = append(c.prefix, BufferPosAndPrefix{prefix: prefix, offset: len(p)})
	return c.buffer.Write(p)
}
func (c *ConcurrentRWBuffer) Read(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buffer.Read(p)
}
func (c *ConcurrentRWBuffer) ReadPrefix(n int64) []BufferPosAndPrefix {
	c.mu.Lock()
	defer c.mu.Unlock()
	totalOffset := 0
	for k, v := range c.prefix {
		if int64(totalOffset+v.offset) == n {
			part1 := c.prefix[:k+1]
			part2 := make([]BufferPosAndPrefix, len(c.prefix[k+1:]))
			copy(part2, c.prefix[k+1:])
			c.prefix = part2
			return part1
		}
		totalOffset += v.offset
	}
	return nil
}
func (c *ConcurrentRWBuffer) Len() int {
	return c.buffer.Len()
}
func (c *ConcurrentRWBuffer) Cap() int {
	return c.buffer.Cap()
}

func ensurePath(storage storage.Iface, filename string) error {
	base := path.Dir(filename)
	baseSplit := strings.Split(base, "/")
	fullPath := ""
	for _, v := range baseSplit {
		fullPath = path.Join(fullPath, v)
		err := storage.EnsurePath(fullPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func mergeBufferPosAndPrefix(a []BufferPosAndPrefix) []BufferPosAndPrefix {
	bufferPosAndPrefix := []BufferPosAndPrefix{}
	for i := 0; i < len(a); i++ {
		offset := a[i].offset
		for y := i; y+1 < len(a) && a[y].prefix == a[y+1].prefix; y++ {
			offset += a[y+1].offset
			i++
		}
		bufferPosAndPrefix = append(bufferPosAndPrefix, BufferPosAndPrefix{
			prefix: a[i].prefix,
			offset: offset,
		})
	}
	return bufferPosAndPrefix
}
