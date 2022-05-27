import pytest
from eth_bloom import BloomFilter
from eth_utils import abi, big_endian_to_int
from hexbytes import HexBytes

from integration_tests.utils import astra_to_aastra, deploy_contract, CONTRACTS, KEYS, ADDRS, send_transaction

pytestmark = pytest.mark.normal


def test_basic(astra):
    w3 = astra.w3
    assert w3.eth.chain_id == 777


def test_events(astra, suspend_capture):
    w3 = astra.w3
    erc20 = deploy_contract(
        w3,
        CONTRACTS["TestERC20A"],
        key=KEYS["validator"],
    )
    tx = erc20.functions.transfer(ADDRS["team"], 10).buildTransaction(
        {"from": ADDRS["validator"]}
    )
    txreceipt = send_transaction(w3, tx, KEYS["validator"])
    assert len(txreceipt.logs) == 1
    expect_log = {
        "address": erc20.address,
        "topics": [
            HexBytes(
                abi.event_signature_to_log_topic("Transfer(address,address,uint256)")
            ),
            HexBytes(b"\x00" * 12 + HexBytes(ADDRS["validator"])),
            HexBytes(b"\x00" * 12 + HexBytes(ADDRS["team"])),
        ],
        "data": "0x000000000000000000000000000000000000000000000000000000000000000a",
        "transactionIndex": 0,
        "logIndex": 0,
        "removed": False,
    }
    assert expect_log.items() <= txreceipt.logs[0].items()

    # check block bloom
    bloom = BloomFilter(
        big_endian_to_int(w3.eth.get_block(txreceipt.blockNumber).logsBloom)
    )
    assert HexBytes(erc20.address) in bloom
    for topic in expect_log["topics"]:
        assert topic in bloom


def test_minimal_gas_price(astra):
    w3 = astra.w3
    gas_price = w3.eth.gas_price
    tx = {
        "to": "0x0000000000000000000000000000000000000000",
        "value": 10000,
    }
    with pytest.raises(ValueError):
        send_transaction(
            w3,
            {**tx, "gasPrice": 1},
            KEYS["team"],
        )
    receipt = send_transaction(
        w3,
        {**tx, "gasPrice": gas_price},
        KEYS["validator"],
    )
    assert receipt.status == 1


def test_simple(astra):
    """
    - check number of validators
    - check vesting account status
    """
    assert len(astra.cosmos_cli(0).validators()) == 2

    # check vesting account
    addr = astra.cosmos_cli(0).address("team")
    account = astra.cosmos_cli(0).account(addr)
    assert account["@type"] == "/ethermint.types.v1.EthAccount"


def test_transfer(astra):
    """
    check simple transfer tx success
    - send 1astra from team to treasury
    """
    team_addr = astra.cosmos_cli(0).address("team")
    treasury_addr = astra.cosmos_cli(0).address("treasury")

    team_balance = astra.cosmos_cli(0).balance(team_addr)
    treasury_balance = astra.cosmos_cli(0).balance(treasury_addr)

    amount_astra = 1
    amount_aastra = astra_to_aastra(amount_astra)

    tx = astra.cosmos_cli(0).transfer(team_addr, treasury_addr, str(amount_astra) + "astra")
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

    assert astra.cosmos_cli(0).balance(team_addr) == team_balance - amount_aastra
    assert astra.cosmos_cli(0).balance(treasury_addr) == treasury_balance + amount_aastra
