<!--
order: 1
-->

# Concepts

## Inflation

In a Proof of Stake (PoS) blockchain, inflation is used as a tool to incentivize
participation in the network. Inflation creates and distributes new tokens to
participants who can use their tokens to either interact with the protocol or
stake their assets to earn rewards and vote for governance proposals.

Especially in an early stage of a network, where staking rewards are high and
there are fewer possibilities to interact with the network, inflation can be
used as the major tool to incentivize staking and thereby securing the network.

With more stakers, the network becomes increasingly stable and decentralized. It
becomes *stable*, because assets are locked up instead of causing price changes
through trading. And it becomes *decentralized,* because the power to vote for
governance proposals is distributed amongst more people.

## Astra Token Model

The Astra Token Model outlines how the Astra network is secured through a
balanced incentivized interest from users, developers and validators. In this
model, inflation plays a major role in sustaining this balance. With an initial
supply of 200 million and over 300 million tokens being issued through inflation
during the first year, the model suggests an exponential decline in inflation to
the rest of Astras in subsequent years.

We implement two different inflation mechanisms to support the token model:

1. linear inflation for team vesting, reward providers; and
2. exponential inflation for staking rewards, dev/community incentives and reserve treasury.

### Linear Inflation - Team Vesting

The Team Vesting distribution in the Token Model is implemented in a way that
minimized the amount of taxable events. An initial supply of 200M allocated to
`vesting accounts` at genesis. This amount is equal to the total inflation
allocatod for team vesting after 4 years (`25% * 1B = 250M`). Over time,
`unvested` tokens on these accounts are converted into `vested` tokens at a
linear rate. Team members cannot delegate, transfer or execute Ethereum
transaction with `unvested` tokens until they are unlocked represented as
`vested` tokens.

### Exponential Inflation - Block Rewards & Reserve Treasury
There will be a total of 30% of the total supply that will be distributed as block rewards. Block rewards are mainly 
distributed as Staking Rewards and Reserve Treasury with the following distribution:
- **Staking Rewards**: 66.67%
- **Reserve Treasury**: 33.33%.

Block rewards are minted in daily epochs, via a decay function. During a period of 365 epochs (i.e, one year), a
daily provision of Astra tokens is minted and allocated to staking rewards and reserve treasury. Within a period,
the epoch provision does not change. The epoch provision is then reduced by a factor of `(1-r)` for subsequent years, 
with a decay factor `r`. Precisely, at the end of each period, the provision is recalculated as follows:
```latex
f(x) = r * (1 - r)^x * R
where
    x = variable = period (i.e, year)
    r = 0.1 = decay factor
    R = total amount of block rewards
```

With the given formula of `f(x)`, we can make sure that the total of minted block rewards never exceed `R`, as:
```latex
R' = \sum_{k=0}^{\inf}{r*(1-r)^k*R} = r*R*\sum_{k=0}^{\inf}{(1-r)^k} = r*R*1/(1-(1-r)) = r*R/r = R
```

With a decay factor of `0.1`, there will be a total of 69%, 89%, 98% of block rewards minted after the first 10, 20, 40 years,
respectively. As a result, most of the block rewards will be distributed after 40 years. The decay factor can be changed via 
governance voting. A higher decay factor means it takes less time to mint most of the block rewards while a lower
decay factor results in a longer minting period.