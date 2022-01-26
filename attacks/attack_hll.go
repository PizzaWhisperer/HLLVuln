package main

import (
	"fmt"
	"hash"
  "math/rand"
	"math"
	"time"
	"encoding/binary"

	"./clarkduvall/hyperloglog"
	"./spaolacci/murmur3"
)

var n = uint8(8)
var m = 1 << n
var maxValue = int(math.Pow(2, 20))-1
var iterations = 50


func main() {

	fmt.Printf("Results for %d iterations:\n", iterations)

  //S1 scenario
var originalEst, finalEst, nItermInserted int
	//originalEst, finalEst, nItermInserted, nResets = runAttack("S1", 0, false, 1)
	//fmt.Printf("Results in the S1 scenario: %d, %d, %d, %d\n", originalEst, finalEst, nItermInserted, nResets)

  //originalEst, finalEst, nItermInserted, nResets = runAttack("S1", 0, false, 100)
  //fmt.Printf("Results in the S1 scenario with B = 100: %d, %d, %d, %d\n", originalEst, finalEst, nItermInserted, nResets)

  //S2 scenario

  originalEst, finalEst, nItermInserted, _ = runAttack("S2", 0, false, 1)
  fmt.Printf("Results in the S2 scenario: %d, %d, %d\n", originalEst, finalEst, nItermInserted)

  originalEst, finalEst, nItermInserted, _ = runAttack("S2", 500, false, 1)
  fmt.Printf("Results in the S2 scenario with 500 initial items: %d, %d, %d\n", originalEst, finalEst, nItermInserted)

  originalEst, finalEst, nItermInserted, _ = runAttack("S2", 1000, false, 1)
  fmt.Printf("Results in the S2 scenario with 1000 initial items: %d, %d, %d\n", originalEst, finalEst, nItermInserted)

  originalEst, finalEst, nItermInserted, _ = runAttack("S2", 2000, false, 1)
  fmt.Printf("Results in the S2 scenario with 2000 initial items: %d, %d, %d\n", originalEst, finalEst, nItermInserted)

  //S4 scenario

  originalEst, finalEst, nItermInserted, _ = runAttack("S4", 0, false, 1)
  fmt.Printf("Results in the S4 scenario with B = 1: %d, %d, %d\n", originalEst, finalEst, nItermInserted)

  originalEst, finalEst, nItermInserted, _ = runAttack("S4", 1000, false, 1)
  fmt.Printf("Results in the S4 scenario with 1000 initial items: %d, %d, %d\n", originalEst, finalEst, nItermInserted)

  //Setup of RT20

  //originalEst, finalEst, nItermInserted, nResets = runAttack("S1", 0, true, 1)
  //fmt.Printf("Results in the S1 scenario, RT20 setup, with B = 1: %d, %d, %d, %d\n", originalEst, finalEst, nItermInserted, nResets)

  //originalEst, finalEst, nItermInserted, nResets = runAttack("S1", 0, true, 100)
  //fmt.Printf("Results in the S1 scenario, RT20 setup, with B = 100: %d, %d, %d, %d\n", originalEst, finalEst, nItermInserted, nResets)

  originalEst, finalEst, nItermInserted, _ = runAttack("S2", 0, true, 1)
  fmt.Printf("Results in the S2 scenario, RT20 setup: %d, %d, %d\n", originalEst, finalEst, nItermInserted)

  originalEst, finalEst, nItermInserted, _ = runAttack("S2", 1000, true, 1)
  fmt.Printf("Results in the S2 scenario, RT20 setup, with 1000 initial items: %d, %d, %d\n", originalEst, finalEst, nItermInserted)

  originalEst, finalEst, nItermInserted, _ = runAttack("S4", 0, true, 1)
  fmt.Printf("Results in the S4 scenario, RT20 setup, with B = 1: %d, %d, %d\n", originalEst, finalEst, nItermInserted)

  originalEst, finalEst, nItermInserted, _ = runAttack("S4", 1000, true, 1)
  fmt.Printf("Results in the S4 scenario, RT20 setup, with 1000 initial items: %d, %d, %d\n", originalEst, finalEst, nItermInserted)

}

//runAttack is sub-function to be called in main. Does all the setup and function calls.
func runAttack(scenario string, nInitialItems int, RT20 bool, B uint32) (int, int, int, int){


	var avgOriginalEst int
  var avgFinalEst int
  var avgNItermInserted int
  var avgNResets int

	for i:=0; i<iterations; i++ {
		var originalEst int
		var finalEst int
		var nItermInserted int
		var nResets int

		rand.Seed(time.Now().UnixNano())

		hll, _ := hyperloglog.New(n)
		hllHash := murmur3.New32

		// Insert initial items
		InsertInitialItems(hll, hllHash(), nInitialItems)

		originalEst = int(hll.Count())

		regCopy := make([]uint8, m)
		copy(regCopy, hll.Reg)

		switch scenario {
		case "S1":
			nItermInserted, nResets = AttackS1(hll, m, hllHash(), RT20, B)
		case "S2":
			nItermInserted = AttackS2(hll, m, hllHash(), RT20, B)
		case "S4":
			nItermInserted = AttackS4(hll, m, hllHash(), RT20, B)
		default:
			println("Not implemented")
		}

		finalEst = int(hll.Count())
		hll.Clear()

		avgOriginalEst+=originalEst
		avgFinalEst+=finalEst
		avgNItermInserted+=nItermInserted
		avgNResets+=nResets

	}
return avgOriginalEst/iterations, avgFinalEst/iterations, avgNItermInserted/iterations, avgNResets/iterations
}


