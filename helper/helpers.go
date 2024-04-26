package helper

import (
	"cmp"
	"errors"
)

func Contains[T comparable](slice []T, item T) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}

func DeleteByIndex[T any](s []T, index int) (error, []T) {

	if index >= len(s) {
		return errors.New("index out of bounds"), nil
	}

	if index < 0 {
		return errors.New("index must be >= 0"), nil
	}

	if index == 0 {
		return nil, s[1:]
	}

	if index == len(s)-1 {
		return nil, s[:len(s)-1]
	}

	slice1 := s[:index]
	slice2 := s[index+1:]

	return nil, append(slice1, slice2...)
}

func IsLte[T cmp.Ordered](lowVal, highVal T, checkEqual bool) bool {

	if checkEqual && lowVal == highVal {
		return true
	}

	return lowVal < highVal
}

func IsGte[T cmp.Ordered](lowVal, highVal T, checkEqual bool) bool {
	return !IsLte[T](highVal, lowVal, checkEqual)
}

func IsBetween[T cmp.Ordered](lowRange, highRange, val T, lte, gte bool) bool {
	return IsGte[T](lowRange, val, gte) && IsLte[T](val, highRange, lte)
}
