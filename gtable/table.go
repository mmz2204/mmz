package gtable

/*
this file is auto generate by table_make.
do NOT edit plz!
*/

import (
	"sync/atomic"
)

type (
	DrawProbabilityMap map[int]*DrawProbability
)

var (
	hot                bool
	mapDrawProbability atomic.Value
)

func LoadAll() {
	hot = false
	LoadDrawProbability()

	hot = true
	callAfterLoad(afterLoadCall)
	afterLoadCall = afterLoadCall[:0]
}

func LoadDrawProbability() {
	tmp := make(DrawProbabilityMap)
	if !hot {
		initHotLoad("drawprobability", LoadDrawProbability)
	}
	load("", "drawprobability", &tmp)
	AfterLoadDrawProbability(&tmp)
}

func GetDrawProbability(key int) *DrawProbability {
	if v, ok := mapDrawProbability.Load().(DrawProbabilityMap); ok {
		return v[key]
	}
	return nil
}

func GetDrawProbabilityMap() *DrawProbabilityMap {
	if v, ok := mapDrawProbability.Load().(DrawProbabilityMap); ok {
		return &v
	}
	return nil
}
