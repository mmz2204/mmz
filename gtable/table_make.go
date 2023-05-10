//go:build tablegen
// +build tablegen

package gtable

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unsafe"

	"github.com/modern-go/reflect2"
)

const (
	defaultKeyName = "id"
)

type TableStructRuntime struct {
	keyField  []*reflect.StructField
	keyType   reflect.Kind
	hasGetKey bool
	varName   string
}

func (p *TableStructRuntime) GenKeyParams() (string, string, string, string, bool) {
	if len(p.keyField) == 1 {
		return "key " + p.keyType.String(), "", "", "", p.keyType == reflect.Int
	}
	allInt := true
	params := make([]string, 0, len(p.keyField))
	placeHold := make([]string, 0, len(p.keyField))
	callParam := make([]string, 0, len(p.keyField))
	getKeyParam := make([]string, 0, len(p.keyField))
	for _, v := range p.keyField {
		params = append(params, GenKeyParam(v))
		callParam = append(callParam, firstCharLower(v.Name))
		getKeyParam = append(getKeyParam, "p."+v.Name)
		switch v.Type.Kind() {
		case reflect.Int:
			placeHold = append(placeHold, "%d")
		case reflect.String:
			placeHold = append(placeHold, "%s")
			allInt = false
		default:
			log.Fatalf("do not support key type, edit func (p *TableStructRuntime) GenKeyParams() code plz!")
		}

	}
	return strings.Join(params, ","), strings.Join(placeHold, "_"), strings.Join(callParam, ","), strings.Join(getKeyParam, ","), allInt
}

func firstCharLower(p string) string {
	tmp := strings.ToLower(p[:1]) + p[1:]
	if tmp == "type" {
		return "typ"
	}
	return tmp
}

func GenKeyParam(p *reflect.StructField) string {
	return firstCharLower(p.Name) + " " + p.Type.String()
}

type TableStruct struct {
	typeName string
	csv      string
	excel    string
	depend   []string
	TableStructRuntime
}

var (
	tables   map[string]*TableStruct
	tagReg   *regexp.Regexp
	feildReg *regexp.Regexp
	fatal    bool
)

func init() {
	fatal = false
	tables = make(map[string]*TableStruct)
	tagReg = regexp.MustCompile(`(?sU)/\*(.+)\*/`)
	feildReg = regexp.MustCompile(`(?m)@(.+)[^\n]`)
}

func GetKeyType(kind reflect.Kind, allIntKey bool, keySize int) string {
	if allIntKey {
		switch keySize {
		case 1:
			return kind.String()
		case 2:
			return "Key2"
		case 3:
			return "Key3"
		default:
			return "string"
		}
	}
	return kind.String()
}

func makeTableStructStuff(ts *TableStruct, output map[string]string) {
	fillKey(ts)

	p, h, c, getKey, allIntKey := ts.GenKeyParams()
	key := ""
	if h != "" {
		if allIntKey {
			key = fmt.Sprintf(keyMake2, len(ts.keyField), c)
		} else {

			key = fmt.Sprintf(keyMake, h, c)
		}
	}

	output["mapType"] += fmt.Sprintf(mapType, ts.typeName, GetKeyType(ts.keyType, allIntKey, len(ts.keyField)), ts.typeName)
	output["mapVar"] += fmt.Sprintf(mapVar, ts.varName)
	output["loadFunc"] += fmt.Sprintf(loadFunc, ts.typeName, ts.typeName, ts.csv, ts.typeName, ts.excel, ts.csv, ts.typeName)
	output["getFunc"] += fmt.Sprintf(getFunc, ts.typeName, p, ts.typeName, key, ts.varName, ts.typeName)
	output["getAllFunc"] += fmt.Sprintf(getAllFunc, ts.typeName, ts.typeName, ts.varName, ts.typeName)
	output["callLoad"] += fmt.Sprintf(callLoad, ts.typeName)
	if ts.hasGetKey {
		if len(ts.keyField) > 1 {
			if allIntKey {
				output["getKey"] += fmt.Sprintf(getKeyNPattern, ts.typeName, len(ts.keyField), getKey)
			} else {
				output["getKey"] += fmt.Sprintf(getKeyStringPattern, ts.typeName, h, getKey)
			}
		} else {
			output["getKey"] += fmt.Sprintf(getKeyPattern, ts.typeName, ts.keyField[0].Name)
		}
	}

}

func makeTableVar(ts *TableStruct, output *[]string, output2 *[]string) {
	*output = append(*output, fmt.Sprintf("dummy%s *%s", ts.typeName, ts.typeName))
	*output2 = append(*output2, fmt.Sprintf("dummy%s = &%s{}", ts.typeName, ts.typeName))
}

func StringBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func WriteTableGo(output map[string]string, filePath string) {
	loadAll := fmt.Sprintf(loadAllFunc, output["callLoad"])
	context := fmt.Sprintf(fileContext, output["mapType"], output["mapVar"], loadAll, output["loadFunc"], output["getKey"], output["getFunc"], output["getAllFunc"])

	ioutil.WriteFile(filePath, StringBytes(context), 0644)
}

func WriteVarGo(output []string, output2 []string, filePath string) {
	context := fmt.Sprintf(varFile, strings.Join(output, "\n\t"), strings.Join(output2, "\n\t"))

	ioutil.WriteFile(filePath, StringBytes(context), 0)
}

func fillKey(ts *TableStruct) {
	realName := "gtable." + ts.typeName
	typ := reflect2.TypeByName(realName)
	if typ == nil {
		log.Fatalf("struct type not exits:%s!", realName)
	}

	realType := typ.Type1()
	hasId := false
	ts.keyField = make([]*reflect.StructField, 0)
	for i := 0; i < realType.NumField(); i++ {
		field := realType.Field(i)
		//统一转成小写，这样不区分大小写比较
		if strings.ToLower(field.Name) == defaultKeyName {
			hasId = true
		}
		if tag, ok := field.Tag.Lookup("gtable"); ok {
			if strings.Index(tag, "key") != -1 {
				ts.keyField = append(ts.keyField, &field)
			}
		}
	}

	if len(ts.keyField) == 0 {
		if !hasId {
			log.Printf("struct %s has not specific key or default key [%s]!", realName, defaultKeyName)
			fatal = true
			return
		}

		//field, _ := realType.FieldByName(defaultKeyName)
		//统一转成小写，这样不区分大小写比较
		field, _ := realType.FieldByNameFunc(func(n string) bool { return strings.ToLower(n) == defaultKeyName })
		ts.keyField = append(ts.keyField, &field)
	} else {
		ts.hasGetKey = true
	}

	if len(ts.keyField) > 1 {
		ts.keyType = reflect.String
	} else {
		ts.keyType = ts.keyField[0].Type.Kind()
	}

	ts.varName = fmt.Sprintf(varName, ts.typeName)
}

func enumFile(dirPth string, prefix string) (files []string, err error) {
	files = make([]string, 0, 10)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	pthSep := string(os.PathSeparator)
	prefix = strings.ToUpper(prefix)
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(fi.Name()), prefix) {
			files = append(files, dirPth+pthSep+fi.Name())
		}
	}
	return files, nil
}

func readAll(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return ioutil.ReadAll(f)
}

func walkFile(path string) {
	//fmt.Println(path)
	fc, err := readAll(path)
	if err != nil {
		return
	}

	lst := tagReg.FindAllString(string(fc), -1)
	for _, v := range lst {
		parseTags(v)
	}
}

func parseTag(tagStr string) []string {
	lst := strings.Split(tagStr, " ")
	if len(lst) != 2 {
		fmt.Printf("incorrect tag string:%s\n", tagStr)
	}
	return lst
}

func parseTags(tagStr string) {
	lst := feildReg.FindAllStringSubmatch(string(tagStr), -1)
	if len(lst) == 0 {
		return
	}
	name := strings.ToLower(lst[0][1])
	t := &TableStruct{typeName: lst[0][1]}
	lst = lst[1:]
	for _, v := range lst {
		tags := parseTag(v[1])
		switch tags[0] {
		case "csv":
			t.csv = tags[1] //strings.ToLower(tags[1])
		case "excel":
			t.excel = tags[1]
		case "depend":
			t.depend = append(t.depend, strings.Split(tags[1], "|")...)
		}
	}
	tables[name] = t
}

func makeTableGo() {
	output := make(map[string]string)
	lst := make([]*TableStruct, 0, len(tables))
	for _, v := range tables {
		lst = append(lst, v)
	}

	log.Printf("----------%d", len(lst))

	sort.Slice(lst, func(i, j int) bool {
		return lst[i].typeName < lst[j].typeName
	})

	for _, v := range lst {
		makeTableStructStuff(v, output)
	}
	if fatal {
		panic(nil)
	}

	makeCustom(lst, output)
	WriteTableGo(output, "./table.go")
}

func makeVar() {
	output := make([]string, 0, len(tables))
	output2 := make([]string, 0, len(tables))
	for _, v := range tables {
		makeTableVar(v, &output, &output2)
	}

	WriteVarGo(output, output2, "./var.go")
}

func makeCustom2() {
	output := make([]string, 0, len(tables))
	for _, v := range tables {
		output = append(output, makeCustomName(v)...)
	}
	WriteVarGo([]string{}, output, "./v_custom.go")
}
