package file

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish"
)

type FilePublisher struct {
	file   *os.File
	name   string
	closed bool
}

func NewFilePublisher(opts publish.FilePublisherOption) (*FilePublisher, error) {
	if len(opts.Name) == 0 {
		return nil, errors.New("empty log name")
	}
	file, err := os.OpenFile(opts.Name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return &FilePublisher{file: file, name: opts.Name, closed: false}, nil
}

func (f *FilePublisher) Publish(event *schema.Event) error {
	if f.closed {
		return nil
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if _, err := os.Stat(f.name); err != nil {
		if os.IsNotExist(err) {
			f.file.Close()
			file, err := os.OpenFile(f.name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			f.file = file
		}
		return err
	}

	if f.closed {
		return nil
	}
	_, err = f.file.Write(bytes)
	if err != nil {
		return err
	}
	f.file.Write([]byte("\n"))
	return nil
}

func (f *FilePublisher) Close() {
	f.closed = true
	f.file.Close()
}
