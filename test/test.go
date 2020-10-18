package main

import "fmt"

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
	var PrimeRootTable = [][2]uint{
		{3, 2},
		{5, 2},
		{17, 3},
		{97, 5},
		{193, 5},
		{257, 3}, // 2^8 + 1
		{7681, 17},
		{12289, 11},
		{40961, 3},
		{65537, 3}, // 2^16 + 1
		{786433, 10},
		{5767169, 3},
		{7340033, 3},
		{16777259, 2}, // 2^24 + 43
		{23068673, 3},
		{104857601, 3},
		{167772161, 3},
		{268435459, 2}, // 2^28 + 3
		{469762049, 3},
		{1004535809, 3},
		{2013265921, 31},
		{2281701377, 3},
		{3221225473, 5},
		{4294967311, 3}, // 2^32 + 15

	}

	for _, pr := range PrimeRootTable {
		checkPrime(pr[0], pr[1])
	}

}
