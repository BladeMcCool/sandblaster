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
	header blockHdr
	tx     []transaction
}

type blockHdr struct {
	version       uint8
	prevBlockHash []byte
	txSetHash     []byte
	timestamp     int64
	target        *big.Int
	nonce         uint64
}

//idaes and notes for bitcanvas or bitmap or whatever this is going to be.
// canvas or layered on pixels weathering over time (self-removing concept)
// what if the data gets really big ..
//  periodic crystalization of canvas state? could lock it into a block

//some output types
//pixelbase : create new raw pixel resource, is the block reward. creates an output that can be spent by owner of the listed pubkey. no input required. no x or y coord associated or rgb info
//xfer      : transfer pixel resource to another pubkey. x,y and r,g,b would be ignored.
//canvas    : add a tile to the global canvas at the x, y position specified. pixel cost of canvas tbd. rgb value supplied determines base color of the canvas tileno rgb needed. must be contiguous with existing canvas. multiple creation of the same canvas tile by different pubkeys in the same block should create the canvas tile with the avg color of all the inputs that created it. the cost per new canvas tile could be a block reward amount. also if multiple tx in one block are creating the same canvas tile could create 'refund' outputs to split the excess fee and return it to the pubkeys that sent it
//            .. what if miners set the fee for this independently? multiple claims on the same canvas time come in at the same time, pick a winner based on fee? any tx that was trying to grab it but didnt win it in the block would just be recognized as a loser tx and removed from mempool/refunded etc.
//fill      : add r,g,b to a x,y coord on the canvas. result color of the tile will be whichever color (or avg of colors) has had the most layers applied to that tile. multiple 'layers' can be applied at once by using a >1 value for pixels field in the output tx.
//

const (
	OutputTypePixelbase = iota
	OutputTypeXfer
	OutputTypeCanvas
	OutputTypeFill
)

type transaction struct {
	inputs     []output
	signatures []string //each output that is being used as an input needs to have a matching valid signature here.
	outputs    []output
}

func (b *block) getTxSetHash() []byte {
	blockDataHasher := sha1.New()
	//do this need to rebuild whole thing each time we call it? could we just add to the hasher data each time new tx comes in instead?
	for _, tx := range b.tx {
		for _, in := range tx.inputs {
			blockDataHasher.Write(in.hash)
		}
		for _, out := range tx.outputs {
			blockDataHasher.Write(out.hash)
		}
	}
	return blockDataHasher.Sum(nil)
}
func (b *block) updateTxSetHash() {
	b.header.txSetHash = b.getTxSetHash()
}

type output struct {
	outType int
	pubkey  string //pubkey of who can spend it now
	pixels  int64
	color   color
	coord   coord
	hash    []byte
}

type color struct {
	r int
	g int
	b int
}
type coord struct {
	x int64
	y int64
}
type pixel struct {
	coord
	color
}

//every pixel in the canvas' will have to have its current canvas state and color value computed by processing the immutable history of the block chain.
var canvas = [][]*pixel{}

type blockReader interface {
	read()
}

func main() {
	// makekey()
	pubkey := slurpkey()
	_ = pubkey

	go miner("pubkey standin for PixelBase tx")
	// go playerControl()
	var tbo = make(chan struct{})
	<-tbo
}

var easiestTarget = new(big.Int)
var currentTarget = new(big.Int)

