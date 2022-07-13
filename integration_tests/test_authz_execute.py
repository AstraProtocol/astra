import json
from datetime import timedelta

import pytest
from dateutil.parser import isoparse

from integration_tests.network import setup_astra
from .utils import (AASTRA_DENOM, AUTHORIZATION_DELEGATE,
                    AUTHORIZATION_REDELEGATE,
                    AUTHORIZATION_UNBOND, AUTHZ,
                    DEFAULT_BASE_PORT,
                    DELEGATE_MSG_TYPE_URL, GENERATE_ONLY, GRANTS,
                    REDELEGATE_MSG_TYPE_URL, UNBOND_MSG_TYPE_URL, delegate_amount, exec_tx_by_grantee, grant_authorization,
                    parse_events, query_command, query_delegation_amount,
                    redelegate_amount,
                    revoke_authorization, unbond_amount,
                    wait_for_block_time, wait_for_block)

pytestmark = pytest.mark.authz_execute


@pytest.fixture(scope="module")
def astra_temp(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    yield from setup_astra(path, DEFAULT_BASE_PORT)


def test_execute_delegate_to_allowed_validator(astra_temp, tmp_path):
    """
    test execute delegate to allowed validator should succeed
    test execute delegate to other validators should fail
    """
    # test execute delegate to allowed validator should succeed
    wait_for_block(astra_temp.cosmos_cli(0), 2)
    spend_limit = 200
    delegate_coins = 100
    allowed_validator_address = astra_temp.cosmos_cli(0).validators()[
        0]["operator_address"]
    another_validator_address = astra_temp.cosmos_cli(0).validators()[
        1]["operator_address"]
    granter_address = astra_temp.cosmos_cli(0).address("community")
    grantee_address = astra_temp.cosmos_cli(0).address("other_partner")
    validator_initial_delegation_amount = int(
        query_delegation_amount(
            astra_temp, granter_address, allowed_validator_address)["amount"]
    )

    grant_authorization(
        astra_temp,
        grantee_address,
        AUTHORIZATION_DELEGATE,
        granter_address,
        spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
        allowed_validators=allowed_validator_address,
    )
    generated_delegate_txt = tmp_path / "generated_delegate.txt"
    generated_delegate_msg1 = delegate_amount(
        astra_temp,
        allowed_validator_address,
        "%s%s" % (delegate_coins, AASTRA_DENOM),
        granter_address,
        GENERATE_ONLY,
    )
    with open(generated_delegate_txt, "w") as opened_file:
        json.dump(generated_delegate_msg1, opened_file)
    exec_tx_by_grantee(astra_temp, generated_delegate_txt, grantee_address)

    assert query_delegation_amount(
        astra_temp, granter_address, allowed_validator_address) == {
               "denom": AASTRA_DENOM,
               "amount": str(validator_initial_delegation_amount + delegate_coins),
           }

    # test execute delegate to other validators not in allowed validators should fail # noqa: E501
    another_validator_initial_delegation_amount = int(
        query_delegation_amount(
            astra_temp, granter_address, another_validator_address)["amount"]
    )
    generated_delegate_msg2 = delegate_amount(
        astra_temp,
        another_validator_address,
        "%s%s" % (delegate_coins, AASTRA_DENOM),
        granter_address,
        GENERATE_ONLY,
    )
    with open(generated_delegate_txt, "w") as opened_file:
        json.dump(generated_delegate_msg2, opened_file)

    with pytest.raises(Exception, match=r".*unauthorized.*"):
        exec_tx_by_grantee(astra_temp, generated_delegate_txt, grantee_address)
    assert query_delegation_amount(
        astra_temp, granter_address, another_validator_address) == {
               "denom": AASTRA_DENOM,
               "amount": str(another_validator_initial_delegation_amount),
           }

    # teardown
    revoke_authorization(
        astra_temp, grantee_address, DELEGATE_MSG_TYPE_URL, granter_address
    )
    wait_for_block(astra_temp.cosmos_cli(0), 2)
    assert (
            len(query_command(astra_temp, AUTHZ, GRANTS,
                              granter_address, grantee_address)["grants"]) == 0
    )


def test_unable_to_execute_delegate_to_deny_validator(astra_temp, tmp_path):
    """
    test execute delegate to deny validator should fail
    """
    wait_for_block(astra_temp.cosmos_cli(0), 2)
    spend_limit = 200
    delegate_coins = 100
    deny_validator_address = astra_temp.cosmos_cli(
        0).validators()[0]["operator_address"]
    granter_address = astra_temp.cosmos_cli(0).address("community")
    grantee_address = astra_temp.cosmos_cli(0).address("other_partner")

    # test execute delegate to deny validator should fail
    grant_authorization(
        astra_temp,
        grantee_address,
        AUTHORIZATION_DELEGATE,
        granter_address,
        spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
        deny_validators=deny_validator_address,
    )
    generated_delegate_txt = tmp_path / "generated_delegate.txt"
    generated_delegate_msg1 = delegate_amount(
        astra_temp,
        deny_validator_address,
        "%s%s" % (delegate_coins, AASTRA_DENOM),
        granter_address,
        GENERATE_ONLY,
    )
    with open(generated_delegate_txt, "w") as opened_file:
        json.dump(generated_delegate_msg1, opened_file)

    with pytest.raises(Exception, match=r".*unauthorized.*"):
        exec_tx_by_grantee(astra_temp, generated_delegate_txt, grantee_address)

    # teardown
    revoke_authorization(
        astra_temp, grantee_address, DELEGATE_MSG_TYPE_URL, granter_address
    )
    wait_for_block(astra_temp.cosmos_cli(0), 2)
    assert (
            len(query_command(astra_temp, AUTHZ, GRANTS,
                              granter_address, grantee_address)["grants"]) == 0
    )


def test_execute_all_staking_operations(astra_temp, tmp_path):
    """
    test execute delegate, unbond, redelegate by grantee
    """
    wait_for_block(astra_temp.cosmos_cli(0), 2)
    spend_limit = 200
    delegate_coins = 100
    unbond_coins = 50
    redelegate_coins = 30
    validator1_address = astra_temp.cosmos_cli(
        0).validators()[0]["operator_address"]
    validator2_address = astra_temp.cosmos_cli(
        0).validators()[1]["operator_address"]
    granter_address = astra_temp.cosmos_cli(0).address("community")
    grantee_address = astra_temp.cosmos_cli(0).address("other_partner")
    validator1_initial_deligation_amount = int(
        query_delegation_amount(
            astra_temp, granter_address, validator1_address)["amount"]
    )
    validator2_initial_deligation_amount = int(
        query_delegation_amount(
            astra_temp, granter_address, validator2_address)["amount"]
    )

    # test execute delegate
    grant_authorization(
        astra_temp,
        grantee_address,
        AUTHORIZATION_DELEGATE,
        granter_address,
        spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
        allowed_validators=validator1_address,
    )
    generated_delegate_txt = tmp_path / "generated_delegate.txt"
    generated_delegate_msg = delegate_amount(
        astra_temp,
        validator1_address,
        "%s%s" % (delegate_coins, AASTRA_DENOM),
        granter_address,
        GENERATE_ONLY,
    )
    with open(generated_delegate_txt, "w") as opened_file:
        json.dump(generated_delegate_msg, opened_file)
    exec_tx_by_grantee(astra_temp, generated_delegate_txt, grantee_address)

    assert query_delegation_amount(astra_temp, granter_address, validator1_address) == {
        "denom": AASTRA_DENOM,
        "amount": str(validator1_initial_deligation_amount + delegate_coins),
    }

    # test execute unbond
    grant_authorization(
        astra_temp,
        grantee_address,
        AUTHORIZATION_UNBOND,
        granter_address,
        spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
        allowed_validators=validator1_address,
    )
    generated_unbond_txt = tmp_path / "generated_unbond.txt"
    generated_unbond_msg = unbond_amount(
        astra_temp,
        validator1_address,
        "%s%s" % (unbond_coins, AASTRA_DENOM),
        granter_address,
        GENERATE_ONLY,
    )
    with open(generated_unbond_txt, "w") as opened_file:
        json.dump(generated_unbond_msg, opened_file)
    rsp = exec_tx_by_grantee(astra_temp, generated_unbond_txt, grantee_address)
    wait_for_block_time(
        astra_temp.cosmos_cli(0),
        isoparse(parse_events(rsp["logs"])["unbond"]
                 ["completion_time"]) + timedelta(seconds=1),
    )

    assert query_delegation_amount(astra_temp, granter_address, validator1_address) == {
        "denom": AASTRA_DENOM,
        "amount": str(
            validator1_initial_deligation_amount + delegate_coins - unbond_coins
        ),
    }

    # test execute redelegate
    grant_authorization(
        astra_temp,
        grantee_address,
        AUTHORIZATION_REDELEGATE,
        granter_address,
        spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
        allowed_validators=validator2_address,
    )
    generated_redelegate_txt = tmp_path / "generated_redelegate.txt"
    generated_redelegate_msg = redelegate_amount(
        astra_temp,
        validator1_address,
        validator2_address,
        "%s%s" % (redelegate_coins, AASTRA_DENOM),
        granter_address,
        GENERATE_ONLY,
    )
    with open(generated_redelegate_txt, "w") as opened_file:
        json.dump(generated_redelegate_msg, opened_file)
    exec_tx_by_grantee(astra_temp, generated_redelegate_txt, grantee_address)

    assert query_delegation_amount(
        astra_temp, granter_address, validator1_address) == {
               "denom": AASTRA_DENOM,
               "amount": str(
                   validator1_initial_deligation_amount +
                   delegate_coins - unbond_coins - redelegate_coins
               ),
           }
    assert query_delegation_amount(astra_temp, granter_address, validator2_address) == {
        "denom": AASTRA_DENOM,
        "amount": str(validator2_initial_deligation_amount + redelegate_coins),
    }

    # teardown
    rsp = revoke_authorization(
        astra_temp, grantee_address, DELEGATE_MSG_TYPE_URL, granter_address
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    rsp = revoke_authorization(
        astra_temp, grantee_address, UNBOND_MSG_TYPE_URL, granter_address
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    rsp = revoke_authorization(
        astra_temp, grantee_address, REDELEGATE_MSG_TYPE_URL, granter_address
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    wait_for_block(astra_temp.cosmos_cli(0), 2)
    assert (
            len(query_command(astra_temp, AUTHZ, GRANTS,
                              granter_address, grantee_address)["grants"]) == 0
    )
