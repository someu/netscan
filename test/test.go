package main

import (
	"fmt"
)

func checkPrime(prime uint, root uint) {
	var inverse uint = 1
	var i uint = 1
	//fmt.Printf("%d^%d = %d\n", root, 0, 1)
	for ; i < prime; i++ {
		inverse = (inverse * root) % prime
		//fmt.Printf("%d^%d = %d\n", root, i, inverse)
	}
	if inverse == 1 {
		fmt.Printf("check %d %d succes\n", prime, root)
	} else {
		fmt.Printf("check %d %d failed\n", prime, root)
	}
}





func main() {

}
