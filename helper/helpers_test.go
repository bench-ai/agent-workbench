package helper

import (
	"testing"
)

func TestContains(t *testing.T) {
	contains := Contains[int]

	testSlice := []int{
		1, 25, 71, 89, 34, 108,
	}

	if contains(testSlice, 99) {
		t.Error("recognized number not in testSlice")
	}

	if !contains(testSlice, 108) {
		t.Error("failed to recognize number in testSlice")
	}
}

func TestDeleteByIndex(t *testing.T) {

	stringSlice := []string{
		"data", "test", "water", "benchai",
	}

	del := DeleteByIndex[string]

	err, stringSlice := del(stringSlice, 0)

	if err != nil {
		t.Error("failed to delete by index")
	}

	if len(stringSlice) != 3 {
		t.Error("slice size did not decrease by 1")
	}

	err, _ = del(stringSlice, -1)

	if err == nil {
		t.Error("failed to detect negative index")
	}

	err, _ = del(stringSlice, 10)

	if err == nil {
		t.Error("failed to detect out of bound index")
	}

	_, stringSlice = del(stringSlice, 2)

	if len(stringSlice) != 2 {
		t.Error("slice size did not decrease by 1 @ index 2")
	}

	_, stringSlice = del(stringSlice, 1)

	if len(stringSlice) != 1 {
		t.Error("slice size did not decrease by 1 @ index 1")
	}

	_, stringSlice = del(stringSlice, 0)

	if len(stringSlice) != 0 {
		t.Error("slice size did not decrease by 1 @ index 1")
	}

	err, _ = del(stringSlice, 0)

	if err == nil {
		t.Error("failed to detect empty slice")
	}
}

func TestIsLte(t *testing.T) {
	lte := IsLte[int]

	if !lte(-10, 100, false) {
		t.Error("did not detect that -10 is < 100")
	}

	if !lte(-10, -10, true) {
		t.Error("did not detect that -10 is <= -10")
	}

	if lte(-10, -100, false) {
		t.Error("did not detect that -10 is > -100")
	}
}

func TestIsGte(t *testing.T) {
	gte := IsGte[int]

	if !gte(-10, 100, false) {
		t.Error("did not detect that 100 is > 100")
	}

	if !gte(10, 10, true) {
		t.Error("did not detect that 10 is <= 10")
	}

	if gte(-10, -100, false) {
		t.Error("did not detect that -10 is > -100")
	}
}

func TestIsBetween(t *testing.T) {
	bte := IsBetween[int]

	if !bte(0, 100, 20, false, false) {
		t.Error("failed to detect value 20 that is between 0 and 100")
	}

	if !bte(0, 0, 0, true, true) {
		t.Error("failed to detect value that 0 is between / equal to 0 and 0")
	}

	if bte(10, 100, -10, true, true) {
		t.Error("failed to detect value that -10 is not between 10 and 100")
	}
}
