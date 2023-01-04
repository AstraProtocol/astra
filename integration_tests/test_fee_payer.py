import json
import time

import pytest

from .utils import sign_single_tx_with_options, wait_for_new_blocks, DEFAULT_BASE_PORT, wait_for_port, DEFAULT_GAS

pytestmark = pytest.mark.normal


def test_different_fee_payer(astra, tmp_path):
    wait_for_port(DEFAULT_BASE_PORT)
    wait_for_new_blocks(astra.cosmos_cli(0), 1)
    transaction_coins = 100
    fee_coins = 10

    receiver_addr = astra.cosmos_cli(0).address("community")
    sender_addr = astra.cosmos_cli(0).address("signer1")
    fee_payer_addr = astra.cosmos_cli(0).address("signer2")

    unsigned_tx_txt = tmp_path / "unsigned_tx.txt"
    partial_sign_txt = tmp_path / "partial_sign.txt"
    signed_txt = tmp_path / "signed.txt"

    receiver_balance = astra.cosmos_cli(0).balance(receiver_addr)
    sender_balance = astra.cosmos_cli(0).balance(sender_addr)
    fee_payer_balance = astra.cosmos_cli(0).balance(fee_payer_addr)
    print("receiver_balance", receiver_balance)
    print("sender_balance", sender_balance)
    print("fee_payer_balance", fee_payer_balance)

    unsigned_tx_msg = astra.cosmos_cli(0).transfer(
        sender_addr,
        receiver_addr,
        "%saastra" % transaction_coins,
        generate_only=True,
        fees="%saastra" % fee_coins,
    )

    unsigned_tx_msg["auth_info"]["fee"]["payer"] = fee_payer_addr
    unsigned_tx_msg["auth_info"]["fee"]["gas_limit"] = DEFAULT_GAS
    with open(unsigned_tx_txt, "w") as opened_file:
        json.dump(unsigned_tx_msg, opened_file)
    partial_sign_tx_msg = sign_single_tx_with_options(
        astra, unsigned_tx_txt, "signer1", sign_mode="amino-json"
    )
    with open(partial_sign_txt, "w") as opened_file:
        json.dump(partial_sign_tx_msg, opened_file)
    signed_tx_msg = sign_single_tx_with_options(
        astra, partial_sign_txt, "signer2", sign_mode="amino-json"
    )
    with open(signed_txt, "w") as opened_file:
        json.dump(signed_tx_msg, opened_file)
    wait_for_new_blocks(astra.cosmos_cli(0), 1)
    tx_result = astra.cosmos_cli(0).broadcast_tx(signed_txt)
    assert int(tx_result["height"]) > 0
    assert tx_result["code"] == 0

    wait_for_new_blocks(astra.cosmos_cli(0), 2)

    assert astra.cosmos_cli(0).balance(receiver_addr) == receiver_balance + transaction_coins
    assert astra.cosmos_cli(0).balance(sender_addr) == sender_balance - transaction_coins
    assert astra.cosmos_cli(0).balance(fee_payer_addr) == fee_payer_balance - fee_coins
