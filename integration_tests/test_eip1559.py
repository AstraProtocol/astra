from integration_tests.utils import ADDRS, send_transaction, KEYS, w3_wait_for_block, get_w3, get_supply


def adjust_base_fee(parent_fee, gas_limit, gas_used):
    "spec: https://eips.ethereum.org/EIPS/eip-1559#specification"
    change_denominator = 8
    elasticity_multiplier = 2
    gas_target = gas_limit // elasticity_multiplier

    delta = parent_fee * (gas_target - gas_used) // gas_target // change_denominator
    return parent_fee - delta


def test_dynamic_fee_tx(cluster):
    """
    test basic eip-1559 tx works:
    - tx fee calculation is compliant to go-ethereum
    - base fee adjustment is compliant to go-ethereum
    """
    w3 = get_w3()
    print("current block", w3.eth.get_block_number())
    amount = 10000
    validator1_address = cluster.address("validator", i=0)
    validator2_address = cluster.address("validator", i=1)

    total_balance_before = cluster.distribution_reward(validator1_address) + cluster.distribution_reward(validator2_address)
    before = w3.eth.get_balance(ADDRS["team"])
    tip_price = 1
    max_price = 1000000000000 + tip_price
    tx = {
        "to": "0x0000000000000000000000000000000000000000",
        "value": amount,
        "gas": 21000,
        "maxFeePerGas": max_price,
        "maxPriorityFeePerGas": tip_price,
    }

    total_supply = int(get_supply(cluster)["supply"][0]["amount"])
    tx_receipt = send_transaction(w3, tx, KEYS["team"])
    assert tx_receipt.status == 1
    blk = w3.eth.get_block(tx_receipt.blockNumber)
    assert tx_receipt.effectiveGasPrice == blk.baseFeePerGas + tip_price

    fee_expected = tx_receipt.gasUsed * tx_receipt.effectiveGasPrice
    print("fee_expected", fee_expected)
    after = w3.eth.get_balance(ADDRS["team"])
    fee_deducted = before - after - amount
    assert fee_deducted == fee_expected

    assert blk.gasUsed == tx_receipt.gasUsed  # we are the only tx in the block

    # check the next block's base fee is adjusted accordingly
    w3_wait_for_block(w3, tx_receipt.blockNumber + 1)
    total_supply_after = int(get_supply(cluster)["supply"][0]["amount"])
    print("supply change", total_supply_after - total_supply)
    total_balance_after = cluster.distribution_reward(validator1_address) + cluster.distribution_reward(validator2_address)
    balance_validator_change = int(total_balance_after) - int(total_balance_before)
    print("balance validator change ", balance_validator_change)

    print("diff", total_supply_after - total_supply - balance_validator_change)
    next_base_price = w3.eth.get_block(tx_receipt.blockNumber + 1).baseFeePerGas

    assert next_base_price == adjust_base_fee(
        blk.baseFeePerGas, blk.gasLimit, blk.gasUsed
    )


def test_base_fee_adjustment(cluster):
    """
    verify base fee adjustment of three continuous empty blocks
    """
    w3 = get_w3()
    begin = w3.eth.block_number
    w3_wait_for_block(w3, begin + 3)

    blk = w3.eth.get_block(begin)
    parent_fee = blk.baseFeePerGas

    for i in range(3):
        fee = w3.eth.get_block(begin + 1 + i).baseFeePerGas
        assert fee == adjust_base_fee(parent_fee, blk.gasLimit, 0)
        parent_fee = fee
