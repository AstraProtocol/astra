import json

import pytest

from integration_tests.network import setup_astra
from .utils import (AASTRA_DENOM, AUTHORIZATION_GENERIC, AUTHORIZATION_SEND, AUTHZ,
                    BLOCK_BROADCASTING, DEFAULT_BASE_PORT,
                    GENERATE_ONLY, GRANTS,
                    SEND_MSG_TYPE_URL,
                    WITHDRAW_DELEGATOR_REWARD_TYPE_URL,
                    delegate_amount, exec_tx_by_grantee, grant_authorization,
                    query_command, query_total_reward_amount, revoke_authorization, transfer, wait_for_new_blocks,
                    withdraw_all_rewards, wait_for_block, wait_for_new_epochs)

pytestmark = pytest.mark.authz


@pytest.fixture(scope="module")
def astra_temp(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    yield from setup_astra(path, DEFAULT_BASE_PORT)


def test_execute_tx_within_authorization_spend_limit(astra_temp, tmp_path):
    """
    test execute transaction within send authorization spend limit
    """
    wait_for_block(astra_temp.cosmos_cli(0), 2)
    spend_limit = 200
    transaction_coins = 100
    granter_address = astra_temp.cosmos_cli(0).address("community")
    grantee_address = astra_temp.cosmos_cli(0).address("other_partner")
    receiver_address = astra_temp.cosmos_cli(0).address("treasury")
    granter_initial_balance = astra_temp.cosmos_cli(0).balance(granter_address)
    receiver_initial_balance = astra_temp.cosmos_cli(
        0).balance(receiver_address)

    grant_authorization(
        astra_temp,
        grantee_address,
        AUTHORIZATION_SEND,
        granter_address,
        spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
    )
    generated_tx_txt = tmp_path / "generated_tx.txt"
    generated_tx_msg = transfer(
        astra_temp,
        granter_address,
        receiver_address,
        "%s%s" % (transaction_coins, AASTRA_DENOM),
        GENERATE_ONLY,
    )
    with open(generated_tx_txt, "w") as opened_file:
        json.dump(generated_tx_msg, opened_file)
    exec_tx_by_grantee(
        astra_temp,
        generated_tx_txt,
        grantee_address,
        broadcast_mode=BLOCK_BROADCASTING,
    )
    wait_for_new_blocks(astra_temp.cosmos_cli(0), 1)

    assert (
            astra_temp.cosmos_cli(0).balance(granter_address)
            == granter_initial_balance - transaction_coins
    )
    assert (
            astra_temp.cosmos_cli(0).balance(receiver_address)
            == receiver_initial_balance + transaction_coins
    )

    # teardown
    revoke_authorization(
        astra_temp, grantee_address, SEND_MSG_TYPE_URL, granter_address
    )
    wait_for_block(astra_temp.cosmos_cli(0), 2)
    assert (
            len(
                query_command(astra_temp, AUTHZ, GRANTS,
                              granter_address, grantee_address)["grants"]
            ) == 0
    )


def test_execute_tx_beyond_authorization_spend_limit(astra_temp, tmp_path):
    """
    test execute transaction beyond send authorization spend limit
    """
    wait_for_block(astra_temp.cosmos_cli(0), 2)
    spend_limit = 50
    transaction_coins = 100
    granter_address = astra_temp.cosmos_cli(0).address("community")
    grantee_address = astra_temp.cosmos_cli(0).address("other_partner")
    receiver_address = astra_temp.cosmos_cli(0).address("treasury")
    granter_initial_balance = astra_temp.cosmos_cli(0).balance(granter_address)
    receiver_initial_balance = astra_temp.cosmos_cli(
        0).balance(receiver_address)

    rsp = grant_authorization(
        astra_temp,
        grantee_address,
        AUTHORIZATION_SEND,
        granter_address,
        spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    generated_tx_txt = tmp_path / "generated_tx.txt"
    generated_tx_msg = transfer(
        astra_temp,
        granter_address,
        receiver_address,
        "%s%s" % (transaction_coins, AASTRA_DENOM),
        GENERATE_ONLY,
    )
    with open(generated_tx_txt, "w") as opened_file:
        json.dump(generated_tx_msg, opened_file)

    with pytest.raises(
            Exception, match=r".*requested amount is more than spend limit.*"
    ):
        exec_tx_by_grantee(astra_temp, generated_tx_txt, grantee_address)
    assert astra_temp.cosmos_cli(0).balance(
        granter_address) == granter_initial_balance
    assert astra_temp.cosmos_cli(0).balance(
        receiver_address) == receiver_initial_balance

    # teardown
    revoke_authorization(
        astra_temp, grantee_address, SEND_MSG_TYPE_URL, granter_address
    )
    wait_for_block(astra_temp.cosmos_cli(0), 2)

    assert (
            len(query_command(astra_temp, AUTHZ, GRANTS,
                              granter_address, grantee_address)["grants"]) == 0
    )


def test_generic_authorization_grant(astra_temp, tmp_path):
    """
    test grant authorization with generic authorization with send msg type
    """
    delegate_coins = 1000000
    validator_address = astra_temp.cosmos_cli(0).validators()[0]["operator_address"]
    granter_address = astra_temp.cosmos_cli(0).address("community")
    grantee_address = astra_temp.cosmos_cli(0).address("other_partner")
    granter_initial_reward_amount = query_total_reward_amount(
        astra_temp, granter_address, validator_address
    )
    rsp = delegate_amount(
        astra_temp,
        validator_address,
        "%s%s" % (delegate_coins, AASTRA_DENOM),
        granter_address,
        broadcast_mode=BLOCK_BROADCASTING,
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    # wait for some reward
    wait_for_new_blocks(astra_temp.cosmos_cli(0), 2)
    granter_balance_bef_rewards_withdrawal = astra_temp.cosmos_cli(0).balance(granter_address)
    rsp = grant_authorization(
        astra_temp,
        grantee_address,
        AUTHORIZATION_GENERIC,
        granter_address,
        msg_type=WITHDRAW_DELEGATOR_REWARD_TYPE_URL,
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    # wait epochs release token
    wait_for_new_epochs(astra_temp.cosmos_cli(0),
                        epoch_identifier=astra_temp.cosmos_cli(0).get_inflation_epoch_identifier())
    generated_tx_txt = tmp_path / "generated_tx.txt"
    generated_tx_msg = withdraw_all_rewards(
        astra_temp,
        granter_address,
        GENERATE_ONLY,
    )
    with open(generated_tx_txt, "w") as opened_file:
        json.dump(generated_tx_msg, opened_file)
    rsp = exec_tx_by_grantee(
        astra_temp,
        generated_tx_txt,
        grantee_address,
        broadcast_mode=BLOCK_BROADCASTING,
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    wait_for_new_blocks(astra_temp.cosmos_cli(0), 1)

    assert (
            astra_temp.cosmos_cli(0).balance(granter_address) - granter_balance_bef_rewards_withdrawal > granter_initial_reward_amount
    )

    # teardown
    rsp = revoke_authorization(
        astra_temp,
        grantee_address,
        WITHDRAW_DELEGATOR_REWARD_TYPE_URL,
        granter_address,
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    wait_for_block(astra_temp.cosmos_cli(0), 2)
    assert (
            len(query_command(astra_temp, AUTHZ, GRANTS,
                              granter_address, grantee_address)["grants"]) == 0
    )
