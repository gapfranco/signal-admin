package main

import (
	"strings"
	"unicode"
)

func validarCNPJ(cnpj string) bool {
	if cnpj == "" {
		return true
	}

	cnpj = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToUpper(r)
		}
		return -1
	}, cnpj)

	if len(cnpj) != 14 {
		return false
	}

	if cnpj[12] < '0' || cnpj[12] > '9' || cnpj[13] < '0' || cnpj[13] > '9' {
		return false
	}

	allSame := true
	for i := 1; i < 14; i++ {
		if cnpj[i] != cnpj[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	charVal := func(c byte) int {
		return int(c - '0')
	}

	calcDigit := func(s string, weights []int) int {
		sum := 0
		for i, w := range weights {
			sum += charVal(s[i]) * w
		}
		if r := sum % 11; r >= 2 {
			return 11 - r
		}
		return 0
	}

	w1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	w2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	return calcDigit(cnpj, w1) == int(cnpj[12]-'0') &&
		calcDigit(cnpj, w2) == int(cnpj[13]-'0')
}
