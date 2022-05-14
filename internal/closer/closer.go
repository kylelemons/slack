// Copyright 2022 Kyle Lemons
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
