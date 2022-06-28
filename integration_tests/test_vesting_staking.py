from pathlib import Path
from time import sleep

import pytest

from integration_tests.network import setup_astra
from integration_tests.utils import DEFAULT_BASE_PORT

pytestmark = pytest.mark.vesting


@pytest.fixture(scope="module")
def astra(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    cfg = Path(__file__).parent / "configs/staking.yaml"
    yield from setup_astra(path, DEFAULT_BASE_PORT, cfg)


# one more test for the vesting account bug
# that one can delegate twice with fee + redelegate
def test_staking_vesting_redelegate(astra):
    sleep(1)
    community_addr = astra.cosmos_cli(0).address("community")
    reserve_addr = astra.cosmos_cli(0).address("team")
    # for the fee payment
    astra.cosmos_cli(0).transfer(community_addr, reserve_addr, "10000aastra")

    signer1_address = astra.cosmos_cli(0).address("team")
    validators = astra.cosmos_cli(0).validators()
    validator1_operator_address = validators[0]["operator_address"]
    validator2_operator_address = validators[1]["operator_address"]
    staking_validator1 = astra.cosmos_cli(0).validator(validator1_operator_address)
    assert validator1_operator_address == staking_validator1["operator_address"]
    staking_validator2 = astra.cosmos_cli(1).validator(validator2_operator_address)
    assert validator2_operator_address == staking_validator2["operator_address"]
    old_bonded = astra.cosmos_cli(0).staking_pool()
    rsp = astra.cosmos_cli(0).delegate_amount(
        validator1_operator_address,
        "2009999498aastra",
        signer1_address,
        "0.025aastra",
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    assert astra.cosmos_cli(0).staking_pool() == old_bonded + 2009999498
    rsp = astra.cosmos_cli(0).delegate_amount(
        validator2_operator_address, "1aastra", signer1_address, "0.025aastra"
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    assert astra.cosmos_cli(0).staking_pool() == old_bonded + 2009999499
    # delegation_info = cluster.get_delegated_amount(signer1_address)
    # old_output = delegation_info["delegation_responses"][0]["balance"]["amount"]
    astra.cosmos_cli(0).redelegate_amount(
        validator1_operator_address,
        validator2_operator_address,
        "2aastra",
        signer1_address,
    )
    # delegation_info = cluster.get_delegated_amount(signer1_address)
    # output = delegation_info["delegation_responses"][0]["balance"]["amount"]
    # assert int(old_output) + 2 == int(output)
    assert astra.cosmos_cli(0).staking_pool() == old_bonded + 2009999499
    account = astra.cosmos_cli(0).account(signer1_address)
    assert account["@type"] == "/cosmos.vesting.v1beta1.DelayedVestingAccount"
    assert account["base_vesting_account"]["original_vesting"] == [
        {"denom": "aastra", "amount": "200000000000000000000"}
    ]