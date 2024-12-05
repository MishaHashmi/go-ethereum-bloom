package main

import (

	// "encoding/json"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"time"

	"github.com/spencerkimball/cbfilter"
)

// import "light"
const (
	N  = 4
	B  = 8
	FP = 0.03
)

func unix_time() int64 {
	currentTime := time.Now().Unix()
	return currentTime
}

type tx_map_val struct {
	gasPrice float64
	nonce    float64
	time     int64
}

func (val *tx_map_val) bumpGasPrice(gasPrice float64) {
	val.gasPrice = gasPrice
}

//global vars
var (
	txCbfLookup    bool
	txShadowLookup bool

	simulationTime int64
	lastExpiryTime int64
	lastDebugTime  int64

	//count
	entryTx int = 0
	exitTx  int = 0

	//added to the block
	blockTx int = 0
	//invalid exit tx
	invalidTx int = 0

	//flag for same nonce from same address
	sameNonce    bool = false
	gasPriceBump bool = false
)

//var err error
var cbf_txpool, err = cbfilter.NewFilter(N, B, FP)

// shadow tx pool``
var shadow_txpool = make(map[string]int64)

//refrenced ouputs tx pool

// var referenced_outputs_txpool = make(map[ref_out_key]string)

// tx map
// "from" in one queue
var txMap = make(map[string]tx_map_val) //[from]

//transaction pool queue
//look for a hashmap where each value is a queue

func main() {
	fmt.Println("hello world")
	// counting bloom filter tx pool

	//initializations Test
	// cbf_txpool.AddKey("81eded2e9862c0a39a2e5cd7332459c4542757aa198690ab76f79344b7dce563")
	// fmt.Println(cbf_txpool.HasKey("81eded2e9862c0a39a2e5cd7332459c4542757aa198690ab76f79344b7dce563"))
	// fmt.Println("cbf: ", cbf_lookup("81eded2e9862c0a39a2e5cd7332459c4542757aa198690ab76f79344b7dce563"))
	// shadow_txpool["81eded2e9862c0a39a2e5cd7332459c4542757aa198690ab76f79344b7dce563"] = unix_time()
	// fmt.Println(shadow_txpool["81eded2e9862c0a39a2e5cd7332459c4542757aa198690ab76f79344b7dce563"])
	// lookup_val, flag := shadow_lookup("81ded2e9862c0a39a2e5cd7332459c4542757aa198690ab76f79344b7dce563")
	// fmt.Println("shadow: ", lookup_val, flag)
	// val := tx_map_val{gasPrice: 1100000000, nonce: 1, time: unix_time()}
	// val.bumpGasPrice(120000)
	// txMap["81eded2e9862c0a39a2e5cd7332459c4542757aa198690ab76f79344b7dce563"] = val

	// fmt.Println(txMap["81eded2e9862c0a39a2e5cd7332459c4542757aa198690ab76f79344b7dce563"])

	fmt.Println("Process Started")
	file, err := os.Open("log.json")
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	dec := json.NewDecoder(file)
	type Item map[string]interface{}

	// Read the open bracket.
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}
	fmt.Println("OK: reading file")
	simulationTime = unix_time()
	initialize(simulationTime)

	// While the json array contains values.
	for dec.More() {
		// Decode an array value.
		var item Item
		err := dec.Decode(&item)
		if err != nil {
			log.Fatal(err)
		}

		// PRINT OUT JSON OBJECT
		fmt.Println("\n", item["TYPE"])
		// fmt.Println(item["garbage"] == nil)

		for key, val := range item {

			fmt.Println(key, ": ", val)
		}

		fmt.Println("\n")

		switch item["TYPE"] {
		case "ENTRY":
			entryTx++

			txCbfLookup = cbf_lookup(item["txHash"].(string))
			fmt.Println("was the entry in cbf tx pool?: ", txCbfLookup)

			_, txShadowLookup = shadow_lookup(item["txHash"].(string))
			fmt.Println("was the entry in shadow tx pool?: ", txShadowLookup)

			//TODO:update performance

			if !txCbfLookup {

				txMapLookup, status := txMapLookup(item["from"].(string))
				if txMapLookup.nonce == item["nonce"] {
					sameNonce = true
				}
				if txMapLookup.gasPrice != item["gasPrice"] {
					gasPriceBump = true
				}

				if status && sameNonce && gasPriceBump {
					changeGasPrice(item["from"].(string), item["gasPrice"].(float64))
				} else if status == false {
					fmt.Println("NEW ITEM")
					addTxpool(item["txHash"].(string))
					addTxMap(item["from"].(string), tx_map_val{gasPrice: item["gasPrice"].(float64), nonce: item["nonce"].(float64), time: unix_time()})
					addShadowpool(item["txHash"].(string))
				}

			}

		case "EXIT":
			exitTx++

			//flush parameters
			txCbfLookup = false
			txShadowLookup = false

			txCbfLookup = cbf_lookup(item["txHash"].(string))
			fmt.Println("was the entry in cbf tx pool?: ", txCbfLookup)

			_, txShadowLookup = shadow_lookup(item["txHash"].(string))
			fmt.Println("was the entry in shadow tx pool?: ", txShadowLookup)

			// TODO: update performance metrics

			if txCbfLookup {
				switch item["REASON"] {
				case "ALREADY_KNOWN":
					blockTx++
					removeTxpool(item["txHash"].(string))
					removeTxMap(item["from"].(string), item["nonce"].(float64))

					removeShadowpool(item["txHash"].(string))
				case "INVALID":
					invalidTx++
					removeTxpool(item["txHash"].(string))
					removeTxMap(item["from"].(string), item["nonce"].(float64))

					removeShadowpool(item["txHash"].(string))

					//TODO:nothing to be done for underpriced and frshly underpriced exit tx for now. when i get actual logs i will have to find out if these transactions exist in our pool and we need to remove them here
				}

			}

		}

		// TODO: simulation time - if exceeded clear stuff out

	}
	// Read the closing bracket.
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	fmt.Println(txMap)
}
