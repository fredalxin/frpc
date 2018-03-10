package core

type SelectMode int

const (
	//RandomSelect is selecting randomly
	Random SelectMode = iota
	//RoundRobin is selecting by round robin
	RoundRobin
	//WeightedRoundRobin is selecting by weighted round robin
	Weighted
	//ConsistentHash is selecting by hashing
	Hash
	//WeightedICMP is selecting by weighted Ping time
	Ping
	//Closest is selecting the closest server
	Closest
)

type RegistryMode int

const (
	Etcd RegistryMode = iota
	Consul
)

type FailMode int

const (
	FailFast FailMode = iota
	FailOver
	FailTry
)