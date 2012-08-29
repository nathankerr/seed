package main

var no_s2s_changes = seedTransformations{
	"000": no_s2s,
	"001": no_s2s,
	"00n": no_s2s,
	"010": no_s2s,
	"011": no_s2s,
	"01n": no_s2s,
	"0n0": no_s2s,
	"0n1": no_s2s,
	"0nn": no_s2s,
	"100": no_s2s,
	"101": no_s2s,
	"10n": no_s2s,
	"110": no_s2s,
	"111": no_s2s,
	"11n": no_s2s,
	"1n0": no_s2s,
	"1n1": no_s2s,
	"1nn": no_s2s,
	"n00": no_s2s,
	"n01": no_s2s,
	"n0n": no_s2s,
	"n10": no_s2s,
	"n11": no_s2s,
	"n1n": no_s2s,
	"nn0": no_s2s,
	"nn1": no_s2s,
	"nnn": no_s2s,
}

func no_s2s(seeds seedCollection, cluster *cluster, seed *seed, sname string) (sc seedCollection, delete_seed bool) {
	transformationinfo()
	// no-op
	return seeds, false
}

var no_s2b_changes = seedToBudTransformations{
	"000": no_s2b,
	"001": no_s2b,
	"00n": no_s2b,
	"010": no_s2b,
	"011": no_s2b,
	"01n": no_s2b,
	"0n0": no_s2b,
	"0n1": no_s2b,
	"0nn": no_s2b,
	"100": no_s2b,
	"101": no_s2b,
	"10n": no_s2b,
	"110": no_s2b,
	"111": no_s2b,
	"11n": no_s2b,
	"1n0": no_s2b,
	"1n1": no_s2b,
	"1nn": no_s2b,
	"n00": no_s2b,
	"n01": no_s2b,
	"n0n": no_s2b,
	"n10": no_s2b,
	"n11": no_s2b,
	"n1n": no_s2b,
	"nn0": no_s2b,
	"nn1": no_s2b,
	"nnn": no_s2b,
}

func no_s2b(buds budCollection, cluster *cluster, seed *seed, sname string) budCollection {
	transformationinfo()
	// no-op
	return buds
}

func no_b2b_changes(buds budCollection) budCollection {
	transformationinfo()
	return buds
}
