//go:build tablegen
// +build tablegen

package gtable

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strings"

	"github.com/modern-go/reflect2"
)

// 0
type IKeySlice interface {
	keySlice(int) bool
}

type IKeySliceName interface {
	keySliceName() []string
}

type IKeySliceSort interface {
	keySliceSort(interface{}, int)
}

// 1
type IValueSlice interface {
	valueSlice(int) bool
}

type IValueSliceName interface {
	valueSliceName() []string
}

type IValueSliceSort interface {
	valueSliceSort(interface{}, int)
}

// 2
type IFilterMap interface {
	filterMap(int) bool
}

type IFilterMapName interface {
	filterMapName() []string
}

type IAfterLoad interface {
	afterLoad(interface{})
}

type TabelCustomStrings struct {
	typePattern string
	varPattern  string
	varName     string
	typeName    string
	implPattern string
}

const customMax = 2

var (
	sliceIFunc     = []reflect.Type{reflect.TypeOf((*IKeySlice)(nil)).Elem(), reflect.TypeOf((*IValueSlice)(nil)).Elem(), reflect.TypeOf((*IFilterMap)(nil)).Elem()}
	sliceIFuncName = []reflect.Type{reflect.TypeOf((*IKeySliceName)(nil)).Elem(), reflect.TypeOf((*IValueSliceName)(nil)).Elem(), reflect.TypeOf((*IFilterMapName)(nil)).Elem()}
	sliceIFuncSort = []reflect.Type{reflect.TypeOf((*IKeySliceSort)(nil)).Elem(), reflect.TypeOf((*IValueSliceSort)(nil)).Elem(), reflect.TypeOf((*IFilterMapName)(nil)).Elem()}
	sliceFuncName  = []string{"keySliceName", "valueSliceName", ""}

	tabelCustomPattern = []TabelCustomStrings{
		{
			typePattern: "%sKeySlice\t%s\n\t",
			varPattern:  "sliceKey%s\tatomic.Value\n\t",
			varName:     "sliceKey%s",
			typeName:    "%sKeySlice",
			implPattern: "keySlice",
		},
		{
			typePattern: "%sValueSlice\t%s\n\t",
			varPattern:  "sliceValue%s\tatomic.Value\n\t",
			varName:     "sliceValue%s",
			typeName:    "%sValueSlice",
			implPattern: "valueSlice",
		},
		// {
		// 	typePattern: "%sFilterMap\t%s\n\t",
		// 	varPattern:  "mapFilter%s\tatomic.Value\n\t",
		// 	varName:     "mapFilter%s",
		// 	implPattern: "filterMap",
		// },
	}
)

func getCustomNames(v interface{}, i int) []string {
	switch i {
	case 0:
		if f, ok := v.(IKeySliceName); ok {
			return f.keySliceName()
		}
	case 1:
		if f, ok := v.(IValueSliceName); ok {
			return f.valueSliceName()
		}
		// case 2:
		// 	if f, ok := v.(IFilterMapName); ok {
		// 		return f.filterMapName()
		// 	}
	}
	return make([]string, 0)
}

func getCustomElemType(ts *TableStruct, i int) string {
	switch i {
	case 0:
		return "[]" + ts.keyType.String()
	case 1:
		return "[]*" + ts.typeName
		// case 2:
		// 	return fmt.Sprintf("map[%s]*%s", ts.keyType.String(), ts.typeName)
	}
	return ts.typeName
}

func MakeImpl(ts *TableStruct, varTmp, varName, name string, i, j int, hasSort bool, _make, _op, _append, _sort *[]string) {
	*_make = append(*_make, fmt.Sprintf(afterMake, varTmp, fmt.Sprintf(tabelCustomPattern[i].typeName, name))) // getCustomElemType(ts, i)))
	switch i {
	case 0:
		*_op = append(*_op, fmt.Sprintf(afterOp, tabelCustomPattern[i].implPattern, j, varTmp, varTmp, "k"))
	case 1:
		*_op = append(*_op, fmt.Sprintf(afterOp, tabelCustomPattern[i].implPattern, j, varTmp, varTmp, "v"))
		// case 2:
		// 	*_op = append(*_op, fmt.Sprintf(afterOpMap, tabelCustomPattern[i].implPattern, j, varTmp))
	}

	*_append = append(*_append, fmt.Sprintf(afterAppend, varName, varTmp))
	if hasSort {
		*_sort = append(*_sort, fmt.Sprintf(afterSort, ts.typeName, tabelCustomPattern[i].implPattern, varTmp, j))
	}
}

