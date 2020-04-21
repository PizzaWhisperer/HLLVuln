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
	userData := flag.Int("userData", 0, "an int")
	iterations := flag.Int("iterations", 1, "an int")
	RT20 := flag.Bool("RT20", true, "a bool")
	log := flag.Bool("log", false, "a bool")
	flag.Parse()

	var originalEst uint64
	var finalEst uint64
	var itemsPicked int
	all := AllItems()

	for i := 0; i < *iterations; i++ {

		//we initialize with 1000 items from a random user if flag is set
		var userItems []string
		userItems = CreateItems(*userData)
		for _, i := range userItems {
			element := hash()
			element.Write([]byte(i))
			hll.Add(element)
		}

		//Initial estimated cardinality
		originalEstTemp := hll.Count()
		if *log {
			fmt.Printf("HLL cardinality approximation at start: %d.\n", originalEstTemp)
		}

		regBeg := make([]uint8, 256)
		copy(regBeg, hll.Reg)

		//Craft and add attacker's packets
		var attackerItems []string
		switch *scenario {
		case "S3":
			attackerItems = AttackS3(hll, nBuckets, hash(), *RT20, all)
		default:
			attackerItems = AttackS2(nBuckets, hash(), *RT20, all)
		}
		itemsPickedTemp := len(attackerItems)
		for _, i := range attackerItems {
			element := hash()
			element.Write([]byte(i))
			hll.Add(element)
		}
		if *log {
			fmt.Printf("Attacker added %d items, so %d %s of the set.\n", len(attackerItems), len(attackerItems)/2500, "%")
		}

		//Final estimated cardinality
		finalEstTemp := hll.Count()
		if *log {
			fmt.Printf("HLL cardinality approximation at the end: %d.\n", finalEstTemp)
		}

		//Checks which buckets changed
		if *log {
			for i, reg := range hll.Reg {
				if regBeg[i] != reg {
					fmt.Printf("Reg %d, was %v and now is %v\n", i, regBeg[i], reg)
				}
			}
		}

		hll.Clear()

		//fmt.Printf("Iteration %d, original card: %d, attacker items: %d, and final card: %d.\n", i, originalEstTemp, itemsPickedTemp, finalEstTemp)
		originalEst += originalEstTemp
		itemsPicked += itemsPickedTemp
		finalEst += finalEstTemp
	}
	originalEst = originalEst / uint64(*iterations)
	itemsPicked = itemsPicked / *iterations
	finalEst = finalEst / uint64(*iterations)

	if *iterations != 1 {
		fmt.Printf("Over %d iterations, original card: %d, attacker items: %d, and final card: %d.\n", *iterations, originalEst, itemsPicked, finalEst)

	}
}

//Attack under scenario S2
func AttackS2(nBuckets int, h hash.Hash32, rt20 bool, all []string) []string {
	var allItems []string
	if rt20 {
		allItems = CreateItems(250000)
	} else {
		allItems = all
	}

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
func AttackS3(hll *hyperloglog.HyperLogLog, nBuckets int, h hash.Hash32, rt20 bool, all []string) []string {
	var allItems []string
	var mask uint32
	var exists1bit uint32

	if rt20 {
		allItems = CreateItems(250000)
	} else {
		allItems = all
	}

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
			if hll.Count() == 0 && bucket == 1 {
				//we are attacking an empty HLL, we only fill one of the empty buckets (1st by default)
				mask = 1 << (32 - 8 - 1)
			} else {
				//we found an empty bucket amongst filled ones, we do not insert anything
				mask = 0
			}
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

//CreateItems outputs n random strings of length 4 (52^4 > 7'000'000 possibilities)
func CreateItems(n int) []string {
	itemsMap := make(map[string]bool)
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	sLen := 4
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

//AllItems outputs all ASCII strings of length 4 (52^4 > 7'000'000 possibilities)
func AllItems() []string {
	var items []string
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	sLen := 4
	for i1 := 0; i1<52; i1++ {
		for i2 := 0; i2<52; i2++ {
			for i3 := 0; i3<52; i3++ {
				for i4 := 0; i4<52; i4++ {
						b := make([]rune, sLen)
							b[0] = letters[i1]
							b[1] = letters[i2]
							b[2] = letters[i3]
							b[3] = letters[i4]
							items = append(items, string(b))
				}
			}
		}
	}
	return items
}

//GenMask generates a mask to check whether a string has less then ci leading 0s
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
