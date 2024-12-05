package main

func initialize(simulationTime int64) {
	lastExpiryTime = simulationTime
	lastDebugTime = simulationTime
	//        last_rejuvenate_time=sim_time;

}
func cbf_lookup(tx string) bool {
	cbf_lookup := cbf_txpool.HasKey(tx)
	//fmt.Println(cbf_lookup)
	return cbf_lookup

}
func shadow_lookup(tx string) (int64, bool) {
	val, cbf_lookup := shadow_txpool[tx]
	return val, cbf_lookup

}
func txMapLookup(from string) (tx_map_val, bool) {
	// var txMap_lookup tx_map_val
	txMapLookup, status := txMap[from]
	return txMapLookup, status
}

func addTxpool(tx string) {
	cbf_txpool.AddKey((tx))
}
func addTxMap(from string, txMapVal tx_map_val) {
	txMap[from] = txMapVal
}
func addShadowpool(tx string) {
	shadow_txpool[tx] = unix_time()
}

func changeGasPrice(from string, gasPrice float64) {
	txMapLookup, _ := txMapLookup(from)
	txMapLookup.bumpGasPrice(gasPrice)
	addTxMap(from, txMapLookup)
}

func removeTxpool(tx string) bool {
	return cbf_txpool.RemoveKey(tx)
}
func removeTxMap(from string, nonce float64) bool {
	tx, ok := txMapLookup(from)
	if ok && tx.nonce == nonce {
		delete(txMap, from)
		return true
	}
	return false
}
func removeShadowpool(tx string) {
	delete(shadow_txpool, tx)
}

func remove_expired(simulationTime int) {}

//end
