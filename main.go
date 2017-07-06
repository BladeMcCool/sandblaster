package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

type block struct {
	prevHash []byte
	nonce    uint64
	data     []byte
	time     int64
}

type blockReader interface {
	read()
}

func main() {
	go miner()
	// makekey()
	pubkey := slurpkey()
	_ = pubkey
	// go playerControl()
	var tbo = make(chan struct{})
	<-tbo
}

func miner() {

	//start with all ff target []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	theTarget := new(big.Int)
	targetIntBytes := []byte{0xFF, 0x11, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x99, 0xFF}
	theTarget.SetBytes(targetIntBytes)
	theFloatTarget := new(big.Float)
	theFloatTarget.SetInt(theTarget)

	fmt.Printf("start target bytes %#v\n", theTarget.Bytes())

	//init last diffadjust time to now.
	lastDiffAdjust := time.Now().UnixNano()
	adjustEvery := 250
	desiredNsPerBlock := 1e8 //1e9=1sec
	debugStop := false
	_, _ = lastDiffAdjust, adjustEvery
	foundBlocks := 0
	blocks := []*block{}
	lastHash := []byte{}
	//loop forever adding to blockchain.
	for {
		//	get blockdata (randomstring)
		blockData := RandStringBytesRmndr(64)
		nonce := uint64(0)
		for {
			//	loop until hash of blockdata + nonce is < the target.
			blockHasher := sha1.New()

			nonceBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(nonceBytes, nonce)

			blockHasher.Write([]byte(blockData))
			blockHasher.Write([]byte(nonceBytes))
			blockHash := blockHasher.Sum(nil)

			hashComparer := new(big.Int)
			hashComparer.SetBytes(blockHash)
			cmp := hashComparer.Cmp(theTarget)
			if cmp == -1 {
				foundBlocks++
				newBlock := &block{
					lastHash,
					nonce,
					[]byte(blockData),
					time.Now().UnixNano(),
				}
				blocks = append(blocks, newBlock)
				lastHash = blockHash
				//break
				// fmt.Printf("Found %d blocks so far.\n", foundBlocks)
				// fmt.Printf("Target was: %s\n", theTarget)
				// fmt.Printf("Hash   was: %s\n", hashComparer)
				// fmt.Printf("Blockdata  was: %s\n", blockData)
				// fmt.Printf("Nonce  was: %d\n", nonce)
				// fmt.Printf("block data: %#v", newBlock)
				fmt.Printf("blocks len: %d\n", len(blocks))
				break
			}

			nonce++
		}

		//	is time for diff adjustment? (blockcount % adjustevery == 0)
		if foundBlocks%adjustEvery == 0 {
			fmt.Printf("foundBlocks == %d, time to adjust difficulty.\n", foundBlocks)
			now := time.Now().UnixNano()
			nsSinceLastDiffAdjust := now - lastDiffAdjust
			nsPerBlock := float64(nsSinceLastDiffAdjust) / float64(adjustEvery)
			blockRateErr := desiredNsPerBlock / nsPerBlock
			// blockrateErr is "number of times too easy it currently seems to be. -- so if it is 2x too easy, we have to double the difficulty by cutting the target in half. (if it is 0.25x too easy, we have to cut the difficulty to 1/4 of current by multiplying the target by 4)"
			fmt.Printf("desired ns per block is %.0f\n", desiredNsPerBlock)
			fmt.Printf("actual  ns per block is %.0f\n", nsPerBlock)
			fmt.Printf("diff is %.2fx too easy\n", blockRateErr)

			blockRateErrBigFloat := big.NewFloat(blockRateErr)
			theFloatTarget.Quo(theFloatTarget, blockRateErrBigFloat)
			theFloatTarget.Int(theTarget)
			fmt.Printf("newinttarget: %.s\n", theTarget)
			fmt.Printf("new target bytes %#v\n", theTarget.Bytes())
			fmt.Printf("lastHash was: %#v\n", lastHash)
			lastDiffAdjust = now
			// break
		}

		if debugStop || (foundBlocks == 5000) {
			// fmt.Println("debug stop.")
			break

		}

		//		how long did it take to make these blocks and what portion of the block rate is that?
		//		divide target by blockrate error to find new target.
		//		set new target in place.
	}

}

func funstuff() {

	randostrings := []string{
		"biacon",
		"testomation",
		"Sae is a cute kitty",
		"Jasmine is a lovely lady",
		"Harmony is a well behaved child",
		"Lets NOT get a chinchilla.",
	}
	_ = randostrings

	// x := 1
	// for {
	// 	x = x * 2
	// 	fmt.Printf("Cash factory has earned this much $CAD: %d\n", x)
	// 	if x == 0 {
	// 		break
	// 	}
	// }

	// for i := 0; i < len(randostrings); i++ {
	// 	mySha := sha1.New()
	// 	mySha.Write([]byte(randostrings[i]))
	// 	digest := mySha.Sum(nil)
	// 	fmt.Printf("%d a: %x %#v\n", i+1, digest, digest)

	// 	verybig := &big.Int{}
	// 	_ = verybig
	// 	verybig.SetBytes(digest)
	// 	fmt.Printf("%d x: %s\n", i+1, verybig)

	// 	fmt.Printf("%d b: %x\n", i+1, sha1.Sum([]byte(randostrings[i])))

	// }
	funzos := [][]byte{
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		[]byte{0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF},
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF},
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x00},
		[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		[]byte{0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE},
		[]byte{0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE},
		[]byte{0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE},
		[]byte{0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00},
	}
	_ = funzos
	// truthiness := map[int]string{
	// 	-1 : ""
	// }

	// func (*Int) Cmp
	// func (x *Int) Cmp(y *Int) (r int)
	// Cmp compares x and y and returns:
	// -1 if x <  y
	//  0 if x == y
	// +1 if x >  y
	tooFastness := 3.5
	tooFastnessBigFloat := big.NewFloat(tooFastness)
	_ = tooFastness
	theTarget := new(big.Int)
	targetIntBytes := []byte{0xFF, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	theTarget.SetBytes(targetIntBytes)
	theFloatTarget := new(big.Float)
	theFloatTarget.SetInt(theTarget)
	fmt.Printf(" toofastness: %.2f\n", tooFastnessBigFloat)
	fmt.Printf("      target: %s\n", theTarget)
	fmt.Printf("float target: %.2f\n", theFloatTarget)
	theFloatTarget.Quo(theFloatTarget, tooFastnessBigFloat)
	fmt.Printf("new   target: %.2f\n", theFloatTarget)
	newIntTarget, _ := theFloatTarget.Int(nil)
	fmt.Printf("newinttarget: %.s\n", newIntTarget)
	newTargetBytes := newIntTarget.Bytes()
	fmt.Printf("new int target bytes: %#v\n", newTargetBytes)
	seeAgain := new(big.Int)
	seeAgain.SetBytes(newTargetBytes)
	fmt.Printf("see again: %s\n", seeAgain)
	fmt.Printf("see again: %s\n", RandStringBytesRmndr(50))
	// newTarget := new(big.Int)
	// newTarget.Div(theTarget, tooFastness)
	// fmt.Printf("target: %s, newtarget: %s", theTarget, newTarget)

	// return

	lastfunner := &big.Int{}
	for i := 0; i < len(funzos); i++ {
		verybig := &big.Int{}
		verybig.SetBytes(funzos[i])
		fmt.Printf("%d x: %s\n", i+1, verybig)

		if i > 0 {
			fmt.Printf("\tthis one cmp last one? : %d\n", verybig.Cmp(lastfunner))
		}
		lastfunner = verybig
	}
}
