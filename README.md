# HLLVuln

doc/ folder contains paper's sources

attacks/ folder contains code and experiments

## Attacks

This code demonstrates how an attacker can perform a cardinality manipulation attack on the HyperLogLog sketch when having access to the low level details of the implementation.

Both against an off-the-shelf implementation and Redis.

### Attacking Clark Duval's implementation

`go build` then `./HLLVuln`

Flag `-scenario` can be set to `"S2"` or `"S3"` to specify the adversarial model.
Flag `-userData` can be set to `X` to add `X` items in the sketch before attacking it.
Flag `-RT20` can be set to true to perform the attack under the setting of RT20 paper.

As example, running `./HLLVuln -scenario="S2" -userData=1000` will perform the attack under S2 scenario on a sketch that previously observed 1000 items from a honest user.

#### Benchmarking

Flag `-iterations` can be set to any int `X` to loop `X` times and output the statistics.
Flag `-log` can be set to true to print more (debug mode).

### Attacking Redis implementation
