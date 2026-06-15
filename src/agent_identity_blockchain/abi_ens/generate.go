package abi_ens

// The ENS registry binding keeps its own package because abigen emits the shared
// KeyValue struct per-contract; folding it into ./abi would double-declare it.
// It reuses the sibling ./abi/gen generators.

//go:generate go run ../abi/gen/abigen_jsons.go . abi_ens
//go:generate go run ../abi/gen/add_raw_methods.go .
