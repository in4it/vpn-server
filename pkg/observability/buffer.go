package observability

import (
	"fmt"
	"io"
	"time"
)

func (o *Observability) WriteBufferToStorage() error {
	file, err := o.Storage.OpenFileForWriting("data-" + time.Now().Format("2003-01-02T02T15:04:05"))
	if err != nil {
		return fmt.Errorf("open file for writing error: %s", err)
	}
	_, err = io.Copy(file, &o.Buffer)
	if err != nil {
		return fmt.Errorf("file write error: %s", err)
	}
	return file.Close()
}
