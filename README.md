# qenvs

automation for qe environments using pulumi

[![Container Repository on Quay](https://quay.io/repository/ariobolo/qenvs/status "Container Repository on Quay")](https://quay.io/repository/ariobolo/qenvs)

## Environment

Create a composable environment with different qe target machines aggregated on different topologies and with specific setups (like vpns, proxys, airgaps,...)

Current available features using cmd `qenvs corp create`

![Environment](./docs/diagrams/base.svg)

## Spot price use case

This module allows to check for best bid price on all regions, to request instances at lower price to reduce costs. To calculate the best option, it is also required to:  

* reduce interruptions
* ensure capacity

to check those requisites the module make use of spot placement scores based on machine requirements. Then best scores are crossed with lowers price from spot price history to pick the most valuable option.

Current use case is working on one machine but it will be exteded to analyze any required environment offered by qenvs (checking with all the machines included on a specific environment).

Current information about supported machines can be checked at [support-matrix](pkg/infra/aws/support-matrix/matrix.go)
