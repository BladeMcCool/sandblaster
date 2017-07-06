package main

import (
	"encoding/pem"
	"io/ioutil"
	// "crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	// "crypto/sha256"
	"fmt"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func makekey() {
	fmt.Println("whee")

	myownPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		fmt.Println(err.Error)
		os.Exit(1)
	}

	myownPublicKey := &myownPrivateKey.PublicKey
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(myownPrivateKey),
		},
	)

	f, err := os.Create("./key")
	check(err)
	defer f.Close()
	n2, err := f.Write(pemdata)
	f.Sync()
	_ = n2

	fmt.Println("Private Key : ", myownPrivateKey)
	fmt.Println("Public key ", myownPublicKey)

}

func slurpkey() *rsa.PublicKey {
	f, err := os.Open("./key")
	check(err)
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	check(err)

	block, _ := pem.Decode(data)
	if block == nil {
		panic("bad nil block")
		return nil
	}
	myownPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	myownPublicKey := &myownPrivateKey.PublicKey

	// fmt.Println("!! Private Key : ", myownPrivateKey)
	fmt.Printf("!! Public key %x %x\n", myownPublicKey.N, myownPublicKey.E)

	check(err)
	return myownPublicKey
}

// 	alistairPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)

// What do I want to be able to do?
//  create world. using random number seed. owner of world can whitelist players.
//  create player. player gets a keypair
//  add a player to the world at x,y
//  remove a player (and all their invitees?)
//  move a player that exists in the world by one unit in any direction
//  players can not pass through one another.
//  player can be in a state of healing, damaging or absorbing
//  player collision causes random damange/heal/absorb
//  switching player state

type worldstate struct {
	owner   *player
	seed    []byte
	sig     []byte
	players []*player
	guilds  []*guild
}
type player struct {
	name     string
	x, y     int
	pubkey   *rsa.PublicKey
	privkey  *rsa.PrivateKey
	invitees []*player
	guild    *guild
	state    int //(0=absorbing,1=damaging,2=healing)
	health   int //max 100
	sig      []byte
}
type guild struct {
	owner   *rsa.PublicKey
	name    string
	members []*player
}

func createworld(ownerPlayer *player, seed string) *worldstate {
	return &worldstate{
		owner:   ownerPlayer,
		seed:    []byte(seed),
		players: []*player{ownerPlayer},
	}
}

func drawWorld() {}

func scrapey() {
	fmt.Printf("Yeah here.\n")
}
