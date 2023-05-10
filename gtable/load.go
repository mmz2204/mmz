package gtable

import (
	"encoding/csv"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
)

type ITableStructInit interface {
	OnInit()
}

type ITableStructKey interface {
	GetKey() interface{}
}

var (
	path    = "./cfg_table/"
	extName = ".xlsx"
)

// InitConfig 加载配置
func InitConfig() {
	LoadAll()
	WatchConfig()
}

// logFatalNotHot 执更新标记时，调用error否则调用fatal
func logFatalNotHot(template string, args ...interface{}) {
	if hot {
		log.Fatalf(template, args...)
	} else {
		log.Fatalf(template, args...)
	}
}

// parseFile 解析文件格式
func parseFile(rows [][]string, filePath string, mapObj interface{}) {
	typ := reflect.TypeOf(mapObj)
	if typ.Kind() != reflect.Ptr {
		logFatalNotHot("load result must be *map!")
	}
	typeMap := typ.Elem()
	if typeMap.Kind() != reflect.Map {
		logFatalNotHot("load result must be *map!")
	}
	typeMapElem := typeMap.Elem()
	if typeMapElem.Kind() != reflect.Ptr {
		logFatalNotHot("load result map elem must be ptr!")
	}

	typeModel := typeMapElem.Elem()
	vMap := reflect.Indirect(reflect.ValueOf(mapObj))

	//0行 中文名称
	//1行 num/string类型
	//2行 客户端类型 INT STRING
	//3行 字段名称
	headType := make(map[int]string)
	filedIdx := make(map[int][]int)

	idMax := 0
	for rowIdx, row := range rows {
		//第一行,名称不需要处理
		if rowIdx == 0 || rowIdx == 1 {
			continue
		}

		//第二行,字段名
		if rowIdx == 3 {
			pModelVal := reflect.New(typeModel)
			for colIdx, colCell := range row {
				//去两边空格
				colVal := strings.TrimSpace(colCell)
				//检查第一列是否是字段id
				/*
					if colIdx == 0 && strings.ToLower(colVal) != "id" {
						logFatalNotHot("column one not found Field id")
						return
					}
				*/

				//在反射结构体查找字段，过滤不需要读取的字段内容
				rval, idxs := caseInsenstiveFieldByName(reflect.Indirect(pModelVal), colVal)
				if !rval.CanSet() {
					continue
				}

				//记录结构体字段索引位置
				filedIdx[colIdx] = idxs
			}
			continue
		}

		//第三行,类型登记
		if rowIdx == 2 {
			for colIdx, colCell := range row {
				//去两边空格
				colVal := strings.TrimSpace(colCell)
				headType[colIdx] = colVal
			}
			continue
		}

		//数据行
		var id int = 0
		var tKey string
		pModelVal := reflect.New(typeModel)
		for colIdx, colCell := range row {
			//去两边空格
			colVal := strings.TrimSpace(colCell)
			//过滤空行
			if colIdx == 0 && colVal == "" {
				id = -1
				break
			}

			//过滤不需读取的列数据
			idxs, ok := filedIdx[colIdx]
			if !ok {
				continue
			}

			for _, index := range idxs {
				rval := reflect.Indirect(pModelVal).Field(index)
				if headType[colIdx] == "INT" {
					//空值默认为0
					if colVal == "" {
						colVal = "0"
					}

					if rval.Type().Name() == "float64" { //小数
						val, err := strconv.ParseFloat(colVal, 64)
						if err != nil {
							logFatalNotHot("config parse to float err:%s,行%d 列%d,file:%s", err, rowIdx, colIdx, filePath)
							return
						}

						rval.Set(reflect.ValueOf(val))
					} else {
						val, err := strconv.ParseInt(colVal, 10, 64)
						if err != nil {
							logFatalNotHot("config parse to int err:%s,行%d 列%d,file:%s", err, rowIdx, colIdx, filePath)
							return
						}
						valInt := int(val)
						rval.Set(reflect.ValueOf(valInt))

						//设置默认key值,因前面已经检查过，这里直接使用
						if colIdx == 0 {
							id = valInt
						}
					}
				} else {
					if colVal == "-1" || colVal == "0" || colVal == "" {
						continue
					}

					if headType[colIdx] == "STRING" {

						//设置默认key值,因前面已经检查过，这里直接使用
						if colIdx == 0 {
							tKey = colVal
						}

						rval.Set(reflect.ValueOf(colVal))
						// } else if headType[colIdx] == "split1" { //[]int,分隔符号:...
						// 	intSlice, _ := ParseStr2Slice(colVal, ":")
						// 	rval.Set(reflect.ValueOf(intSlice))
						// } else if headType[colIdx] == "item" { //[]*Item,分隔符号:|...
						// 	item, _ := ParseStr2Item(colVal)
						// 	rval.Set(reflect.ValueOf(item))
						// } else if headType[colIdx] == "attr" { //*attr,分隔符号:|...
						// 	attr := ParseStr2Attr(colVal)
						// 	rval.Set(reflect.ValueOf(attr))
						// } else if headType[colIdx] == "split2" { //[][]int分隔符号:|...
						// 	twoArray, _ := ParseStr2TwoArray(colVal)
						// 	rval.Set(reflect.ValueOf(twoArray))
					} else {
						continue
					}
				}
			}
		}

		//空行过滤掉
		if id < 0 {
			continue
		}

		if id == 0 {
			idMax++
			id = idMax
		} else {
			idMax = id
		}

		curVal := pModelVal.Interface()
		if iOnLoad, ok := curVal.(ITableStructInit); ok {
			iOnLoad.OnInit()
		}

		if iKey, ok := curVal.(ITableStructKey); ok {
			key := iKey.GetKey()
			vMap.SetMapIndex(reflect.ValueOf(key), pModelVal)
		} else if len(tKey) > 0 {
			vMap.SetMapIndex(reflect.ValueOf(tKey), pModelVal)
		} else {
			vMap.SetMapIndex(reflect.ValueOf(id), pModelVal)
		}
	}
}