func makeCustomGet(i int, name, varName, typeName string, getMap map[string]string) {
	if i < 2 {
		getMap["getAllFunc"] += fmt.Sprintf(getCustomFunc, name, typeName, varName, typeName)
	}
}

func makeCustomOne(ts *TableStruct, output map[string]string, getMap map[string]string) {
	realName := "gtable." + ts.typeName
	typ := reflect2.TypeByName(realName)
	if typ == nil {
		log.Fatalf("struct type not exits:%s!", realName)
	}
	_make := make([]string, 0)
	_op := make([]string, 0)
	_append := make([]string, 0)
	_sort := make([]string, 0)
	realType := typ.Type1()
	varIndex := 1
	hasK := false
	for i := 0; i < customMax; i++ {
		if !reflect.PtrTo(realType).Implements(sliceIFunc[i]) {
			continue
		}

		if i == 0 {
			hasK = true
		}
		val := reflect.New(realType).Interface()
		names := getCustomNames(val, i)
		isDefaultName := true
		if len(names) == 0 {
			isDefaultName = false
			names = append(names, ts.typeName)
		}
		bHasSort := reflect.PtrTo(realType).Implements(sliceIFuncSort[i])

		for j, v := range names {
			output["typePattern"] += fmt.Sprintf(tabelCustomPattern[i].typePattern, v, getCustomElemType(ts, i))
			output["varPattern"] += fmt.Sprintf(tabelCustomPattern[i].varPattern, v)
			varTmp := fmt.Sprintf("var%d", varIndex)
			varName := fmt.Sprintf(tabelCustomPattern[i].varName, v)
			varIndex++
			MakeImpl(ts, varTmp, varName, v, i, j, bHasSort, &_make, &_op, &_append, &_sort)
			methodName := v
			if !isDefaultName {
				methodName = fmt.Sprintf(tabelCustomPattern[i].typeName, v)
			}
			makeCustomGet(i, methodName, varName, fmt.Sprintf(tabelCustomPattern[i].typeName, v), getMap)
		}
	}
	callStructAfterLoad := ""
	if reflect.PtrTo(realType).Implements(reflect.TypeOf((*IAfterLoad)(nil)).Elem()) {
		callStructAfterLoad = fmt.Sprintf(structAfterLoad, ts.typeName)
	}
	if len(_make) > 0 {
		strK := "_"
		if hasK {
			strK = "k"
		}
		output["implPattern"] += fmt.Sprintf(afterFunc, ts.typeName, ts.typeName, callStructAfterLoad, strings.Join(_make, "\n"), strK, strings.Join(_op, "\n"), strings.Join(_sort, "\n\t"), ts.varName, strings.Join(_append, "\n"))
	} else {
		output["implPattern"] += fmt.Sprintf(afterFunc2, ts.typeName, ts.typeName, callStructAfterLoad, ts.varName)
	}
}

func makeCustom(lst []*TableStruct, getMap map[string]string) {
	output := make(map[string]string)
	for _, v := range lst {
		makeCustomOne(v, output, getMap)
	}

	WriteCustomGo(output, `.\table_after_load.go`)
}

func WriteCustomGo(output map[string]string, filePath string) {
	context := fmt.Sprintf(customFile, output["typePattern"], output["varPattern"], output["implPattern"])

	ioutil.WriteFile(filePath, StringBytes(context), 0644)
}

func makeCustomName(ts *TableStruct) []string {
	realName := "gtable." + ts.typeName
	typ := reflect2.TypeByName(realName)
	if typ == nil {
		log.Fatalf("struct type not exits:%s!", realName)
	}

	realType := typ.Type1()
	r := make([]string, 0)
	for i := 0; i < customMax; i++ {
		if !reflect.PtrTo(realType).Implements(sliceIFuncName[i]) {
			continue
		}
		r = append(r, fmt.Sprintf("_ = dummy%s.%s()", ts.typeName, sliceFuncName[i]))
	}

	return r
}
