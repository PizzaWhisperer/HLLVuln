package main

import (
	"flag"
	"fmt"
	"hash"
	"math/rand"
	"time"

	"./clarkduvall/hyperloglog"

	"./spaolacci/murmur3"
)

func main() {

	//Setup
	p := uint8(8)
	hll, _ := hyperloglog.New(p)
	hash := murmur3.New32
	nBuckets := 1 << p

	//Parse the command line arguments
	scenario := flag.String("scenario", "S2", "a string")
	userData := flag.Bool("userData", true, "a bool")
	benchmark := flag.Int("benchmark", 1, "an int")
	flag.Parse()

	var originalEst uint64
	var finalEst uint64
	var itemsPicked int

	for i := 0; i < *benchmark; i++ {

		//we initialize with 1000 items from a random user if flag is set
		var userItems []string
		if *userData {
			userItems = CreateItems(1000)
		}
		for _, i := range userItems {
			element := hash()
			element.Write([]byte(i))
			hll.Add(element)
		}

		//Initial estimated cardinality
		originalEstTemp := hll.Count()
		if *benchmark == 1 {
			fmt.Printf("HLL cardinality approximation at start: %d.\n", originalEstTemp)
		}

		//Craft and add attacker's packets
		var attackerItems []string
		switch *scenario {
		case "S3":
			attackerItems = AttackS3(hll, nBuckets, hash())
		default:
			attackerItems = AttackS2(nBuckets, hash())
		}
		itemsPickedTemp := len(attackerItems)
		for _, i := range attackerItems {
			element := hash()
			element.Write([]byte(i))
			hll.Add(element)
		}
		if *benchmark == 1 {
			fmt.Printf("Attacker added %d items, so %d %s of the set.\n", len(attackerItems), len(attackerItems)/2500, "%")
		}

		//Final estimated cardinality
		finalEstTemp := hll.Count()
		if *benchmark == 1 {
			fmt.Printf("HLL cardinality approximation at the end: %d.\n", finalEstTemp)
		}

		hll.Clear()

		//fmt.Printf("Iteration %d, original card: %d, attacker items: %d, and final card: %d.\n", i, originalEstTemp, itemsPickedTemp, finalEstTemp)
		originalEst += originalEstTemp
		itemsPicked += itemsPickedTemp
		finalEst += finalEstTemp
	}
	originalEst = originalEst / uint64(*benchmark)
	itemsPicked = itemsPicked / *benchmark
	finalEst = finalEst / uint64(*benchmark)

	if *benchmark != 1 {
		fmt.Printf("Over %d iterations, original card: %d, attacker items: %d, and final card: %d.\n", *benchmark, originalEst, itemsPicked, finalEst)

	}
}

//CreateItems outputs n random strings of length 5 (52^5 > 380'000'000 possibilities)
func CreateItems(n int) []string {
	itemsMap := make(map[string]bool)
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	sLen := 5
	for len(itemsMap) != n {
		b := make([]rune, sLen)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		if !itemsMap[string(b)] {
			itemsMap[string(b)] = true
		}
	}
	var items []string
	for i := range itemsMap {
		items = append(items, i)
	}
	return items
}

//Attack under scenario S2
func AttackS2(nBuckets int, h hash.Hash32) []string {
	allItems := CreateItems(250000)

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
	return items
}

//Attack under S3 scenario
func AttackS3(hll *hyperloglog.HyperLogLog, nBuckets int, h hash.Hash32) []string {
	allItems := CreateItems(250000)

	var mask uint32
	var exists1bit uint32

	var items []string
	discarded := 0

	for _, i := range allItems {
		_, err := h.Write([]byte(i))
		if err != nil {
			fmt.Printf("Craft: err %v\n", err)
			continue
		}
		result := h.Sum32()

		bucket := (result & uint32(((1<<8)-1)<<24)) >> 24
		ci := hll.Reg[bucket]
		if ci == 0 {
			//we can adopt two strategies, or we skip this bucket, i.e., we discard
			//elements falling into that bucket, or we keeps the one with leading 0s
			//like in S2. For now we use 2nd strategy.
			mask = 1 << (32 - 8 - 1)
		} else {
			mask = GenMask(ci)
		}
		exists1bit = result & mask

		if exists1bit != 0 {
			items = append(items, i)
		} else {
			discarded++
		}
		h.Reset()
	}
	return items
}

func GenMask(ci uint8) uint32 {
	mask := uint32(0)
	for i := uint8(0); i < ci; i++ {
		mask = mask << 1
		mask += 1
	}
	for i := ci; i < 24; i++ {
		mask = mask << 1
	}
	return mask
}
