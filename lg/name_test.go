package lg

import (
	"fmt"
	"testing"
)

type TestStruct struct {
	Age int
}

func FunHandle() {
	fmt.Println("this is test func handle")
}

func TestFuncName(t *testing.T) {
	t.Run("TestFuncNAme", func(t *testing.T) {
		t.Logf("func name: %v", FuncName(FunHandle))
	})
}

func TestStructName(t *testing.T) {
	t.Run("TestStructName", func(t *testing.T) {
		t.Logf("struct name: %v", StructName(&TestStruct{}))
	})
}
