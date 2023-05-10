package gtable

import (
	"log"
	"sync/atomic"
)

var (
	mapCustom map[string]atomic.Value
)

func init() {
	mapCustom = make(map[string]atomic.Value)
}

func getCustom(name string) *atomic.Value {
	v, ok := mapCustom[name]
	if ok {
		return &v
	}
	return nil
}

func storeCustom(name string, val interface{}) {
	if hot {
		atomicVal := getCustom(name)
		if atomicVal == nil {
			log.Fatalf("custom table %s not exist!", name)
		} else {
			var atomicVal atomic.Value
			atomicVal.Store(val)
			mapCustom[name] = atomicVal
		}
	} else {
		var atomicVal atomic.Value
		atomicVal.Store(val)
		mapCustom[name] = atomicVal
	}
}
