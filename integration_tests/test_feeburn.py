from time import sleep
import pytest
from pathlib import Path

from eth_bloom import BloomFilter
from eth_utils import abi, big_endian_to_int
from hexbytes import HexBytes
from integration_tests.network import setup_astra

from integration_tests.utils import astra_to_aastra, deploy_contract, CONTRACTS, KEYS, ADDRS, send_transaction, \
    wait_for_block, wait_for_new_blocks, GAS_USE, DEFAULT_BASE_PORT

pytestmark = pytest.mark.feeburn

genesis_total_supply = 5000000000000000000000


@pytest.fixture(scope="module")
def astra(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    cfg = Path(__file__).parent / "configs/feeburn.yaml"
    yield from setup_astra(path, DEFAULT_BASE_PORT, cfg)


def test_transfer(astra):
    """
    check simple transfer tx success
    - send 1astra from team to treasury
    """
    team_addr = astra.cosmos_cli(0).address("team")
    addr = "astra1wyzq5uv53tf7lqxn4qjmujlg8fcsmtafhs97ph"
    amount_astra = 1
    amount_aastra = astra_to_aastra(amount_astra)
    fee_coins = 1000000000
    old_block_height = astra.cosmos_cli(0).block_height()
    tx = astra.cosmos_cli(0).transfer(team_addr, addr, str(amount_astra) + "astra", fees="%saastra" % fee_coins)
    tx_block_height = int(tx["height"])
    print("tx_block_height", tx_block_height)
    assert tx["logs"] == [
        {
            "events": [
                {
                    "attributes": [
                        {"key": "receiver", "value": addr},
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
                        {"key": "recipient", "value": addr},
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
    new_total_minted_provision = int(astra.cosmos_cli(0).total_minted_provision())
    new_total_supply = int(astra.cosmos_cli(0).total_supply()["supply"][0]["amount"])
    total_fee_burn = int(astra.cosmos_cli(0).total_fee_burn())
    print("total_fee_burn", total_fee_burn)
    assert genesis_total_supply + new_total_minted_provision == new_total_supply + total_fee_burn
    assert total_fee_burn == int(fee_coins / 2)


def test_no_tx(astra):
    old_block_height = int(astra.cosmos_cli(0).block_height())
    print("old_block_height", old_block_height)
    old_total_supply = int(astra.cosmos_cli(0).total_supply()["supply"][0]["amount"])
    print("old_total_supply", old_total_supply)
    wait_for_new_blocks(astra.cosmos_cli(0), 1)
    new_total_supply = int(astra.cosmos_cli(0).total_supply()["supply"][0]["amount"])
    print("new_total_supply", new_total_supply)
    block_provisions = int(astra.cosmos_cli(0).block_provisions())
    print("block_provisions", block_provisions)
    fee_burn = block_provisions - (new_total_supply - old_total_supply)
    print(new_total_supply - old_total_supply, fee_burn)
    new_block_height = int(astra.cosmos_cli(0).block_height())
    assert new_block_height == old_block_height + 1
    assert fee_burn == 0
