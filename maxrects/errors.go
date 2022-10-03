package maxrects

import (
	"errors"
	"strings"
	"sync"
)

type errorGroup struct {
	errors []error
	mu     sync.Mutex
}

func (group *errorGroup) Empty() bool {
	return group.errors == nil || len(group.errors) == 0
}

func (group *errorGroup) Collect() error {
	errs := make([]string, len(group.errors))

	for i, err := range group.errors {
		errs[i] = err.Error()
	}

	return errors.New(strings.Join(errs, "\n"))
}

func (group *errorGroup) Add(err error) {
	group.mu.Lock()
	defer group.mu.Unlock()

	group.errors = append(group.errors, err)
}

func (group *errorGroup) Reset() {
	group.errors = nil
}
