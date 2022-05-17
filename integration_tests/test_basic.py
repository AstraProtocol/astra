from integration_tests.utils import astra_to_aastra


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
