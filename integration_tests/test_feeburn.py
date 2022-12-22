from time import sleep
import pytest
from pathlib import Path

from eth_bloom import BloomFilter
from eth_utils import abi, big_endian_to_int
from hexbytes import HexBytes
from integration_tests.network import setup_astra

from integration_tests.utils import astra_to_aastra, deploy_contract, CONTRACTS, KEYS, ADDRS, send_transaction, wait_for_block, GAS_USE, DEFAULT_BASE_PORT

pytestmark = pytest.mark.feeburn


@pytest.fixture(scope="module")
def astra(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    cfg = Path(__file__).parent / "configs/new.yaml"
    yield from setup_astra(path, DEFAULT_BASE_PORT, cfg)


def test_transfer(astra):
    """
    check simple transfer tx success
    - send 1astra from team to treasury
    """
    current_total_supply = int(float(astra.cosmos_cli(0).total_supply()["supply"][0]["amount"]))
    print("current_total_supply", current_total_supply)
    team_addr = astra.cosmos_cli(0).address("team")
    addr = "astra1wyzq5uv53tf7lqxn4qjmujlg8fcsmtafhs97ph"

    team_balance = astra.cosmos_cli(0).balance(team_addr)

    amount_astra = 10
    amount_aastra = astra_to_aastra(amount_astra)
    fee_coins = 10

    # tx = astra.cosmos_cli(0).transfer(team_addr, addr, str(amount_astra) + "astra", fees="%saastra" % fee_coins)
    # print("transfer tx", tx["txhash"])
#     assert tx["logs"] == [
#         {
#             "events": [
#                 {
#                     "attributes": [
#                         {"key": "receiver", "value": addr},
#                         {"key": "amount", "value": str(amount_aastra) + "aastra"},
#                     ],
#                     "type": "coin_received",
#                 },
#                 {
#                     "attributes": [
#                         {"key": "spender", "value": team_addr},
#                         {"key": "amount", "value": str(amount_aastra) + "aastra"},
#                     ],
#                     "type": "coin_spent",
#                 },
#                 {
#                     "attributes": [
#                         {"key": "action", "value": "/cosmos.bank.v1beta1.MsgSend"},
#                         {"key": "sender", "value": team_addr},
#                         {"key": "module", "value": "bank"},
#                     ],
#                     "type": "message",
#                 },
#                 {
#                     "attributes": [
#                         {"key": "recipient", "value": addr},
#                         {"key": "sender", "value": team_addr},
#                         {"key": "amount", "value": str(amount_aastra) + "aastra"},
#                     ],
#                     "type": "transfer",
#                 },
#             ],
#             "log": "",
#             "msg_index": 0,
#         }
#     ]

    wait_for_block(astra.cosmos_cli(0), 2)

    new_total_supply = int(float(astra.cosmos_cli(0).total_supply()["supply"][0]["amount"]))
    assert new_total_supply == current_total_supply - int(fee_coins / 2)