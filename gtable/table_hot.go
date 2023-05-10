package gtable

import (
	"log"
)

var (
	void struct{}
)

type StringSet map[string]struct{}

func NewStringSet() StringSet {
	return make(StringSet)
}
func (p StringSet) Insert(v string) bool {
	_, r := p[v]
	if !r {
		p[v] = void
	}
	return !r
}
func (p StringSet) Empty() bool {
	return len(p) == 0
}
func (p StringSet) Clear() {
	for k := range p {
		delete(p, k)
	}
}

// LoadAll appendHotLoad luanchHotUpdate 只能由此3个入口修改
var (
	hotMap        map[string]func()
	hotWaitList   StringSet
	afterLoadCall []func()
)

func init() {
	hotMap = make(map[string]func())
	hotWaitList = NewStringSet()
	afterLoadCall = make([]func(), 0)
}

func initHotLoad(name string, f func()) {
	hotMap[name] = f
}

// 禁止多线程访问, 只能由watch线程调用
func appendHotLoad(name string) {
	hotWaitList.Insert(name)
}

// table package初始化完毕后,禁止多线程访问, 只能由watch线程调用
func appendAfterLoad(f func()) {
	afterLoadCall = append(afterLoadCall, f)
}

// 禁止多线程访问, 只能由watch线程调用
func luanchHotUpdate() {
	if hotWaitList.Empty() {
		return
	}

	for k := range hotWaitList {
		if f, ok := hotMap[k]; ok {
			f()
			log.Printf("table file:%s is loaded", k)
		} else {
			log.Printf("do not exist data in file:%s", k)
		}
	}
	/*
		callList := afterLoadCall
		afterLoadCall = make([]func(), 0)
			tools.PushTask(func() {
				callAfterLoad(callList)
			})
	*/
	hotWaitList.Clear()
}

func callAfterLoad(lst []func()) {
	for _, f := range lst {
		f()
	}
}
