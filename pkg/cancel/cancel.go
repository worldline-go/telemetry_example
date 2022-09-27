package cancel

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type Function struct {
	Fn   func() error
	Name string
}

type Registry struct {
	functions   []Function
	mutex       sync.Mutex
	isCancelled bool
}

func (r *Registry) Register(fn ...Function) *Registry {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.functions = append(r.functions, fn...)

	return r
}

func (r *Registry) Cancel() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// cancel opposite
	for i := len(r.functions) - 1; i >= 0; i-- {
		if err := r.functions[i].Fn(); err != nil {
			log.Error().Err(err).Msg(r.functions[i].Name)
		}
	}

	r.isCancelled = true
}

func (r *Registry) IsCancelled() bool {
	return r.isCancelled
}
