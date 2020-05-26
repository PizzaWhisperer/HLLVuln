package main

import (
	"fmt"
	"hash"
	"math/rand"
	"time"
)

//Attack under scenario S1
//Note: h should not be given but it is nessecary to have the Add oracle for CreateBatch and CheckItem.

//Attack under scenario S1
//Note: h should not be given but it is nessecary to have the Add oracle for CreateBatch and CheckItem.
func AttackS12(nBuckets int, h hash.Hash32, rt20 bool, all []string) []string {
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

//Attack under S1 but with less resets
func AttackS1LessResets(nBuckets int, h hash.Hash32, rt20 bool, all []string, resetsAllowed int) []string {
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

	resets := 0
	for resets < resetsAllowed {
		fmt.Printf("Step %d\n", resets)
		resets++
		//shuffle list
		//rand.Seed(time.Now().UnixNano())
		//rand.Shuffle(len(itemsCandList), func(i, j int) { itemsCandList[i], itemsCandList[j] = itemsCandList[j], itemsCandList[i] })
		//hll with 3M items / card
		//hll := FillSketch(nBuckets, itemsCandList, h)
		hll := EmptyHLL()
		//for each item, we check if it increases the count, if it does, we rm it
		//for iteration := 0; iteration < 5; iteration++ {
		//try with sane hll several times ?
		for _, item := range itemsCandList {
			if CheckBadItemWithoutReset(item, nBuckets, h, hll) {
				delete(itemsCand, item)
				itemsDisc = append(itemsDisc, item)
			}
		}
		fmt.Printf("disc size %d\n", len(itemsDisc))
		itemsCandList = Map2List(itemsCand)
		//}
		fmt.Printf("itemsCand size %d\n", len(itemsCand))
	}
	return itemsCandList
}
