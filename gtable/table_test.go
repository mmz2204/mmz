//go:build tablegen
// +build tablegen
package gtable

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestStep1(t *testing.T) {
	os.Rename("./var.go", "./var2.go")
	files, _ := enumFile(".", "c_")
	for _, v := range files {
		walkFile(v)
	}

	makeVar()
	os.Remove("./var2.go")
}

func TestStep2(t *testing.T) {
	//go 1.16+ 必须执行这一步
	os.Rename("./v_custom.go", "./v_custom2.go")
	files, _ := enumFile(".", "c_")
	for _, v := range files {
		walkFile(v)
	}

	makeCustom2()
	os.Remove("./v_custom2.go")
}

func TestStep3(t *testing.T) {
	os.Rename("./table.go", "./table2.go")
	os.Rename("./table_after_load.go", "./table_after_load2.go")
	files, _ := enumFile(".", "c_")
	for _, v := range files {
		walkFile(v)
	}

	makeTableGo()
	os.Remove("./var.go")
	os.Remove("./v_custom.go")
	os.Remove("./table2.go")
	os.Remove("./table_after_load2.go")
	goFmt("table_after_load.go")
	goFmt("table.go")
}

func goFmt(file string) {
	c := exec.Command("gofmt", "-l", "-w", file)
	c.Dir = "./"
	c.Run()
	out, _ := c.Output()
	fmt.Println(string(out))
	os.Chmod("./"+file, 0444)
}
