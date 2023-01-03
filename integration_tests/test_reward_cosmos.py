import random

import pytest
from pathlib import Path
from .network import setup_astra
from .test_reward_cosmos_util import block_provisions, num_tests, next_inflation_rate, mult_decimals, \
    decimal_equal, add_decimals, sub_decimals, approximate_equal, get_astra_foundation_address, \
    get_inflation_distribution, mult, expected_next_block_provision, decimal_int_equal, round_floor, round_ceiling
from .utils import DEFAULT_BASE_PORT, \
    wait_for_new_epochs, \
    wait_for_new_inflation_periods, parse_int, wait_for_new_blocks

pytestmark = pytest.mark.reward


@pytest.fixture(scope="module")
def astra(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    cfg = Path(__file__).parent / "configs/reward.yaml"
    yield from setup_astra(path, DEFAULT_BASE_PORT, cfg)


@pytest.mark.skip
def test_inflation_parameters(astra):
    inflation_params = astra.cosmos_cli(0).get_mint_params()
    print("inflation_params:", inflation_params)


@pytest.mark.skip
def test_correct_next_block_provision(astra):
    cli = astra.cosmos_cli(0)
    minter_params = cli.get_mint_params()
    inflation_parameters = minter_params["inflation_parameters"]
    print("num_tests:", num_tests)
    for i in range(0, num_tests):
        old_block = cli.block_height()

        old_supply = cli.get_circulating_supply()
        old_inflation_rate = cli.get_inflation_rate()
        bonded_ratio = cli.get_bonded_ratio()

        expected_block_provision = expected_next_block_provision(old_inflation_rate, inflation_parameters, bonded_ratio,
                                                                 old_supply)

        wait_for_new_blocks(cli, 1)
        # must ensure new_block = old_block + 1 for the test to work
        new_block = cli.block_height()
        if new_block - old_block != 1:
            i -= 1
            continue

        new_block_provision = cli.block_provisions()

        print(i, expected_block_provision, new_block_provision)

        assert decimal_equal(new_block_provision, expected_block_provision)

        print("SUCCESS WITH i =", i, "\n\n")


@pytest.mark.skip
def test_correct_next_inflation_rate(astra):
    cli = astra.cosmos_cli(0)
    minter_params = cli.get_mint_params()
    inflation_parameters = minter_params["inflation_parameters"]
    print("num_tests:", num_tests)
    for i in range(0, num_tests):
        old_block = cli.block_height()

        old_inflation_rate = cli.get_inflation_rate()
        bonded_ratio = cli.get_bonded_ratio()

        expected_next_inflation_rate = next_inflation_rate(old_inflation_rate, inflation_parameters, bonded_ratio)

        wait_for_new_blocks(cli, 1)
        # must ensure new_block = old_block + 1 for the test to work
        new_block = cli.block_height()
        if new_block - old_block != 1:
            i -= 1
            continue

        new_inflation_rate = cli.get_inflation_rate()

        # print(i, old_inflation_rate, bonded_ratio, expected_next_inflation_rate, new_inflation_rate, new_block)

        assert decimal_equal(new_inflation_rate, expected_next_inflation_rate)

        print("SUCCESS WITH i =", i, "\n\n")


def test_correct_supplies(astra):
    # This test also covers the following tests:
    #   - test_correct_next_block_provision
    #   - test_correct_next_inflation_rate
    cli = astra.cosmos_cli(0)
    minter_params = cli.get_mint_params()
    inflation_parameters = minter_params["inflation_parameters"]
    print("num_tests:", num_tests)
    for i in range(0, num_tests):
        old_block = cli.block_height()

        old_supply = cli.get_circulating_supply()
        old_inflation_rate = cli.get_inflation_rate()
        bonded_ratio = cli.get_bonded_ratio()

        expected_block_provision = expected_next_block_provision(old_inflation_rate, inflation_parameters, bonded_ratio,
                                                                 old_supply)

        wait_for_new_blocks(cli, 1)
        # must ensure new_block = old_block + 1 for the test to work
        new_block = cli.block_height()
        if new_block - old_block != 1:
            i -= 1
            continue

        new_supply = cli.get_circulating_supply()
        assert decimal_equal(new_supply, add_decimals(old_supply, expected_block_provision))

        print("SUCCESS WITH i =", i, "\n\n")


def test_should_distribute_rewards_to_validators_when_new_epochs_arrive(astra):
    cli0 = astra.cosmos_cli(0)
    cli1 = astra.cosmos_cli(1)

    # query the community tax
    params = cli0.distribution_params()
    community_tax = float(params["community_tax"])

    minter_params = cli0.get_mint_params()
    inflation_parameters = minter_params["inflation_parameters"]

    foundation_address = get_astra_foundation_address(minter_params)
    inflation_distribution = get_inflation_distribution(minter_params)

    # get validators' information
    validator1_address = cli0.address("validator")
    validator2_address = cli1.address("validator")
    validator1_operator_address = cli0.address("validator", bech="val")
    validator2_operator_address = cli1.address("validator", bech="val")

    print("num_tests:", num_tests)
    for i in range(0, num_tests):
        wait_for_new_blocks(cli0, 1)
        # retrieve the current "old" block
        old_block = cli0.block_height()
        old_supply = cli0.get_circulating_supply()
        old_inflation_rate = cli0.get_inflation_rate()
        bonded_ratio = cli0.get_bonded_ratio()

        # calculate the old reward amounts
        old_reward_amount = cli0.distribution_reward(validator1_address)
        old_reward_amount2 = cli0.distribution_reward(validator2_address)
        old_commission_amount = cli0.distribution_commission(validator1_operator_address)
        old_commission_amount2 = cli0.distribution_commission(
            validator2_operator_address,
        )
        old_total_amount = old_reward_amount + old_commission_amount
        old_total_amount2 = old_reward_amount2 + old_commission_amount2

        old_community_amount = cli0.distribution_community()
        old_foundation_amount = cli0.balance(foundation_address)
        # print("[old]", old_foundation_amount, old_community_amount, old_total_amount, old_total_amount2)

        # get the next block provision
        block_provision = expected_next_block_provision(old_inflation_rate, inflation_parameters, bonded_ratio,
                                                        old_supply)

        # wait for a new block
        wait_for_new_blocks(cli0, 1)
        # must ensure new_block = old_block + 1 for the test to work
        new_block = cli0.block_height()
        if new_block - old_block != 1:
            i -= 1
            continue

        # calculate the current reward amounts
        reward_amount = cli0.distribution_reward(validator1_address)
        reward_amount2 = cli0.distribution_reward(validator2_address)
        commission_amount = cli0.distribution_commission(validator1_operator_address)
        commission_amount2 = cli0.distribution_commission(
            validator2_operator_address,
        )
        total_amount = reward_amount + commission_amount
        total_amount2 = reward_amount2 + commission_amount2

        # retrieve the new balance of the community pool
        community_amount = cli0.distribution_community()
        # retrieve the new balance of the foundation
        foundation_amount = cli0.balance(foundation_address)
        # print("[new]", foundation_amount, community_amount, total_amount, total_amount2)

        reward_increase = total_amount + total_amount2 - old_total_amount - old_total_amount2
        community_increase = community_amount - old_community_amount
        foundation_increase = foundation_amount - old_foundation_amount
        # print("[increase]", foundation_increase, community_increase, reward_increase,
        #       foundation_increase + community_increase + reward_increase,
        #       block_provision)

        # block_provision should be distributed to all validators + community_tax
        expected_reward_amount_for_validators_increase = mult_decimals(
            block_provision,
            inflation_distribution[0],
        )
        expected_foundation_amount_increase = round_floor(mult_decimals(block_provision,
                                                                        inflation_distribution[1]))
        expected_community_amount_increase = round_ceiling(
            mult_decimals(
                block_provision,
                inflation_distribution[2])) + round_floor(
            mult_decimals(
                expected_reward_amount_for_validators_increase,
                community_tax,
            )
        )
        expected_reward_amount_for_validators_increase = round_floor(mult_decimals(
            expected_reward_amount_for_validators_increase,
            (1 - community_tax)))

        assert decimal_int_equal(foundation_increase, expected_foundation_amount_increase)
        assert decimal_int_equal(community_increase, expected_community_amount_increase)
        assert decimal_int_equal(reward_increase, expected_reward_amount_for_validators_increase)
        assert decimal_int_equal(reward_increase + community_increase + foundation_increase, block_provision)
        print("SUCCEEDED WITH i =", i, "\n\n")

