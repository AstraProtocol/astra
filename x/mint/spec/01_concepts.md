<!--
order: 0
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

## ASA Token Model

The Astra Token Model outlines how the Astra network is secured through a
balanced incentivized interest from users, developers and validators. In this
model, inflation plays a major role in sustaining this balance. At launch, there is a total of
1,200,000,000 ASAs being minted. Afterwards, there will be more ASAs minted through inflation following
the minting mechanism presented below, with an initial inflation set to 10%.

### Genesis Distribution
At the genesis launch, there is a total of 1,200,000,000 ASAs distributed to Genesis Partners, Strategic Partners (Backers),
Astra Foundation, Core Team, and the Community Pool. Here are the details:

|                        |    % |         #ASAs | Minted at genesis |                                                                                  Vesting schedule |
|:-----------------------|-----:|--------------:|------------------:|--------------------------------------------------------------------------------------------------:|
| **Genesis Partners**   |  35% |   420,000,000 |       420,000,000 |                                                               100% minted and unlocked at genesis |
| **Strategic Partners** |  40% |   480,000,000 |       480,000,000 |                                                               100% minted and unlocked at genesis |
| **Astra Foundation**   |  10% |   120,000,000 |       120,000,000 |                                                               100% minted and unlocked at genesis |
| **Core Team**          |  10% |   120,000,000 |       120,000,000 |                                                50% unlocked at genesis, 50% unlocked after 1 year |
| **Community Pool**     |   5% |    60,000,000 |        60,000,000 |                                          100% minted & locked at genesis, unlocked with GOV rules |
| **Total**              | 100% | 1,200,000,000 |     1,200,000,000 |                                                                                                   |

## The Minting Mechanism

The minting mechanism was designed to:

- allow for a flexible inflation rate determined by market demand targeting a particular bonded-stake ratio
- effect a balance between market liquidity and staked supply

In order to best determine the appropriate market rate for inflation rewards, a
moving change rate is used.  The moving change rate mechanism ensures that if
the % bonded is either over or under the goal %-bonded, the inflation rate will
adjust to further incentivize or disincentivize being bonded, respectively. Setting the goal
%-bonded at less than 100% encourages the network to maintain some non-staked tokens
which should help provide some liquidity.

It can be broken down in the following way (with detail presented [here](./03_begin_block.md)):

- If the inflation rate is below the goal %-bonded the inflation rate will
   increase until a maximum value is reached
- If the goal % bonded (i.e, 50%) is maintained, then the inflation
   rate will stay constant
- If the inflation rate is above the goal %-bonded the inflation rate will
   decrease until a minimum value is reached