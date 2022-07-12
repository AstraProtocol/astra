import random

import pytest
from pathlib import Path
from .network import setup_astra
from .utils import DEFAULT_BASE_PORT, \
    wait_for_new_epochs, \
    wait_for_new_inflation_periods

pytestmark = pytest.mark.reward


@pytest.fixture(scope="module")
def astra(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    cfg = Path(__file__).parent / "configs/reward.yaml"
    yield from setup_astra(path, DEFAULT_BASE_PORT, cfg)


def approximate_equal(a, b, diff_rate=1e-9):
    if b == 0:
        return a == b
    return abs(1-float(a)/float(b)) <= diff_rate


def test_period_correctly_increases(astra):
    old_inflation_period_info = astra.cosmos_cli(0).get_inflation_period()
    print("old_period_info:", old_inflation_period_info)

    old_period = int(old_inflation_period_info["period"])

    # wait for some periods
    num_periods = random.randint(1, 10)
    print("num_periods:", num_periods)
    wait_for_new_inflation_periods(astra.cosmos_cli(0), n=num_periods)

    inflation_period_info = astra.cosmos_cli(0).get_inflation_period()
    print("new_period_info:", inflation_period_info)
    new_period = int(inflation_period_info["period"])

    assert new_period == old_period + num_periods


def test_correct_mint_provisions(astra):
    inflation_params = astra.cosmos_cli(0).get_inflation_params()
    print("inflation_params:", inflation_params)

    max_staking_rewards = float(inflation_params["inflation_parameters"]["max_staking_rewards"])
    r = float(inflation_params["inflation_parameters"]["r"])

    num_tests = random.randint(1, 10)
    print("num_tests:", num_tests)
    for i in range(0, num_tests):
        inflation_period_info = astra.cosmos_cli(0).get_inflation_period()
        old_period = int(inflation_period_info["period"])

        num_periods = random.randint(1, 3)
        wait_for_new_inflation_periods(astra.cosmos_cli(0), n=num_periods)

        inflation_period_info = astra.cosmos_cli(0).get_inflation_period()
        period = int(inflation_period_info["period"])
        epochs_per_period = int(inflation_period_info["epochs_per_period"])
        assert period == old_period + num_periods

        mint_provision = astra.cosmos_cli(0).get_epoch_mint_provision()

        expected_mint_provision = int(max_staking_rewards*r*(1-r)**period) / epochs_per_period
        print("old_period, period, num_periods, mint_provision, expected_provision:", old_period, period, num_periods,
              mint_provision, expected_mint_provision)

        # we temporarily use approximation to overcome the float precision of python
        assert approximate_equal(mint_provision, expected_mint_provision)


def test_correct_supplies(astra):
    num_epochs = random.randint(1, 10)
    print("num_epochs:", num_epochs)
    epoch_identifier = astra.cosmos_cli(0).get_inflation_epoch_identifier()
    for i in range(0, num_epochs):
        old_circulating_supply = astra.cosmos_cli(0).get_circulating_supply()
        old_total_supply = int(float(astra.cosmos_cli(0).total_supply()["supply"][0]["amount"]))

        current_mint_provision = astra.cosmos_cli(0).get_epoch_mint_provision()

        # sleep for an epoch
        wait_for_new_epochs(astra.cosmos_cli(0), epoch_identifier=epoch_identifier)

        # calculate new supplies
        new_circulating_supply = astra.cosmos_cli(0).get_circulating_supply()
        new_total_supply = int(float(astra.cosmos_cli(0).total_supply()["supply"][0]["amount"]))

        cir_supply_diff = new_circulating_supply - old_circulating_supply
        total_supply_diff = new_total_supply - old_total_supply

        # total supply increase should be greater than or equal to the circulating supply increase.
        assert total_supply_diff >= cir_supply_diff

        # circulating supply increase should be approximately equal to the current mint provision.
        assert approximate_equal(cir_supply_diff, current_mint_provision)
        print("SUCCESS WITH i =", i, "\n\n")


def test_should_distribute_rewards_to_validators_when_new_epochs_arrive(astra):
    # query the community tax
    params = astra.cosmos_cli(0).distribution_params()
    community_tax = float(params["community_tax"])
    # print("community_tax:", community_tax)

    # get validators' information
    validator1_address = astra.cosmos_cli(0).address("validator")
    validator2_address = astra.cosmos_cli(1).address("validator")
    validator1_operator_address = astra.cosmos_cli(0).address("validator", bech="val")
    validator2_operator_address = astra.cosmos_cli(1).address("validator", bech="val")

    num_tests = random.randint(1, 10)
    print("num_tests:", num_tests)
    for i in range(0, num_tests):
        # calculate the old reward amounts
        old_reward_amount = int(astra.cosmos_cli(0).distribution_reward(validator1_address))
        old_reward_amount2 = int(astra.cosmos_cli(0).distribution_reward(validator2_address))
        old_commission_amount = int(astra.cosmos_cli(0).distribution_commission(validator1_operator_address))
        old_commission_amount2 = int(astra.cosmos_cli(0).distribution_commission(
            validator2_operator_address,
        ))
        old_total_amount = old_reward_amount + old_commission_amount
        old_total_amount2 = old_reward_amount2 + old_commission_amount2
        old_community_amount = int(astra.cosmos_cli(0).distribution_community())

        # get the current mint-provision
        current_mint_provision = astra.cosmos_cli(0).get_epoch_mint_provision()

        # wait for a new epoch
        wait_for_new_epochs(astra.cosmos_cli(0),
                            epoch_identifier=astra.cosmos_cli(0).get_inflation_epoch_identifier())

        # calculate the current reward amounts
        reward_amount = int(astra.cosmos_cli(0).distribution_reward(validator1_address))
        reward_amount2 = int(astra.cosmos_cli(0).distribution_reward(validator2_address))
        commission_amount = int(astra.cosmos_cli(0).distribution_commission(validator1_operator_address))
        commission_amount2 = int(astra.cosmos_cli(0).distribution_commission(
            validator2_operator_address,
        ))
        total_amount = reward_amount + commission_amount
        total_amount2 = reward_amount2 + commission_amount2
        community_amount = int(astra.cosmos_cli(0).distribution_community())

        reward_increase = total_amount + total_amount2 - old_total_amount - old_total_amount2
        community_increase = community_amount - old_community_amount

        # mint_provision should be distributed to all validators + community_tax
        expected_reward_amount_for_validators = int((1-community_tax) * current_mint_provision)
        expected_community_amount_increase = int(community_tax * current_mint_provision)

        assert approximate_equal(reward_increase, expected_reward_amount_for_validators, diff_rate=1e-12)
        assert approximate_equal(community_increase, expected_community_amount_increase, diff_rate=1e-12)
        print("SUCCESS WITH i =", i, "\n\n")


def test_should_distribute_fees_to_validators_when_execute_tx(astra):
    # query the community tax
    params = astra.cosmos_cli(0).distribution_params()
    community_tax = float(params["community_tax"])
    print("community_tax:", community_tax)

    # get addresses
    sender_address = astra.cosmos_cli(0).address("signer1")
    receiver_address = astra.cosmos_cli(0).address("signer2")

    # get validators' information
    validator1_address = astra.cosmos_cli(0).address("validator")
    validator2_address = astra.cosmos_cli(1).address("validator")
    validator1_operator_address = astra.cosmos_cli(0).address("validator", bech="val")
    validator2_operator_address = astra.cosmos_cli(1).address("validator", bech="val")

    # num_tests = random.randint(1, 10)
    num_tests = 5
    print("num_tests:", num_tests)
    for i in range(0, num_tests):
        # wait for the current epoch to be passed
        wait_for_new_epochs(astra.cosmos_cli(0),
                            epoch_identifier=astra.cosmos_cli(0).get_inflation_epoch_identifier())

        # calculate old balances
        old_sender_balance = int(astra.cosmos_cli(0).balance(sender_address))
        old_receiver_balance = int(astra.cosmos_cli(0).balance(receiver_address))

        amount_to_send = random.randint(1, 20) * 10 ** 15
        tx_fee = random.randint(10, 20) * 10 ** 14
        print("amount_to_send, tx_fee:", amount_to_send, tx_fee)

        # calculate the old reward amounts
        old_reward_amount = int(astra.cosmos_cli(0).distribution_reward(validator1_address))
        old_reward_amount2 = int(astra.cosmos_cli(0).distribution_reward(validator2_address))
        old_commission_amount = int(astra.cosmos_cli(0).distribution_commission(validator1_operator_address))
        old_commission_amount2 = int(astra.cosmos_cli(0).distribution_commission(
            validator2_operator_address,
        ))
        old_total_amount = old_reward_amount + old_commission_amount
        old_total_amount2 = old_reward_amount2 + old_commission_amount2
        old_community_amount = int(astra.cosmos_cli(0).distribution_community())

        # get the current mint-provision
        current_mint_provision = astra.cosmos_cli(0).get_epoch_mint_provision()

        # transfer with fees
        astra.cosmos_cli(0).transfer(
            sender_address,
            receiver_address,
            f"{amount_to_send}aastra",
            fees=f"{tx_fee}aastra",
        )

        # wait for a new epoch (an epoch has passed => new blocks has passed)
        wait_for_new_epochs(astra.cosmos_cli(0),
                            epoch_identifier=astra.cosmos_cli(0).get_inflation_epoch_identifier())

        # assert the balances
        sender_balance = int(astra.cosmos_cli(0).balance(sender_address))
        receiver_balance = int(astra.cosmos_cli(0).balance(receiver_address))
        assert receiver_balance == old_receiver_balance + amount_to_send
        assert sender_balance == old_sender_balance - amount_to_send - tx_fee

        # calculate the current reward amounts
        reward_amount = int(astra.cosmos_cli(0).distribution_reward(validator1_address))
        reward_amount2 = int(astra.cosmos_cli(0).distribution_reward(validator2_address))
        commission_amount = int(astra.cosmos_cli(0).distribution_commission(validator1_operator_address))
        commission_amount2 = int(astra.cosmos_cli(0).distribution_commission(
            validator2_operator_address,
        ))
        total_amount = reward_amount + commission_amount
        total_amount2 = reward_amount2 + commission_amount2
        community_amount = int(astra.cosmos_cli(0).distribution_community())

        reward_increase = total_amount + total_amount2 - old_total_amount - old_total_amount2
        community_increase = community_amount - old_community_amount

        # all rewards now: mint_provision + tx_fee
        all_rewards = current_mint_provision + tx_fee
        assert approximate_equal(all_rewards, reward_increase + community_increase, diff_rate=1e-12), \
            all_rewards-reward_increase-community_increase

        # all_rewards should be distributed to all validators + community_tax
        expected_reward_amount_for_validators = int((1-community_tax) * all_rewards)
        expected_community_amount_increase = int(community_tax * all_rewards)

        assert approximate_equal(reward_increase, expected_reward_amount_for_validators, diff_rate=1e-12)
        assert approximate_equal(community_increase, expected_community_amount_increase, diff_rate=1e-12)
        print("SUCCESS WITH i =", i, "\n\n")

