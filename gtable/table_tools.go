package gtable

import (
	"log"
)

type Key2 struct {
	Vala, Valb int
}

type Key3 struct {
	Vala, Valb, Valc int
}

func MakeKey2(a, b int) Key2 {
	return Key2{
		Vala: a,
		Valb: b,
	}
}

func MakeKey3(a, b, c int) Key3 {
	return Key3{
		Vala: a,
		Valb: b,
		Valc: c,
	}
}

// GoTryContinue 循环调用捕获异常协程
func GoTryContinue(f func()) {
	go func() {
		SafeCallContinue(99999, f)
	}()
}

// SafeCallContinue 循环安全调用接口
func SafeCallContinue(retry int, f func()) {
	for i := 0; i < retry; i++ {
		if SafeCall(f) {
			return
		}
		log.Printf("!!!SafeCallContinue RETRY:%d!!!", i+1)
	}
	log.Fatalf("!!!SafeCall PANIC Looped!!! retry count: %d", retry)
}

// SafeCall 安全调用接口
func SafeCall(f func()) (r bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("!!!SafeCall PANIC!!!error=", err)
			r = false
		}
	}()
	f()
	return true
}
