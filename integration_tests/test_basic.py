from integration_tests.utils import astra_to_aastra, wait_for_block


def test_simple(cluster):
    """
    - check number of validators
    - check vesting account status
    """
    assert len(cluster.validators()) == 2

    # check vesting account
    addr = cluster.address("team")
    account = cluster.account(addr)
    assert account["@type"] == "/ethermint.types.v1.EthAccount"


def test_transfer(cluster):
    """
    check simple transfer tx success
    - send 1astra from team to treasury
    """
    team_addr = cluster.address("team")
    treasury_addr = cluster.address("treasury")

    team_balance = cluster.balance(team_addr)
    treasury_balance = cluster.balance(treasury_addr)

    amount_astra = 1
    amount_aastra = astra_to_aastra(amount_astra)

    tx = cluster.transfer(team_addr, treasury_addr, str(amount_astra) + "astra")
    print("transfer tx", tx["txhash"])
    assert tx["logs"] == [
        {
            "events": [
                {
                    "attributes": [
                        {"key": "receiver", "value": treasury_addr},
                        {"key": "amount", "value": str(amount_aastra) + "aastra"},
                    ],
                    "type": "coin_received",
                },
                {
                    "attributes": [
                        {"key": "spender", "value": team_addr},
                        {"key": "amount", "value": str(amount_aastra) + "aastra"},
                    ],
                    "type": "coin_spent",
                },
                {
                    "attributes": [
                        {"key": "action", "value": "/cosmos.bank.v1beta1.MsgSend"},
                        {"key": "sender", "value": team_addr},
                        {"key": "module", "value": "bank"},
                    ],
                    "type": "message",
                },
                {
                    "attributes": [
                        {"key": "recipient", "value": treasury_addr},
                        {"key": "sender", "value": team_addr},
                        {"key": "amount", "value": str(amount_aastra) + "aastra"},
                    ],
                    "type": "transfer",
                },
            ],
            "log": "",
            "msg_index": 0,
        }
    ]

    assert cluster.balance(team_addr) == team_balance - amount_aastra
    assert cluster.balance(treasury_addr) == treasury_balance + amount_aastra


def test_statesync(cluster):
    """
    - create a new node with statesync enabled
    - check it works
    """
    # wait the first snapshot to be created
    wait_for_block(cluster, 10)

    # add a statesync node
    i = cluster.create_node(moniker="statesync", statesync=True)
    cluster.supervisor.startProcess(f"{cluster.chain_id}-node{i}")

    # discovery_time is set to 5 seconds, add extra seconds for processing
    wait_for_block(cluster.cosmos_cli(i), 10)
    print("succesfully syncing")
