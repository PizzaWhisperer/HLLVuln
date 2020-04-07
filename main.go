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

	scenario := flag.String("scenario", "S2", "a string")
	attackerOnly := flag.Bool("attackerOnly", true, "a bool")
	benchmark := flag.Bool("benchmark", false, "a bool")
	flag.Parse()

	if *benchmark {

		var originalEst uint64
		var finalEst uint64
		var itemsPicked int

		for i := 0; i < 30; i++ {

			var userItems []string
			var attackerItems []string

			if !*attackerOnly {
				userItems = CreateItems(1000)
			}

			for _, i := range userItems {
				element := hash()
				element.Write([]byte(i))
				hll.Add(element)
			}

			originalEstTemp := hll.Count()

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

			finalEstTemp := hll.Count()

			hll.Clear()

			fmt.Printf("Iteration %d, original card: %d, attacker items: %d, and final card: %d.\n", i, originalEstTemp, itemsPickedTemp, finalEstTemp)
			originalEst += originalEstTemp
			itemsPicked += itemsPickedTemp
			finalEst += finalEstTemp
		}
		originalEst = originalEst / 30
		itemsPicked = itemsPicked / 30
		finalEst = finalEst / 30

		fmt.Printf("Over 30 iterations, original card: %d, attacker items: %d, and final card: %d.\n", originalEst, itemsPicked, finalEst)
	} else {

		var userItems []string
		var attackerItems []string

		//we initialize with 1000 items from a random user
		if !*attackerOnly {
			fmt.Printf("Generating the honest users items and adding them...\n")
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
		fmt.Printf("Generating the attackers random items, picking and adding them...\n")
		switch *scenario {
		case "S3":
			attackerItems = AttackS3(hll, nBuckets, hash())
		default:
			attackerItems = AttackS2(nBuckets, hash())
		}
		fmt.Printf("Attacker found %d items meeting the requirements, discarding %d.\n", len(attackerItems), 100000-len(attackerItems))
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
	return items
}

//Attack under S3 scenario
func AttackS3(hll *hyperloglog.HyperLogLog, nBuckets int, h hash.Hash32) []string {
	allItems := CreateItems(100000)

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
