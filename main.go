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
	for _, i := range items {
		element := hash()
		element.Write([]byte(i))
		hll.Add(element)
	}

	cEnd := hll.Count()
	fmt.Printf("HLL cardinality approximation after adding the packets: %d.\n", cEnd)
}

//CreateItems outputs n random strings of length 16 (IP address len)
func CreateItems(n int) []string {
	var items []string
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	sLen := 16
	for i := 0; i < n; i++ {
		b := make([]rune, sLen)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		items = append(items, string(b))
	}
	//fmt.Printf("Set of strings: %v\n", items)
	return items
	//return []string{"hip", "hip", "hip", "hip", "hop", "hiphop", "test", "stuff", "blabl", "kjhkjh", "kjh"}
}

func Attack(hll *hyperloglog.HyperLogLog, nBuckets int, h hash.Hash32) []string {
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
		//		fmt.Printf("Craft: result %d, %b\n", result, result)
		leadingOne = result & mask
		//		fmt.Printf("Craft: leadingOne %d\n", leadingOne)
		if leadingOne != 0 {
			items = append(items, i)
		} else {
			discarded++
		}
		h.Reset()
	}
	fmt.Printf("Attacker crafted %d items, discarding %d.\n", len(items), discarded)
	return items
}
