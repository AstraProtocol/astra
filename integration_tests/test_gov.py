from datetime import timedelta

import pytest
from dateutil.parser import isoparse

from .utils import astra_to_aastra, parse_events, wait_for_block, wait_for_block_time

pytestmark = pytest.mark.gov


# @pytest.mark.parametrize("vote_option", ["yes", "no", "no_with_veto", "abstain", None])
@pytest.mark.parametrize("vote_option", ["yes"])
def test_param_proposal(cluster, vote_option):
    """
    - send proposal to change max_validators
    - all validator vote same option (None means don't vote)
    - check the result
    - check deposit refunded
    """
    max_validators = cluster.staking_params()["max_validators"]

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

    amount = cluster.balance(cluster.address("team"))
    rsp = cluster.gov_deposit("team", proposal_id, "10000000000aastra")
    assert rsp["code"] == 0, rsp["raw_log"]
    assert cluster.balance(cluster.address("team")) == amount - 10000000000

    proposal = cluster.query_proposal(proposal_id)
    print(proposal)
    assert proposal["status"] == "PROPOSAL_STATUS_VOTING_PERIOD", proposal

    if vote_option is not None:
        rsp = cluster.gov_vote("validator", proposal_id, vote_option)
        assert rsp["code"] == 0, rsp["raw_log"]
        rsp = cluster.gov_vote("validator", proposal_id, vote_option, i=1)
        assert rsp["code"] == 0, rsp["raw_log"]
        assert (
                int(cluster.query_tally(proposal_id, i=1)[vote_option])
                == cluster.staking_pool()
        ), "all voted"
    else:
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
        assert cluster.balance(cluster.address("team")) == amount - 100000000
    else:
        # refunded, no matter passed or rejected
        assert cluster.balance(cluster.address("team")) == amount