//Inserts nInitialItems random items into the HLL sketch
func InsertInitialItems(hll *hyperloglog.HyperLogLog, hllHash hash.Hash32, nItems int) {

  //Shuffle all items and take first nItems
  items := rand.Perm(maxValue)[:nItems]

  //Insert them
  for _, i := range items {
    hllHash.Write(itob(i))
    hll.Add(hllHash)
		hllHash.Reset()
  }
}

//Converts int to byte array to write into hash
func itob(i int) []byte {
	b := make([]byte, binary.MaxVarintLen32)
	binary.PutUvarint(b, uint64(i))
	return b
}

//Generates an int of length 32-n with ci leading 1's and the rest 0's
func GenMask(ci uint8) uint32 {
	mask := uint32(0)
	for i := uint8(0); i < ci; i++ {
		mask = mask << 1
		mask += 1
	}
	for i := ci; i < 32-n; i++ {
		mask = mask << 1
	}
	return mask
}

//Attack in the S1 scenario as described in the paper
//We note that the reset (hll.new), insert (hash.Write - hll.Add - hash.Reset)
//and count (hll.Count) lines are typically perfromed by the attacker via an API.
//We do not implement such API, hence why we are "giving" and using information
//such as the hash to the Attack function, altought, as described in our paper,
//it is not used by the attacker.
func AttackS1(hll *hyperloglog.HyperLogLog, m int, hllHash hash.Hash32, RT20 bool, B uint32) (int, int) {
    nResets := 0

    items := rand.Perm(maxValue)
  if RT20 {
    items = items[:250000]
  }

  switchPoint := uint64(float64(m) * 2.5)
  targetEstimate := uint64(float64(m) * math.Log(float64(m)/float64(m-int(B))))
  var card uint64

  filteredItems := items

  //Stop condition: the cardinality at the end of the iteration does not go above targetEstimate.
  for card == 0 || card > targetEstimate {

    //Reset HLL
		hll, _ := hyperloglog.New(n)
    nResets++
    card = 0

    l := len(filteredItems)
		id := 0

		for id < l {
      //Insert item in HLL
      hllHash.Write(itob(filteredItems[id]))
      hll.Add(hllHash)
      hllHash.Reset()

      if (hll.Count() < targetEstimate) || (card == hll.Count() && hll.Count() < switchPoint) {
        id++
      } else if card != hll.Count() && hll.Count() < switchPoint {
        //Discard them
        filteredItems = append(filteredItems[:id], filteredItems[id+1:]...)
        l--
      } else { //hll.Count > switchPoint
        card = hll.Count()
        break
      }
      card = hll.Count()
		}
  }

return len(filteredItems), nResets -1 //-1 to account for first reset of the loop
}

//Attack in the S2 scenario as described in the paper
func AttackS2(hll *hyperloglog.HyperLogLog, m int, hllHash hash.Hash32, RT20 bool, B uint32) int {
    inserted := 0
    items := rand.Perm(maxValue)
    if  RT20 {
      items = items[:250000]
    }
    mask := GenMask(1)

    emptyBool := hll.Count() == 0

    for _, i := range items {

      _, err := hllHash.Write(itob(i))
      if err != nil {
        fmt.Printf("Hash error: err %v\n", err)
        continue
      }
      result := hllHash.Sum32()
      bucket := (result & uint32(((1<<8)-1)<<24)) >> 24

      if (emptyBool && bucket < B) || (!emptyBool && (result & mask != 0)) {
        //There is a leading 1, or in case of empty sketch, to targeted bucket.
        hll.Add(hllHash)
        inserted++
      }

      hllHash.Reset()
    }

    return inserted
}

//Attack in the S4 scenario as described in the paper
func AttackS4(hll *hyperloglog.HyperLogLog, m int, hllHash hash.Hash32, RT20 bool, B uint32) int {
  inserted := 0
  items := rand.Perm(maxValue)
    if  RT20 {
      items = items[:250000]
    }
    //First we check if we are attacking an empty sketch.
    //If so, we will target only B buckets
    emptyBool := hll.Count() == 0
    for _, i := range items {
      _, err := hllHash.Write(itob(i))
      if err != nil {
        fmt.Printf("Hash error: err %v\n", err)
        continue
      }
      result := hllHash.Sum32()
      bucket := (result & uint32(((1<<8)-1)<<24)) >> 24
			var mask uint32
      ci := hll.Reg[bucket]
      if ci == 0 {
        //We insert in an empty bucket only if it is among the targeted
        //in an empty sketch
        if emptyBool && bucket < B {
          mask = GenMask(1) //@Kenny should it be GenMask(24)? i.e. we fill the bucket to the max
        } else {
          mask = 0
        }
      } else {
        mask = GenMask(ci)
      }
      if result & mask != 0 {
        //There is a 1 in the ci leading bits
        hll.Add(hllHash)
        inserted++
      }
      hllHash.Reset()
    }
    return inserted

}
