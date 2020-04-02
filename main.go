package main

import (
	"flag"
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

	scenario := flag.String("scenario", "S2", "a string")
	attackerOnly := flag.Bool("attackerOnly", true, "a bool")
	flag.Parse()

	var userItems []string
	var attackerItems []string

	//we initialize with 1000 items from a random user
	if !*attackerOnly {
		userItems = CreateItems(1000)
	}

	//Add honest user's items
	for _, i := range userItems {
		element := hash()
		element.Write([]byte(i))
		hll.Add(element)
	}

	cBeg := hll.Count()
	fmt.Printf("HLL cardinality approximation at start: %d.\n", cBeg)

	//Craft packets
	switch *scenario {
	case "S3":
		attackerItems = AttackS3(hll, nBuckets, hash())
	default:

		attackerItems = AttackS2(nBuckets, hash())
	}

	//Add the attacker's items
	for _, i := range attackerItems {
		element := hash()
		element.Write([]byte(i))
		hll.Add(element)
	}

	//Result
	cEnd := hll.Count()
	fmt.Printf("HLL cardinality approximation after adding the packets: %d.\n", cEnd)
}

//CreateItems outputs n random strings of length 5 (52^5 > 380'000'000 possibilities)
func CreateItems(n int) []string {
	fmt.Printf("Generating %d random strings...\n", n)
	var items []string
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	sLen := 5
	for len(items) != n {
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

//Attack under scenario S2
func AttackS2(nBuckets int, h hash.Hash32) []string {
	allItems := CreateItems(100000)
	fmt.Printf("Attacker is selecting the items from the random set...\n")

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

//Attack under S3 scenario
func AttackS3(hll *hyperloglog.HyperLogLog, nBuckets int, h hash.Hash32) []string {
	return nil
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
