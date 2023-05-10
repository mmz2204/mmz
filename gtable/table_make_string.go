//go:build tablegen
// +build tablegen

package gtable

const (
	varName  = "map%s"
	mapType  = "%sMap map[%s]*%s\n\t"
	mapVar   = "%s atomic.Value\n\t"
	callLoad = "\tLoad%s()\n"
	keyMake  = `key := fmt.Sprintf("%s",%s)
	`
	keyMake2 = `key := MakeKey%d(%s)
	`
	loadFunc = `
func Load%s() {
	tmp := make(%sMap)
	if !hot {
		initHotLoad("%s", Load%s)
	}
	load("%s", "%s", &tmp)
	AfterLoad%s(&tmp)
}	
`
	loadAllFunc = `
func LoadAll() {
	hot = false
%s
	hot = true
	callAfterLoad(afterLoadCall)
	afterLoadCall = afterLoadCall[:0]
}	
`
	getFunc = `
func Get%s(%s) *%s {
	%sif v, ok := %s.Load().(%sMap); ok {
		return v[key]
	}
	return nil
}
`
	getAllFunc = `
func Get%sMap() *%sMap {
	if v, ok := %s.Load().(%sMap); ok {
		return &v
	}
	return nil
}
`

	fileContext = `package gtable
/*
this file is auto generate by table_make.
do NOT edit plz!
*/

import (
	"sync/atomic"
)

type(
	%s)

var(
	hot bool
	%s)
%s
%s
%s
%s
%s	
`
)

const (
	varFile = `//+build tablegen
	
package gtable
var(
	%s
)

func init(){
	%s
}
`
)

const (
	customFile = `package gtable
/*
this file is auto generate by table_make.
do NOT edit plz!
*/

import (
	"sync/atomic"
)

type(
	%s
)
var(
	DummyData     atomic.Value
	%s
)

%s
`
)

const (
	afterFunc = `func AfterLoad%s(m *%sMap) {%s
%s
	for %s, v := range *m {
%s}
	%s
	appendAfterLoad(func() {
		%s.Store(*m)
%s
	})
}
`
	afterFunc2 = `func AfterLoad%s(m *%sMap) {%s
	appendAfterLoad(func() {
		%s.Store(*m)
	})
}
`
	getCustomFunc = `func Get%s() %s {
	if v, ok := %s.Load().(%s); ok {
		return v
	}
	return nil
}

`
	structAfterLoad = `
	((*%s)(nil)).afterLoad(*m)`
	afterMake = "\t%s := make(%s, 0)"
	afterOp   = `		if v.%s(%d) {
			%s = append(%s, %s)
		}
	`
	afterSort  = "((*%s)(nil)).%sSort(%s,%d)"
	afterOpMap = `		if v.%s(%d) {
			%s[k]=v
		}
`
	afterAppend = "\t\t%s.Store(%s)"
)

const (
	getKeyStringPattern = `func (p *%s) GetKey() interface{} {
	return fmt.Sprintf("%s", %s)
}
`
	getKeyNPattern = `func (p *%s) GetKey() interface{} {
	return MakeKey%d(%s)
}
`
	getKeyPattern = `func (p *%s) GetKey() interface{} {
	return p.%s
}
`
)
