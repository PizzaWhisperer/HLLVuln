# HLLVuln

doc/ folder contains paper's sources

attacks/ folder contains code and experiments

## Attacks

This code demonstrates how an attacker can perform a cardinality manipulation attack on the HyperLogLog sketch when having access to the low level details of the implementation.

Both against a popular off-the-shelf implementation ([clarkduvall's](https://github.com/clarkduvall/hyperloglog)) and [Redis's](https://github.com/redis/redis).


### Attacking Clark Duval's implementation

Running

`go run attack_hll.go`

will print some attack results.

// Depending on your setup you might have a module error, `GO111MODULE=off go run attack_hll.go` does the trick.

### Attacking Redis implementation


## Note

Be aware that we set the number of iterations to 50, causing the attacks results to take quite some time to be printed. You can lower this number to have results faster.
