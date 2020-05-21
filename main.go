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

var p = uint8(8)
var nBuckets = 1 << p

func main() {

	//Setup
	hll, _ := hyperloglog.New(p)
	hash := murmur3.New32

	//Parse the command line arguments
	scenario := flag.String("scenario", "S2", "a string")
	userData := flag.Int("userData", 0, "an int")
	iterations := flag.Int("iterations", 1, "an int")
	RT20 := flag.Bool("RT20", true, "a bool")
	log := flag.Int("log", 0, "an int")
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
		if *log != 0 {
			fmt.Printf("HLL cardinality approximation at start: %d.\n", originalEstTemp)
		}

		regBeg := make([]uint8, 256)
		copy(regBeg, hll.Reg)

		//Craft and add attacker's packets
		var attackerItems []string
		switch *scenario {
		case "S3":
			attackerItems = AttackS3(hll, nBuckets, hash(), *RT20, all)
		case "S1":
			attackerItems = AttackS1(nBuckets, hash(), *RT20, all)
		default:
			attackerItems = AttackS2(nBuckets, hash(), *RT20, all)
		}
		itemsPickedTemp := len(attackerItems)
		for _, i := range attackerItems {
			element := hash()
			element.Write([]byte(i))
			hll.Add(element)
		}
		if *log == 2 {
			fmt.Printf("Attacker added %d items, so %d %s of the set.\n", len(attackerItems), len(attackerItems)/2500, "%")
		}

		//Final estimated cardinality
		finalEstTemp := hll.Count()
		if *log != 0 {
			fmt.Printf("HLL cardinality approximation at the end: %d.\n", finalEstTemp)
		}

		//Checks which buckets changed
		if *log == 3 {
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

//Attack under scenario S1
//Note: h should not be given but it is nessecary to have the Add oracle for CreateBatch and CheckItem.
func AttackS1(nBuckets int, h hash.Hash32, rt20 bool, all []string) []string {
	var allItems []string
	if rt20 {
		allItems = CreateItems(250000)
	} else {
		allItems = all
	}

	itemsCand := make(map[string]bool, len(allItems))
	for _, s := range allItems {
		itemsCand[s] = true
	}
	itemsCandList := Map2List(itemsCand)
	var itemsDisc []string

	//goal := len(allItems) / 2
	//We want to keep the 50% best
	step := 0
	for step < 1 {
		fmt.Printf("Step %d\n", step)
		step++
		//shuffle list
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(itemsCandList), func(i, j int) { itemsCandList[i], itemsCandList[j] = itemsCandList[j], itemsCandList[i] })
		//hll with registers filled
		hll := FillRegisters(nBuckets, itemsCandList, h)
		//for each item, we check if it increases the count, if it does, we rm it
		for _, item := range itemsCandList {
			if CheckBadItem(item, nBuckets, h, hll) {
				delete(itemsCand, item)
				itemsDisc = append(itemsDisc, item)
			}
		}
		itemsCandList = Map2List(itemsCand)
		//fmt.Printf("itemsCand size %d\n", len(itemsCand))
		//fmt.Printf("disc size %d\n", len(itemsDisc))
	}
	return itemsCandList
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

	emptyBool := hll.Count() == 0
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
			if emptyBool && bucket == 1 {
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
