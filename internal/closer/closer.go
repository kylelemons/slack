package closer

import (
	"io"
	"sync"
)

type Once struct {
	once sync.Once
	err  error

	closer io.Closer
}

func (c *Once) Close() error {
	c.once.Do(func() {
		c.err = c.closer.Close()
	})
	return c.err
}

func For(c io.Closer) Once {
	return Once{closer: c}
}
