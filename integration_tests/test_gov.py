from datetime import timedelta

import pytest
from dateutil.parser import isoparse

from .utils import DEFAULT_BASE_PORT, astra_to_aastra, delegate_amount, parse_events, wait_for_block, wait_for_block_time, wait_for_port

pytestmark = pytest.mark.gov


@pytest.mark.parametrize("vote_option", ["yes", "no", "no_with_veto", "abstain", None])
def test_param_proposal(cluster, vote_option):
    """
    - send proposal to change max_validators
    - all validator vote same option (None means don't vote)
    - check the result
    - check deposit refunded
    """
    max_validators = cluster.staking_params()["max_validators"]

    rsp = cluster.gov_propose(
        "team",
        "param-change",
        {
            "title": "Increase number of max validators",
            "description": "ditto",
            "changes": [
                {
                    "subspace": "staking",
                    "key": "MaxValidators",
                    "value": max_validators + 1,
                }
            ],
        },
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    # get proposal_id
    ev = parse_events(rsp["logs"])["submit_proposal"]
    assert ev["proposal_type"] == "ParameterChange", rsp
    proposal_id = ev["proposal_id"]

    proposal = cluster.query_proposal(proposal_id)
    assert proposal["content"]["changes"] == [
        {
            "subspace": "staking",
            "key": "MaxValidators",
            "value": str(max_validators + 1),
        }
    ], proposal
    assert proposal["status"] == "PROPOSAL_STATUS_DEPOSIT_PERIOD", proposal
    # deposit_amount >= gov:min_deposit
    deposit_amount = 10000000000
    amount = cluster.balance(cluster.address("team"))
    rsp = cluster.gov_deposit("team", proposal_id, "%daastra" % deposit_amount)
    assert rsp["code"] == 0, rsp["raw_log"]
    assert cluster.balance(cluster.address("team")) == amount - deposit_amount

    proposal = cluster.query_proposal(proposal_id)
    assert proposal["status"] == "PROPOSAL_STATUS_VOTING_PERIOD", proposal

    if vote_option is not None:
        wait_for_port(DEFAULT_BASE_PORT)
        # node #1
        rsp = cluster.gov_vote("validator", proposal_id, vote_option)
        assert rsp["code"] == 0, rsp["raw_log"]
        wait_for_port(DEFAULT_BASE_PORT)
        # node #2
        rsp = cluster.gov_vote("validator", proposal_id, vote_option, i=1)
        assert rsp["code"] == 0, rsp["raw_log"]
        wait_for_port(DEFAULT_BASE_PORT)
        assert (
                int(cluster.query_tally(proposal_id, i=1)[vote_option])
                == cluster.staking_pool()
        ), "all voted"
    else:
        wait_for_port(DEFAULT_BASE_PORT)
        assert cluster.query_tally(proposal_id) == {
            "yes": "0",
            "no": "0",
            "abstain": "0",
            "no_with_veto": "0",
        }

    wait_for_block_time(
        cluster, isoparse(proposal["voting_end_time"]) + timedelta(seconds=5)
    )

    proposal = cluster.query_proposal(proposal_id)
    if vote_option == "yes":
        assert proposal["status"] == "PROPOSAL_STATUS_PASSED", proposal
    else:
        assert proposal["status"] == "PROPOSAL_STATUS_REJECTED", proposal

    new_max_validators = cluster.staking_params()["max_validators"]
    if vote_option == "yes":
        assert new_max_validators == max_validators + 1
    else:
        assert new_max_validators == max_validators

    if vote_option in ("no_with_veto", None):
        # not refunded
        assert cluster.balance(cluster.address("team")) == amount - deposit_amount
    else:
        # refunded, no matter passed or rejected
        assert cluster.balance(cluster.address("team")) == amount   


def test_deposit_period_expires(cluster):
    """
    - proposal and partially deposit
    - wait for deposit period end and check
    - proposal deleted
    - no refund
    """
    amount1 = cluster.balance(cluster.address("team"))
    # deposit_amount < gov:min_deposit
    deposit_amount = 10000
    rsp = cluster.gov_propose(
        "team",
        "param-change",
        {
            "title": "Increase number of max validators",
            "description": "ditto",
            "changes": [
                {
                    "subspace": "staking",
                    "key": "MaxValidators",
                    "value": 1,
                }
            ],
            "deposit": "%daastra" % deposit_amount,
        },
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    ev = parse_events(rsp["logs"])["submit_proposal"]
    assert ev["proposal_type"] == "ParameterChange", rsp
    proposal_id = ev["proposal_id"]

    proposal = cluster.query_proposal(proposal_id)
    assert proposal["total_deposit"] == [{"denom": "aastra", "amount": str(deposit_amount)}]

    assert cluster.balance(cluster.address("team")) == amount1 - deposit_amount

    amount2 = cluster.balance(cluster.address("community"))
    rsp = cluster.gov_deposit("community", proposal["proposal_id"], "%daastra" % deposit_amount)
    assert rsp["code"] == 0, rsp["raw_log"]
    proposal = cluster.query_proposal(proposal_id)
    assert proposal["total_deposit"] == [{"denom": "aastra", "amount": str(deposit_amount * 2)}]

    assert cluster.balance(cluster.address("community")) == amount2 - deposit_amount

    # wait for deposit period passed
    wait_for_block_time(
        cluster, isoparse(proposal["submit_time"]) + timedelta(seconds=15)
    )

    # proposal deleted
    with pytest.raises(Exception):
        proposal = cluster.query_proposal(proposal_id)

    # deposits don't get refunded
    assert cluster.balance(cluster.address("team")) == amount1 - deposit_amount
    assert cluster.balance(cluster.address("community")) == amount2 - deposit_amount


def test_community_pool_spend_proposal(cluster):
    """
    - proposal a community pool spend
    - pass it
    """
    # need at least several blocks to populate community pool
    wait_for_block(cluster, 3)

    amount = int(cluster.distribution_community())
    print("Distribution community amount: %d" % amount)
    assert amount > 0, "need positive pool to proceed this test"

    recipient = cluster.address("community")
    old_amount = cluster.balance(recipient)
    print("Community old amount: %d" % old_amount)

    deposit_amount = 10000001

    rsp = cluster.gov_propose(
        "community",
        "community-pool-spend",
        {
            "title": "Community Pool Spend",
            "description": "Pay me some astra!",
            "recipient": recipient,
            "amount": "%daastra" % amount,
            "deposit": "%daastra" % deposit_amount,
        },
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    # get proposal_id
    ev = parse_events(rsp["logs"])["submit_proposal"]
    assert ev["proposal_type"] == "CommunityPoolSpend", rsp
    proposal_id = ev["proposal_id"]

    # vote
    rsp = cluster.gov_vote("validator", proposal_id, "yes")
    assert rsp["code"] == 0, rsp["raw_log"]
    rsp = cluster.gov_vote("validator", proposal_id, "yes", i=1)
    assert rsp["code"] == 0, rsp["raw_log"]

    # wait for voting period end
    proposal = cluster.query_proposal(proposal_id)
    assert proposal["total_deposit"] == [{"denom": "aastra", "amount": str(deposit_amount)}]
    assert proposal["status"] == "PROPOSAL_STATUS_VOTING_PERIOD", proposal
    wait_for_block_time(
        cluster, isoparse(proposal["voting_end_time"]) + timedelta(seconds=1)
    )

    proposal = cluster.query_proposal(proposal_id)
    assert proposal["status"] == "PROPOSAL_STATUS_PASSED", proposal
    assert cluster.balance(recipient) == old_amount + amount


def test_change_vote(cluster):
    """
    - submit proposal with deposit
    - vote yes
    - check tally
    - change vote
    - check tally
    """
    deposit_amount = 10000000
    rsp = cluster.gov_propose(
        "community",
        "param-change",
        {
            "title": "Increase number of max validators",
            "description": "ditto",
            "changes": [
                {
                    "subspace": "staking",
                    "key": "MaxValidators",
                    "value": 1,
                }
            ],
            "deposit": "%dastra" % deposit_amount,
        },
    )
    assert rsp["code"] == 0, rsp["raw_log"]

    voting_power = int(
        cluster.validator(cluster.address("validator", bech="val"))["tokens"]
    )
    print(voting_power)

    proposal_id = parse_events(rsp["logs"])["submit_proposal"]["proposal_id"]

    rsp = cluster.gov_vote("validator", proposal_id, "yes")
    assert rsp["code"] == 0, rsp["raw_log"]

    cluster.query_tally(proposal_id) == {
        "yes": str(voting_power),
        "no": "0",
        "abstain": "0",
        "no_with_veto": "0",
    }

    # change vote to no
    rsp = cluster.gov_vote("validator", proposal_id, "no")
    assert rsp["code"] == 0, rsp["raw_log"]

    cluster.query_tally(proposal_id) == {
        "no": str(voting_power),
        "yes": "0",
        "abstain": "0",
        "no_with_veto": "0",
    }


def test_inherit_vote(cluster):
    """
    - submit proposal with deposits
    - A delegate to V
    - V vote Yes
    - check tally: {yes: a + v}
    - A vote No
    - change tally: {yes: v, no: a}
    """
    deposit_amount = 10000000
    rsp = cluster.gov_propose(
        "community",
        "param-change",
        {
            "title": "Increase number of max validators",
            "description": "ditto",
            "changes": [
                {
                    "subspace": "staking",
                    "key": "MaxValidators",
                    "value": 1,
                }
            ],
            "deposit": "%daastra" % deposit_amount,
        },
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    proposal_id = parse_events(rsp["logs"])["submit_proposal"]["proposal_id"]

    delegate_amount = 10
    voter1 = cluster.address("community")
    cluster.delegate_amount(
        # to_addr   amount  from_addr   node
        cluster.address("validator", 1, bech="val"), "%daastra" % delegate_amount, voter1, 0      # delegate to validator #2
    )

    staked_amount_val2 = 1002

    # validator 2 takes the gov vote
    rsp = cluster.gov_vote("validator", proposal_id, "yes", 1)
    assert rsp["code"] == 0, rsp["raw_log"]
    assert cluster.query_tally(proposal_id, 1) == {
        "yes": str(astra_to_aastra(staked_amount_val2) + delegate_amount),
        "no": "0",
        "abstain": "0",
        "no_with_veto": "0",
    }

    rsp = cluster.gov_vote(voter1, proposal_id, "no")
    assert rsp["code"] == 0, rsp["raw_log"]
    
    assert cluster.query_tally(proposal_id) == {
        "yes": str(astra_to_aastra(staked_amount_val2)),
        "no": str(delegate_amount),
        "abstain": "0",
        "no_with_veto": "0",
    }