func miner(pubkey string) {

	//start with all ff target []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	targetIntBytes := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	currentTarget.SetBytes(targetIntBytes)
	easiestTarget.SetBytes(targetIntBytes)
	currentTargetFloat := new(big.Float)
	currentTargetFloat.SetInt(currentTarget)

	fmt.Printf("start target bytes %#v\n", currentTarget.Bytes())

	currentDiff := new(big.Int)
	currentDiff.Div(easiestTarget, currentTarget)

	fmt.Printf("start difficulty %.s\n", currentDiff)

	// panic("stop")
	//init last diffadjust time to now.
	lastDiffAdjust := time.Now().UnixNano()
	adjustEvery := 250
	desiredNsPerBlock := 1e8 //1e9=1sec
	debugStop := false
	_, _ = lastDiffAdjust, adjustEvery
	foundBlocks := 0
	blocks := []*block{}
	lastHash := []byte{}
	setPixelBasePubkey(pubkey)
	//loop forever adding to blockchain.
	for {
		//	get blockdata (randomstring)
		// blockData := RandStringBytesRmndr(64)
		// validBlock, lastHash := getBlock(blockData, lastHash)
		validBlock, lastHash := getBlock(lastHash, currentTarget)
		blocks = append(blocks, validBlock)
		foundBlocks++
		// fmt.Printf(".")
		fmt.Printf("validblock (%x) nonce (%d)\n", lastHash, validBlock.header.nonce)
		// fmt.Printf("targetIntBytes size %d\n", len(targetIntBytes))
		// fmt.Printf("currentTarget bytes size %d\n", len(currentTarget.Bytes()))
		// fmt.Printf("block hash bytes size %d\n", len(lastHash))

		//	is time for diff adjustment? (blockcount % adjustevery == 0)
		if foundBlocks%adjustEvery == 0 {
			fmt.Printf("\n\n")

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
			currentTargetFloat.Quo(currentTargetFloat, blockRateErrBigFloat)
			currentTargetFloat.Int(currentTarget)

			// currentDiff.Sub(easiestTarget, currentTarget)
			currentDiff.Div(easiestTarget, currentTarget)
			currentTargetBytes := currentTarget.Bytes()
			currentTargetZeropadd := make([]byte, len(targetIntBytes)-len(currentTargetBytes))
			currentTargetZeropadd = append(currentTargetZeropadd, currentTargetBytes...)

			fmt.Printf("lastHash was: %#v\n", lastHash)
			fmt.Printf("newdifficulty: %.s\n", currentDiff)
			fmt.Printf("newinttarget: %.s\n", currentTarget)
			fmt.Printf("new target bytes %#v\n", currentTarget.Bytes())
			fmt.Printf("new target: %x\n", currentTargetZeropadd)
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

func setPixelBasePubkey(pubkey string) {

}

// func gotNewTx() {
// 	//got a new tx so add it to the tx list and rebuild the block data
// 	currentBlock.getTxSetHash()
// }
// func gotNewBlock() {
// 	//change last block hash to the hash of this new block
// 	//any tx that were in that new block that we just got must be made sure to not exist in the block we're working on (because they would make it invalid)
// 	//for now, just wipe all the tx and start over with pixelbase tx only.
// 	currentBlock.prevHash = lastHash
// 	rebuildBlockData()
// }

// func staleBlockTimestamp() {
// 	//if we havent called rebuildBlockData for a while then freshen the timestamp used in the block data
// 	//blockdata.timestamp = now
// 	currentBlock.time = time.Now().UnixNano()
// 	rebuildBlockData()
// }

var blockData string = ""
var blockTxSetHash []byte = []byte{}
var lastHash []byte

// var currentBlock =

func rebuildBlockHdr() {
	// blockHeaderString =
	// //update the block data string we are combining with the nonce during mining to be based on the latest info.
	// blockData = ""
	// blockDataHasher := sha1.New()
	// // for range()
	// blockDataHasher.Write()

	// currentBlock.data = []byte(blockData)
	// blockTxSetHash =
	// currentBlock.getTxSetHash()
}

func newEmptyBlock(lastHash []byte, target *big.Int) *block {
	var block = &block{
		blockHdr{
			1,
			lastHash,
			[]byte{},
			time.Now().UnixNano(),
			target,
			0,
		},

		[]transaction{
		//TODO: put a pixelbase tx here.
		},
	}
	block.updateTxSetHash()
	return block
}

// func getBlock(blockData string, lastHash []byte) (*block, []byte) {
func getBlock(lastHash []byte, target *big.Int) (*block, []byte) {
	// nonce := uint64(0)
	block := newEmptyBlock(lastHash, target)
	// header := &block.header
	nonce := &(block.header.nonce)
	for {
		//	loop until hash of blockdata + nonce is < the target.

		timeBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(timeBytes, uint64(time.Now().UnixNano()))

		nonceBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(nonceBytes, *nonce)
		// fmt.Printf("nonce bytes. (%#v) \n", nonceBytes)

		blockHasher := sha1.New()
		blockHasher.Write([]byte{byte(block.header.version)})
		blockHasher.Write(block.header.prevBlockHash)
		blockHasher.Write(block.header.txSetHash)
		blockHasher.Write(timeBytes)
		blockHasher.Write(currentTarget.Bytes())
		blockHasher.Write(nonceBytes)

		blockHash := blockHasher.Sum(nil)
		hashComparer := new(big.Int)
		hashComparer.SetBytes(blockHash)
		cmp := hashComparer.Cmp(currentTarget)

		if cmp == -1 {
			// currentBlock.nonce = nonce
			// currentBlock.prevHash = lastHash

			// newBlock := &block{
			// 	lastHash,
			// 	nonce,
			// 	[]byte(blockData),
			// 	time.Now().UnixNano(),
			// }
			// lastHash = blockHash
			// fmt.Printf("in1 (%d) \n", block.header.nonce)
			// fmt.Printf("in1.header (%#v) \n", block.header)
			// fmt.Printf("in2 (%d) \n", *nonce)
			// panic("whoops")
			return block, blockHash
			//break
			// fmt.Printf("Found %d blocks so far.\n", foundBlocks)
			// fmt.Printf("Target was: %s\n", currentTarget)
			// fmt.Printf("Hash   was: %s\n", hashComparer)
			// fmt.Printf("Blockdata  was: %s\n", blockData)
			// fmt.Printf("Nonce  was: %d\n", nonce)
			// fmt.Printf("block data: %#v", newBlock)
			// fmt.Printf("blocks len: %d\n", len(blocks))
			// break
		}

		*nonce++
		// fmt.Printf("var 1 %d\n", nonce)
		// fmt.Printf("var 2 %d\n", &block.header.nonce)
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

// ---------------
