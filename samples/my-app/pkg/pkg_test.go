package pkg_test

import (
	"testing"

	"sample/pkg"
)

func TestAdd_Fail(t *testing.T) {
	actual := pkg.Add(1, 2)
	expected := -1

	if actual != expected {
		t.Errorf("err")
	}
}

func TestAdd_Succeed(t *testing.T) {
	actual := pkg.Add(1, 2)
	expected := 3

	if actual != expected {
		t.Errorf("err")
	}
}
