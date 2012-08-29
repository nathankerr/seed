package main

var split_seeds = seedTransformations{
	"000": split,
	"001": split,
	"00n": split,
	"010": split,
	"011": split,
	"01n": split,
	"0n0": split,
	"0n1": split,
	"0nn": split,
	"100": split,
	"101": split,
	"10n": split,
	"110": split,
	"111": split,
	"11n": split,
	"1n0": split,
	"1n1": split,
	"1nn": split,
	"n00": split,
	"n01": split,
	"n0n": split,
	"n10": split,
	"n11": split,
	"n1n": split,
	"nn0": split,
	"nn1": split,
	"nnn": split,
}

func split(seeds seedCollection, cluster *cluster, seed *seed, sname string) (sc seedCollection, delete_seed bool) {
	return seeds, true
}
