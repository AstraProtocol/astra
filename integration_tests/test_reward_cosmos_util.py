import os
from decimal import *

num_tests = min(1 + int(os.urandom(1)[0]), 20)


def get_astra_foundation_address(inflation_params):
    return inflation_params["foundation_address"]


def get_inflation_distribution(inflation_params):
    distribution = inflation_params["inflation_distribution"]
    return [
        Decimal(distribution["staking_rewards"]),
        Decimal(distribution["foundation"]),
        Decimal(distribution["community_pool"]),
    ]


def mult_decimals(a, b, prec=28):
    getcontext().prec = prec
    return Decimal(a) * Decimal(b)


def add_decimals(a, b, prec=28):
    getcontext().prec = prec
    return Decimal(a) + Decimal(b)


def sub_decimals(a, b, prec=28):
    getcontext().prec = prec
    return Decimal(a) - Decimal(b)


def mult(a, b, base=int(100)):
    b = int(b * base)
    return a * b // base


def approximate_equal(a, b, diff_rate=1e-18):
    if b == 0:
        assert a == b
    else:
        rate = abs(1 - float(a) / float(b))
        assert rate <= diff_rate

    return True


def decimal_equal(a, b, prec=15):
    getcontext().prec = prec
    res = Decimal(Decimal("1") - (Decimal(a) / Decimal(b))).copy_abs()
    return float(res) == 0


def decimal_int_equal(a, b, max_diff=10):
    res = Decimal(Decimal(a) - Decimal(b)).copy_abs()

    return int(res) <= int(max_diff)


def round_floor(a):
    return Decimal(a).to_integral(rounding=ROUND_FLOOR)


def round_ceiling(a):
    return Decimal(a).to_integral(rounding=ROUND_CEILING)


def block_provisions(annual_provisions, blocks_per_year):
    a = Decimal(annual_provisions) // Decimal(blocks_per_year)

    return int(a)


def next_inflation_rate(old_inflation_rate, inflation_parameters, bonded_ratio, prec=28):
    getcontext().prec = prec
    goal_bonded = Decimal(inflation_parameters["goal_bonded"])
    inflation_max = Decimal(inflation_parameters["inflation_max"])
    inflation_min = Decimal(inflation_parameters["inflation_min"])
    blocks_per_year = Decimal(inflation_parameters["blocks_per_year"])
    inflation_rate_change = Decimal(inflation_parameters["inflation_rate_change"])

    inflation_change = (Decimal("1") - Decimal(bonded_ratio) / goal_bonded) * inflation_rate_change
    inflation_change = inflation_change / blocks_per_year

    expected_inflation_rate = add_decimals(old_inflation_rate, inflation_change)
    if expected_inflation_rate > inflation_max:
        expected_inflation_rate = inflation_max
    if expected_inflation_rate < inflation_min:
        expected_inflation_rate = inflation_min

    return expected_inflation_rate


def expected_next_block_provision(old_inflation_rate, inflation_parameters, bonded_ratio, old_supply, prec=28):
    getcontext().prec = prec
    expected_next_inflation_rate = next_inflation_rate(old_inflation_rate, inflation_parameters, bonded_ratio, prec)
    annual_provision = mult_decimals(expected_next_inflation_rate, old_supply)

    return block_provisions(annual_provision, inflation_parameters["blocks_per_year"])


def check_next_inflation_rate(old_inflation_rate, inflation_parameters, bonded_ratio, actual_inflation_rate):
    expected_inflation_rate = next_inflation_rate(old_inflation_rate, inflation_parameters, bonded_ratio)

    assert expected_inflation_rate == Decimal(actual_inflation_rate)
