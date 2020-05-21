package main

//Attack under scenario S1
//Note: h should not be given but it is nessecary to have the Add oracle for CreateBatch and CheckItem.
import (
	"fmt"
	"hash"
	"math/rand"
	"time"

	"./clarkduvall/hyperloglog"
)

func AttackS1Ori(nBuckets int, h hash.Hash32, rt20 bool, all []string) []string {
	var allItems []string
	if rt20 {
		allItems = CreateItems(250000)
	} else {
		allItems = all
	}
	var itemsCand []string
	var itemsDisc []string

	for iteration := 0; iteration < 2; iteration++ {
		itemsCand = nil
		for !(allItems == nil || len(itemsCand) > 100000) {
			batch, count := CreateBatch(nBuckets, allItems, h)
			if batch == nil {
				return itemsCand
			}
			for i, item := range batch {
				if CheckItem(i, batch, count, nBuckets, h) {
					itemsCand = append(itemsCand, item)
				} else {
					itemsDisc = append(itemsDisc, item)
				}
			}
		}
		fmt.Printf("itemsCand size %d\n", len(itemsCand))
		fmt.Printf("allItems size %d\n", len(allItems))
		fmt.Printf("disc size %d\n", len(itemsDisc))
		allItems = RMItems(allItems, itemsDisc)
		fmt.Printf("allItems size %d\n", len(allItems))
		//allItems = append(allItems, itemsCand...)
		//fmt.Printf("itemsCand size %d\n", len(itemsCand))
	}
	fmt.Printf("itemsCand size %d\n", len(itemsCand))
	fmt.Printf("allItems size %d\n", len(allItems))
	return itemsCand
}

func AttackS1Or(nBuckets int, h hash.Hash32, rt20 bool, all []string) []string {
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

	//Step 1: find at least M = 2^n items such that we fill all buckets
	//Goal: hll.Count > 2.5*est but all counters filled so no LCPA correction
	//Step 2: try to add all items and discard the ones that increment the counter
	//Goal: rm items with lots of leading zeros

	goal := len(allItems) / 2
	//We want to keep the 50% best
	step := 0
	for len(itemsDisc) <= goal {
		fmt.Printf("Step %d\n", step)
		step++
		//shuffle list
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(itemsCandList), func(i, j int) { itemsCandList[i], itemsCandList[j] = itemsCandList[j], itemsCandList[i] })
		//hll with registers filled
		hll := FillRegisters(nBuckets, itemsCandList, h)
		var hllcpy *hyperloglog.HyperLogLog
		hllcpy = hll
		//for each item, we check if it increases the count, if it does, we rm it
		for _, item := range itemsCandList {
			if CheckBadItem(item, nBuckets, h, hllcpy) {
				delete(itemsCand, item)
				itemsDisc = append(itemsDisc, item)
				break
			}
		}
		itemsCandList = Map2List(itemsCand)
		fmt.Printf("itemsCand size %d\n", len(itemsCand))
		fmt.Printf("disc size %d\n", len(itemsDisc))
	}
	return itemsCandList
}
