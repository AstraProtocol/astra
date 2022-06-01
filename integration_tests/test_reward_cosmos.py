import pytest
from pathlib import Path
from .network import setup_astra
from .utils import wait_for_block, wait_for_new_blocks, ADDRS, DEFAULT_BASE_PORT

pytestmark = pytest.mark.reward


@pytest.fixture(scope="module")
def astra(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    cfg = Path(__file__).parent / "configs/reward.yaml"
    yield from setup_astra(path, DEFAULT_BASE_PORT, cfg)


def test_reward_block_proposal(astra):
    # starts with astra

    validator1_address = astra.cosmos_cli(0).address("validator")
    validator2_address = astra.cosmos_cli(1).address("validator")
    total_supply_before = int(astra.cosmos_cli(0).total_supply()["supply"][0]["amount"])
    print("total_supply_before", total_supply_before)
    # starts with astraval
    validator1_operator_address = astra.cosmos_cli(0).address("validator", bech="val")
    validator2_operator_address = astra.cosmos_cli(1).address("validator", bech="val")
    # wait for initial reward processed, so that distribution values can be read
    old_commission_amount = astra.cosmos_cli(0).distribution_commission(validator1_operator_address)
    old_commission_amount2 = astra.cosmos_cli(0).distribution_commission(
        validator2_operator_address,
    )
    old_community_amount = astra.cosmos_cli(0).distribution_community()
    old_reward_amount = astra.cosmos_cli(0).distribution_reward(validator1_address)
    old_reward_amount2 = astra.cosmos_cli(0).distribution_reward(validator2_address)

    print("old reward", old_reward_amount, old_reward_amount2)

    # wait for fee reward receive
    wait_for_new_blocks(astra.cosmos_cli(0), 1)
    total_supply_after = int(astra.cosmos_cli(0).total_supply()["supply"][0]["amount"])
    commission_amount = astra.cosmos_cli(0).distribution_commission(validator1_operator_address)
    commission_amount2 = astra.cosmos_cli(0).distribution_commission(validator2_operator_address)
    commission_amount_diff = (commission_amount - old_commission_amount) + (
            commission_amount2 - old_commission_amount2
    )
    community_amount = astra.cosmos_cli(0).distribution_community()
    community_amount_diff = community_amount - old_community_amount
    reward_amount = astra.cosmos_cli(0).distribution_reward(validator1_address)
    reward_amount2 = astra.cosmos_cli(0).distribution_reward(validator2_address)
    print("new reward", reward_amount, reward_amount2)
    reward_amount_diff = (reward_amount - old_reward_amount) + (
            reward_amount2 - old_reward_amount2
    )
    supply_diff = total_supply_after - total_supply_before
    print("supply_diff", supply_diff)
    print(commission_amount_diff, community_amount_diff, reward_amount_diff)
    total_diff = commission_amount_diff + community_amount_diff + reward_amount_diff
    print(total_diff)
    assert community_amount_diff <= supply_diff * 2.0 / 100 + 1
    assert reward_amount_diff + commission_amount_diff <= supply_diff * 98.0 / 100 + 100

    total_validator1_change = reward_amount + commission_amount - old_reward_amount - old_commission_amount
    total_validator2_change = reward_amount2 + commission_amount2 - old_reward_amount2 - old_commission_amount2
    print(total_validator1_change, total_validator2_change)
    print("old", old_reward_amount + old_commission_amount, old_reward_amount2 + old_commission_amount2)
    assert total_diff <= supply_diff + 10


def test_reward_when_execute_tx(astra):
    # starts with astra
    signer1_address = astra.cosmos_cli(0).address("signer1")
    signer2_address = astra.cosmos_cli(0).address("signer2")
    validator1_address = astra.cosmos_cli(0).address("validator")
    validator2_address = astra.cosmos_cli(1).address("validator")
    # starts with astraval
    validator1_operator_address = astra.cosmos_cli(0).address("validator", bech="val")
    validator2_operator_address = astra.cosmos_cli(1).address("validator", bech="val")
    signer1_old_balance = astra.cosmos_cli(0).balance(signer1_address)
    amount_to_send = 2 * 10 ** 18
    fees = 100_000
    # wait for initial reward processed, so that distribution values can be read
    wait_for_block(astra.cosmos_cli(0), 2)
    old_commission_amount = astra.cosmos_cli(0).distribution_commission(validator1_operator_address)
    old_commission_amount2 = astra.cosmos_cli(0).distribution_commission(
        validator2_operator_address,
    )
    old_community_amount = astra.cosmos_cli(0).distribution_community()
    old_reward_amount = astra.cosmos_cli(0).distribution_reward(validator1_address)
    old_reward_amount2 = astra.cosmos_cli(0).distribution_reward(validator2_address)
    # transfer with fees
    astra.cosmos_cli(0).transfer(
        signer1_address,
        signer2_address,
        f"{amount_to_send}aastra",
        fees=f"{fees}aastra",
    )
    # wait for fee reward receive
    wait_for_new_blocks(astra.cosmos_cli(0), 1)
    signer1_balance = astra.cosmos_cli(0).balance(signer1_address)
    assert signer1_balance + fees + amount_to_send == signer1_old_balance
    commission_amount = astra.cosmos_cli(0).distribution_commission(validator1_operator_address)
    commission_amount2 = astra.cosmos_cli(0).distribution_commission(validator2_operator_address)
    commission_amount_diff = (commission_amount - old_commission_amount) + (
            commission_amount2 - old_commission_amount2
    )
    community_amount = astra.cosmos_cli(0).distribution_community()
    community_amount_diff = community_amount - old_community_amount
    reward_amount = astra.cosmos_cli(0).distribution_reward(validator1_address)
    reward_amount2 = astra.cosmos_cli(0).distribution_reward(validator2_address)
    reward_amount_diff = (reward_amount - old_reward_amount) + (
            reward_amount2 - old_reward_amount2
    )
    total_diff = commission_amount_diff + community_amount_diff + reward_amount_diff
    minted_value = total_diff - fees
    # these values are generated by minting
    # if there is system-overload, minted_value can be larger than expected
    # fee is computed at EndBlock, AllocateTokens
    # commission = proposerReward * proposerCommissionRate
    # communityFunding = feesCollectedDec * communityTax
    # poolReceived = feesCollectedDec - proposerReward - communityFunding
    assert 77000.0 <= minted_value <= 1245838471873.0142
