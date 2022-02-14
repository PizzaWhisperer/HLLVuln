# HLLVuln

latex/ folder contains paper's sources

code/ folder contains code and experiments

## Attacks

This code demonstrates how an attacker can perform a cardinality manipulation attack on the HyperLogLog sketch when having access to the low level details of the implementation.

Both against a popular off-the-shelf implementation ([clarkduvall's](https://github.com/clarkduvall/hyperloglog)) and [Redis's](https://github.com/redis/redis).


### Attacking Clark Duval's implementation

Running

`go run attack_hll.go`

will print some attack results.

Depending on your setup you might have a module error, `GO111MODULE=off go run attack_hll.go` does the trick.

#### Note

Be aware that we set the number of iterations to 50, causing the attacks results to take quite some time to be printed. You can lower this number to have results faster.

### Attacking Redis implementation

First run `make` to install Reids, then spawn the server in a first tab with `./src/redis-server`, and finally the client in a second tab with `./src/redis-cli`. Running `PFADD_ATTACK` on the client side will add the items to the server's HLL sketch and return the number of inserted items. The initial cardinality of the sketch will be printed on the server's console (after inserting `RAND_ITEMS` honest items first). To observe the resulting change in cardinality, the command `PFCOUNT` can be called after the attack.

The number of initial items to insert `RAND_ITEMS` and the number of buckets targeted by the attacker `B` can be modified in the lines 208 and 209 respectively of the `src/hyperloglog.c` file.
