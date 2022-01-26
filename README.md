# HLLVuln

doc/ folder contains paper's sources

attacks/ folder contains code and experiments

## Attacks

This code demonstrates how an attacker can perform a cardinality manipulation attack on the HyperLogLog sketch when having access to the low level details of the implementation.

Both against an off-the-shelf implementation and Redis.

### Attacking Clark Duval's implementation

Running 

`GO111MODULE=off go run attack_hll.go` 

will print some attack results.

### Attacking Redis implementation
