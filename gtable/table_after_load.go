package gtable

/*
this file is auto generate by table_make.
do NOT edit plz!
*/

import (
	"sync/atomic"
)

type ()

var (
	DummyData atomic.Value
)

func AfterLoadDrawProbability(m *DrawProbabilityMap) {
	appendAfterLoad(func() {
		mapDrawProbability.Store(*m)
	})
}