// 转化成纯小写
func caseInsenstiveFieldByName(v reflect.Value, name string) (reflect.Value, []int) {
	name = strings.ToLower(name)
	if f, ok := v.Type().FieldByNameFunc(func(n string) bool { return strings.ToLower(n) == name }); ok {
		return v.FieldByIndex(f.Index), f.Index
	}

	return reflect.Value{}, nil
}

// loadExcel 加载excel文件
func loadExcel(fileName string, mapObj interface{}) {
	filePath := path + fileName + ".xlsx"

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		logFatalNotHot("Config load err:%s %s", err, filePath)
		return
	}

	sheetName := f.GetSheetName(1)
	//fmt.Println(sheetName)
	rows := f.GetRows(sheetName)

	parseFile(rows, filePath, mapObj)
}

// loadFileCsv 加载csv文件
func loadFileCsv(fileName string, mapObj interface{}) {
	//filePath := path + strings.ToLower(fileName) + ".csv"
	filePath := path + fileName + ".csv"
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		logFatalNotHot("loadFileCsv error: %#v ", err)
	}

	r := csv.NewReader(strings.NewReader(string(b)))

	/*建议1.16以后用此方法
	fp, err := os.Open(filePath)
	if err != nil {
		logFatalNotHot("loadFileCsv error: %#v ", err)
		return
	}

	r := csv.NewReader(fp)*/

	//使用','分隔符号
	r.Comma = ','

	data, err2 := r.ReadAll()
	if err2 != nil {
		log.Fatalf(fileName + " " + err2.Error())
	}

	parseFile(data, filePath, mapObj)
}

// load 文件加载
func load(fileName string, csvName string, mapObj interface{}) {
	// if tableIsCsv {
	// 	loadFileCsv(csvName, mapObj)
	// } else {
	// 	loadExcel(fileName, mapObj)
	// }
	loadFileCsv(csvName, mapObj)
}
