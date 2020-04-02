package main

import (
	"fmt"
	"hash"
	"math/rand"
	"time"

	"github.com/clarkduvall/hyperloglog"
	"github.com/spaolacci/murmur3"
)

func main() {

	//Setup
	p := uint8(8)
	hll, _ := hyperloglog.New(p)
	hash := murmur3.New32
	nBuckets := 1 << p

	cBeg := hll.Count()
	fmt.Printf("HLL cardinality approximation at start: %d.\n", cBeg)

	//Craft packets
	items := Attack(hll, nBuckets, hash())

	//Add them
	for _, i := range items {
		element := hash()
		element.Write([]byte(i))
		hll.Add(element)
	}

	//Result
	cEnd := hll.Count()
	fmt.Printf("HLL cardinality approximation after adding the packets: %d.\n", cEnd)
}

//CreateItems outputs n random strings of length 40
func CreateItems(n int) []string {
	fmt.Printf("Generating %d random strings...\n", n)
	var items []string
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	sLen := 40
	for i := 0; i < n; i++ {
		b := make([]rune, sLen)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		if !Contains(items, string(b)) {
			items = append(items, string(b))
		}
	}
	fmt.Printf("Done.\n")
	//fmt.Printf("Set of strings: %v\n", items)
	return items
}

//Attack selects packets from items that satisfy the attacker requirements
func Attack(hll *hyperloglog.HyperLogLog, nBuckets int, h hash.Hash32) []string {
	fmt.Printf("Attacker is selecting the items from the random set...\n")
	allItems := CreateItems(100000)

	var mask uint32
	mask = 1 << (32 - 8 - 1)

	var leadingOne uint32

	var items []string
	discarded := 0

	for _, i := range allItems {
		_, err := h.Write([]byte(i))
		if err != nil {
			fmt.Printf("Craft: err %v\n", err)
			continue
		}
		result := h.Sum32()
		leadingOne = result & mask
		if leadingOne != 0 {
			items = append(items, i)
		} else {
			discarded++
		}
		h.Reset()
	}
	fmt.Printf("Attacker found %d items meeting the requirements, discarding %d.\n", len(items), discarded)
	return items
}

//Contains checks if the array a contains x
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
