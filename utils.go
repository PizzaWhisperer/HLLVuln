package main

import (
	"hash"
	"math"
	"math/rand"
	"time"

	"./clarkduvall/hyperloglog"
)

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
	for i1 := 0; i1 < 52; i1++ {
		for i2 := 0; i2 < 52; i2++ {
			for i3 := 0; i3 < 52; i3++ {
				for i4 := 0; i4 < 52; i4++ {
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

//Create batch picks elements such that the estimated cardinality is > 2.5*nBuckets
func CreateBatch(nBuckets int, items []string, h hash.Hash32) ([]string, uint64) {
	hll := EmptyHLL()

	var batch []string
	for _, i := range items {
		batch = append(batch, i)
		h.Write([]byte(i))
		hll.Add(h)
		h.Reset()
		if float64(hll.Count()) > float64(nBuckets)*2.5 {
			return batch, hll.Count()
		}
	}
	return nil, 0
}

//CheckItem looks if the ith item satisfy the attack requirements
func CheckItem(index int, batch []string, count uint64, nBuckets int, h hash.Hash32) bool {
	hll := EmptyHLL()

	rest := make([]string, len(batch))

	copy(rest, batch)
	rest[index] = rest[len(rest)-1]
	rest = rest[:len(rest)-1]
	for _, i := range rest {
		h.Write([]byte(i))
		hll.Add(h)
		h.Reset()
	}
	// println("prev")
	prev := hll.Count()
	// println("now")
	// hll.Count()
	return prev == count
}

//EmptyHLL returns an emtpy HLL, it is used to simulate the reset function
func EmptyHLL() *hyperloglog.HyperLogLog {
	hll, _ := hyperloglog.New(p)
	return hll
}

func RMItems(items []string, rm []string) []string {
	rmMap := make(map[string]struct{}, len(rm))
	for _, x := range rm {
		rmMap[x] = struct{}{}
	}
	var result []string
	for _, x := range items {
		if _, found := rmMap[x]; !found {
			result = append(result, x)
		}
	}
	return result
}

func Map2List(x map[string]bool) []string {
	var list []string
	for s, _ := range x {
		list = append(list, s)
	}
	return list
}

func FillRegisters(nBuckets int, items []string, h hash.Hash32) *hyperloglog.HyperLogLog {
	hll := EmptyHLL()

	//1st Step: find big set which fills all registers
	bigSetLoop := true
	var batch []string
	for bigSetLoop {
		hll = EmptyHLL()
		batch = nil
		for _, i := range items {
			batch = append(batch, i)
			h.Write([]byte(i))
			hll.Add(h)
			h.Reset()
			//Based on coupon collector, we can stop after nBuckets*log(nBuckets)+nBuckets items
			stopLen := int(float64(nBuckets)*math.Log(float64(nBuckets)) + float64(nBuckets))
			if len(batch) >= stopLen {
				bigSetLoop = false
				if countZeros(hll.Reg) != 0 {
					//fmt.Printf("WE FAILED AT 1st step COUNT %d and %d non 0 reg\n", hll.Count(), countZeros(hll.Reg))
				} else {
					//fmt.Println("YAS")
					break
				}
			}
		}
	}
	//2nd Step: make that set as small as possible (technically, need M items)
	smallSetLoop := true
	var newBatch []string
	newBatch = batch
	prevCount := hll.Count()
	//fmt.Printf("At the begining : count %d, items %d\n", hll.Count(), len(newBatch))

	for smallSetLoop {
		var tempBatch []string
		tempBatch = newBatch
		for _, item := range newBatch {
			hll = EmptyHLL()
			//try without i
			rest := RMItems(newBatch, []string{item})
			for _, r := range rest {
				h.Write([]byte(r))
				hll.Add(h)
				h.Reset()
			}
			if prevCount == hll.Count() {
				tempBatch = RMItems(tempBatch, []string{item})
				break
			}
		}
		newBatch = tempBatch
		prevCount = hll.Count()
		//fmt.Printf("At the middle : count %d, items %d, 0 regs %d\n", hll.Count(), len(newBatch), countZeros(hll.Reg))
		stopL := nBuckets
		if len(newBatch) <= stopL {
			if countZeros(hll.Reg) != 0 {
				//fmt.Printf("WE FAILED AT 2nd step COUNT %d and %d non 0 reg\n", hll.Count(), countZeros(hll.Reg))
			} else {
				//fmt.Println("YAS 2nd step")
				smallSetLoop = false
				break
			}
		}
	}
	//Final HLL
	hll = EmptyHLL()

	for _, r := range newBatch {
		h.Write([]byte(r))
		hll.Add(h)
		h.Reset()
	}
	//fmt.Printf("HLL with count %d, items %d, 0-reg %d\n", hll.Count(), len(newBatch), countZeros(hll.Reg))
	return hll
}

//return true if item is bad
func CheckBadItem(item string, nBuckets int, h hash.Hash32, hllOrig *hyperloglog.HyperLogLog) bool {
	//work on new struct
	hll, _ := hyperloglog.New(p)
	copy(hll.Reg, hllOrig.Reg)

	prev := hll.Count()
	h.Write([]byte(item))
	hll.Add(h)
	h.Reset()
	new := hll.Count()
	return prev < new
}

func countZeros(s []uint8) uint32 {
	var c uint32
	for _, v := range s {
		if v == 0 {
			c++
		}
	}
	return c
}
