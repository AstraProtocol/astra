import json
import time
from datetime import timedelta
from pathlib import Path

import pytest
from dateutil.parser import isoparse
from pystarport.ports import rpc_port
from integration_tests.network import setup_astra

from .utils import (
    DEFAULT_BASE_PORT,
    cluster_fixture,
    parse_events,
    wait_for_block,
    wait_for_block_time,
    wait_for_new_blocks,
    wait_for_port,
)

pytestmark = pytest.mark.stake


@pytest.fixture(scope="module")
def astra(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    cfg = Path(__file__).parent / "configs/staking.yaml"
    yield from setup_astra(path, DEFAULT_BASE_PORT, cfg)

def test_staking_delegate(astra):
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