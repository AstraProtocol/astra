import json
from time import sleep
from datetime import timedelta
from pathlib import Path

import pytest
from dateutil.parser import isoparse
from pystarport.ports import rpc_port
from integration_tests.network import setup_astra

from .utils import (
    DEFAULT_BASE_PORT,
    parse_events,
    wait_for_block,
    wait_for_block_time,
    wait_for_port,
)

pytestmark = pytest.mark.staking


@pytest.fixture(scope="module")
def astra(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    cfg = Path(__file__).parent / "configs/staking.yaml"
    yield from setup_astra(path, DEFAULT_BASE_PORT, cfg)


def test_staking_delegate(astra):
    sleep(1)
    signer1_address = astra.cosmos_cli(0).address("signer1")
    validators = astra.cosmos_cli(0).validators()
    validator1_operator_address = validators[0]["operator_address"]
    validator2_operator_address = validators[1]["operator_address"]
    staking_validator1 = astra.cosmos_cli(0).validator(validator1_operator_address)
    assert validator1_operator_address == staking_validator1["operator_address"]
    staking_validator2 = astra.cosmos_cli(1).validator(validator2_operator_address)
    assert validator2_operator_address == staking_validator2["operator_address"]
    old_amount = astra.cosmos_cli(0).balance(signer1_address)
    old_bonded = astra.cosmos_cli(0).staking_pool()
    rsp = astra.cosmos_cli(0).delegate_amount(
        validator1_operator_address, "2aastra", signer1_address
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    assert astra.cosmos_cli(0).staking_pool() == old_bonded + 2
    new_amount = astra.cosmos_cli(0).balance(signer1_address)
    assert old_amount == new_amount + 2


def test_staking_unbond(astra):
    sleep(1)
    signer1_address = astra.cosmos_cli(0).address("signer1")
    validators = astra.cosmos_cli(0).validators()
    validator1_operator_address = validators[0]["operator_address"]
    validator2_operator_address = validators[1]["operator_address"]
    staking_validator1 = astra.cosmos_cli(0).validator(validator1_operator_address)
    assert validator1_operator_address == staking_validator1["operator_address"]
    staking_validator2 = astra.cosmos_cli(1).validator(validator2_operator_address)
    assert validator2_operator_address == staking_validator2["operator_address"]
    old_amount = astra.cosmos_cli(0).balance(signer1_address)
    old_bonded = astra.cosmos_cli(0).staking_pool()
    rsp = astra.cosmos_cli(0).delegate_amount(
        validator1_operator_address, "3aastra", signer1_address
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    rsp = astra.cosmos_cli(0).delegate_amount(
        validator2_operator_address, "4aastra", signer1_address
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    assert astra.cosmos_cli(0).staking_pool() == old_bonded + 7
    assert astra.cosmos_cli(0).balance(signer1_address) == old_amount - 7

    old_unbonded = astra.cosmos_cli(0).staking_pool(bonded=False)
    rsp = astra.cosmos_cli(0).unbond_amount(
        validator2_operator_address, "2aastra", signer1_address
    )
    assert rsp["code"] == 0, rsp
    assert astra.cosmos_cli(0).staking_pool(bonded=False) == old_unbonded + 2

    wait_for_block_time(
        astra.cosmos_cli(0),
        isoparse(parse_events(rsp["logs"])["unbond"]["completion_time"])
        + timedelta(seconds=1),
    )

    assert astra.cosmos_cli(0).balance(signer1_address) == old_amount - 5


def test_staking_redelegate(astra):
    sleep(1)
    signer1_address = astra.cosmos_cli(0).address("signer1")
    validators = astra.cosmos_cli(0).validators()
    validator1_operator_address = validators[0]["operator_address"]
    validator2_operator_address = validators[1]["operator_address"]
    staking_validator1 = astra.cosmos_cli(0).validator(validator1_operator_address)
    assert validator1_operator_address == staking_validator1["operator_address"]
    staking_validator2 = astra.cosmos_cli(1).validator(validator2_operator_address)
    assert validator2_operator_address == staking_validator2["operator_address"]
    rsp = astra.cosmos_cli(0).delegate_amount(
        validator1_operator_address, "3aastra", signer1_address
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    rsp = astra.cosmos_cli(0).delegate_amount(
        validator2_operator_address, "4aastra", signer1_address
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    delegation_info = astra.cosmos_cli(0).get_delegated_amount(signer1_address)
    old_output = delegation_info["delegation_responses"][0]["balance"]["amount"]
    cli = astra.cosmos_cli(0)
    rsp = json.loads(
        cli.raw(
            "tx",
            "staking",
            "redelegate",
            validator2_operator_address,
            validator1_operator_address,
            "2aastra",
            "-y",
            "--gas",
            "300000",
            home=cli.data_dir,
            from_=signer1_address,
            keyring_backend="test",
            chain_id=cli.chain_id,
            node=cli.node_rpc,
        )
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    delegation_info = astra.cosmos_cli(0).get_delegated_amount(signer1_address)
    output = delegation_info["delegation_responses"][0]["balance"]["amount"]
    assert int(old_output) + 2 == int(output)


def test_join_validator(astra):
    sleep(1)
    i = astra.cosmos_cli(0).create_node(moniker="new joined")
    addr = astra.cosmos_cli(i).address("validator")
    # transfer 1astra from community account
    assert astra.cosmos_cli(0).transfer(astra.cosmos_cli(0).address("community"), addr, "1astra")["code"] == 0
    assert astra.cosmos_cli(0).balance(addr) == 10 ** 18

    # start the node
    print("START NEW NODE...")
    astra.cosmos_cli(0).start_node(i)
    wait_for_port(rpc_port(astra.cosmos_cli(0).base_port(i)))

    count1 = len(astra.cosmos_cli(0).validators())

    # wait for the new node to sync
    wait_for_block(astra.cosmos_cli(i), astra.cosmos_cli(0).block_height())

    # wait for the new node to sync
    wait_for_block(astra.cosmos_cli(i), astra.cosmos_cli(0).block_height())
    # create validator tx
    assert astra.cosmos_cli(i).create_validator("1astra")["code"] == 0
    sleep(1)

    count2 = len(astra.cosmos_cli(0).validators())
    assert count2 == count1 + 1, "new validator should joined successfully"

    val_addr = astra.cosmos_cli(i).address("validator", bech="val")
    val = astra.cosmos_cli(0).validator(val_addr)
    assert not val["jailed"]
    assert val["status"] == "BOND_STATUS_BONDED"
    assert val["tokens"] == str(10 ** 18)
    assert val["description"]["moniker"] == "new joined"
    assert val["commission"]["commission_rates"] == {
        "rate": "0.100000000000000000",
        "max_rate": "0.200000000000000000",
        "max_change_rate": "0.010000000000000000",
    }
    assert (
            astra.cosmos_cli(i).edit_validator(commission_rate="0.2")["code"] == 12
    ), "commission cannot be changed more than once in 24h"
    assert astra.cosmos_cli(i).edit_validator(moniker="awesome node")["code"] == 0
    assert astra.cosmos_cli(0).validator(val_addr)["description"]["moniker"] == "awesome node"


def test_min_self_delegation(astra):
    """
    - validator unbond min_self_delegation
    - check not in validator set anymore
    """
    sleep(1)
    assert len(astra.cosmos_cli(0).validators()) == 4, "wrong validator set"

    oper_addr = astra.cosmos_cli(2).address("validator", bech="val")
    acct_addr = astra.cosmos_cli(2).address("validator")
    rsp = astra.cosmos_cli(2).unbond_amount(oper_addr, "90000000aastra", acct_addr)
    assert rsp["code"] == 0, rsp["raw_log"]

    def find_validator():
        return next(
            iter(
                val
                for val in astra.cosmos_cli(0).validators()
                if val["operator_address"] == oper_addr
            )
        )

    assert (
            find_validator()["status"] == "BOND_STATUS_UNBONDING"
    ), "validator get removed"
