package validator

import (
	"context"
	"reflect"
	"testing"
)

func TestVar(t *testing.T) {
	v, err := Var("gt=0,lt=10")(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, 5) {
		t.Fatalf("get %v want %v", v, 5)
	}
}

func TestVarWithValue(t *testing.T) {
	v, err := VarWithValue("other", "eqcsfield")(context.Background(), "other")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, "other") {
		t.Fatalf("get %v want %v", v, "other")
	}
}

func TestStruct(t *testing.T) {
	var me = struct {
		Name string `json:"name" validate:"required,printascii"`
	}{"233"}

	v, err := Struct()(context.Background(), me)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, me) {
		t.Fatalf("get %v want %v", v, me)
	}
}
