import json
from datetime import timedelta
<<<<<<< HEAD
=======
from pathlib import Path
>>>>>>> origin/main

import pytest
from dateutil.parser import isoparse

<<<<<<< HEAD
from integration_tests.network import setup_astra

=======
>>>>>>> origin/main
from .utils import (
    AUTHORIZATION_DELEGATE,
    AUTHORIZATION_GENERIC,
    AUTHORIZATION_REDELEGATE,
    AUTHORIZATION_SEND,
    AUTHORIZATION_UNBOND,
    AUTHZ,
    AASTRA_DENOM,
    BLOCK_BROADCASTING,
<<<<<<< HEAD
    DEFAULT_BASE_PORT,
=======
>>>>>>> origin/main
    DELEGATE_MSG_TYPE_URL,
    GENERATE_ONLY,
    GRANTS,
    REDELEGATE_MSG_TYPE_URL,
    SEND_MSG_TYPE_URL,
    UNBOND_MSG_TYPE_URL,
    WITHDRAW_DELEGATOR_REWARD_TYPE_URL,
<<<<<<< HEAD
=======
    cluster_fixture,
>>>>>>> origin/main
    delegate_amount,
    exec_tx_by_grantee,
    grant_authorization,
    parse_events,
    query_command,
    query_delegation_amount,
    query_total_reward_amount,
    redelegate_amount,
    revoke_authorization,
    transfer,
    unbond_amount,
    wait_for_block_time,
    wait_for_new_blocks,
    withdraw_all_rewards,
)


@pytest.fixture(scope="class")
<<<<<<< HEAD
def astra_temp(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    yield from setup_astra(path, DEFAULT_BASE_PORT)


class TestAuthzModule:

    def test_execute_tx_within_authorization_spend_limit(self, astra_temp, tmp_path):
=======
def cluster_temp(worker_index, tmp_path_factory):
    "override cluster fixture for this test module"
    yield from cluster_fixture(
        Path(__file__).parent / "configs/default.yaml",
        worker_index,
        tmp_path_factory.mktemp("data"),
        None,
        None,
        "astrad"
    )


class TestAuthzModule:  
    def test_execute_tx_within_authorization_spend_limit(self, cluster, tmp_path):
>>>>>>> origin/main
        """
        test execute transaction within send authorization spend limit
        """
        spend_limit = 200
        transaction_coins = 100
<<<<<<< HEAD
        granter_address = astra_temp.cosmos_cli(0).address("community")
        grantee_address = astra_temp.cosmos_cli(0).address("other_partner")
        receiver_address = astra_temp.cosmos_cli(0).address("treasury")
        granter_initial_balance = astra_temp.cosmos_cli(0).balance(granter_address)
        receiver_initial_balance = astra_temp.cosmos_cli(0).balance(receiver_address)

        grant_authorization(
            astra_temp,
=======
        granter_address = cluster.address("community")
        grantee_address = cluster.address("other_partner")
        receiver_address = cluster.address("treasury")
        granter_initial_balance = cluster.balance(granter_address)
        receiver_initial_balance = cluster.balance(receiver_address)

        grant_authorization(
            cluster,
>>>>>>> origin/main
            grantee_address,
            AUTHORIZATION_SEND,
            granter_address,
            spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
        )
        generated_tx_txt = tmp_path / "generated_tx.txt"
        generated_tx_msg = transfer(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            granter_address,
            receiver_address,
            "%s%s" % (transaction_coins, AASTRA_DENOM),
            GENERATE_ONLY,
        )
        with open(generated_tx_txt, "w") as opened_file:
            json.dump(generated_tx_msg, opened_file)
        exec_tx_by_grantee(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            generated_tx_txt,
            grantee_address,
            broadcast_mode=BLOCK_BROADCASTING,
        )
<<<<<<< HEAD
        wait_for_new_blocks(astra_temp.cosmos_cli(0), 1)

        assert (
            astra_temp.cosmos_cli(0).balance(granter_address)
            == granter_initial_balance - transaction_coins
        )
        assert (
            astra_temp.cosmos_cli(0).balance(receiver_address)
=======
        wait_for_new_blocks(cluster, 1)

        assert (
            cluster.balance(granter_address)
            == granter_initial_balance - transaction_coins
        )
        assert (
            cluster.balance(receiver_address)
>>>>>>> origin/main
            == receiver_initial_balance + transaction_coins
        )

        # teardown
        revoke_authorization(
<<<<<<< HEAD
            astra_temp, grantee_address, SEND_MSG_TYPE_URL, granter_address
        )
        assert (
            len(
                query_command(astra_temp, AUTHZ, GRANTS, granter_address, grantee_address)["grants"]
            ) == 0
        )


    def test_execute_tx_beyond_authorization_spend_limit(self, astra_temp, tmp_path):
=======
            cluster, grantee_address, SEND_MSG_TYPE_URL, granter_address
        )
        assert (
            len(
                query_command(
                    cluster, AUTHZ, GRANTS, granter_address, grantee_address
                )["grants"]
            )
            == 0
        )

    def test_execute_tx_beyond_authorization_spend_limit(self, cluster, tmp_path):
>>>>>>> origin/main
        """
        test execute transaction beyond send authorization spend limit
        """
        spend_limit = 50
        transaction_coins = 100
<<<<<<< HEAD
        granter_address = astra_temp.cosmos_cli(0).address("community")
        grantee_address = astra_temp.cosmos_cli(0).address("other_partner")
        receiver_address = astra_temp.cosmos_cli(0).address("treasury")
        granter_initial_balance = astra_temp.cosmos_cli(0).balance(granter_address)
        receiver_initial_balance = astra_temp.cosmos_cli(0).balance(receiver_address)

        grant_authorization(
            astra_temp,
=======
        granter_address = cluster.address("community")
        grantee_address = cluster.address("other_partner")
        receiver_address = cluster.address("treasury")
        granter_initial_balance = cluster.balance(granter_address)
        receiver_initial_balance = cluster.balance(receiver_address)

        grant_authorization(
            cluster,
>>>>>>> origin/main
            grantee_address,
            AUTHORIZATION_SEND,
            granter_address,
            spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
        )
        generated_tx_txt = tmp_path / "generated_tx.txt"
        generated_tx_msg = transfer(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            granter_address,
            receiver_address,
            "%s%s" % (transaction_coins, AASTRA_DENOM),
            GENERATE_ONLY,
        )
        with open(generated_tx_txt, "w") as opened_file:
            json.dump(generated_tx_msg, opened_file)

        with pytest.raises(
            Exception, match=r".*requested amount is more than spend limit.*"
        ):
<<<<<<< HEAD
            exec_tx_by_grantee(astra_temp, generated_tx_txt, grantee_address)
        assert astra_temp.cosmos_cli(0).balance(
            granter_address) == granter_initial_balance
        assert astra_temp.cosmos_cli(0).balance(
            receiver_address) == receiver_initial_balance

        # teardown
        revoke_authorization(
            astra_temp, grantee_address, SEND_MSG_TYPE_URL, granter_address
        )
        assert(
            len(query_command(astra_temp, AUTHZ, GRANTS, granter_address, grantee_address)["grants"]) == 0
        )


    def test_revoke_authorization(self, astra_temp, tmp_path):
=======
            exec_tx_by_grantee(cluster, generated_tx_txt, grantee_address)
        assert cluster.balance(granter_address) == granter_initial_balance
        assert cluster.balance(receiver_address) == receiver_initial_balance

        # teardown
        revoke_authorization(
            cluster, grantee_address, SEND_MSG_TYPE_URL, granter_address
        )
        assert (
            len(
                query_command(
                    cluster, AUTHZ, GRANTS, granter_address, grantee_address
                )["grants"]
            )
            == 0
        )

    def test_revoke_authorization(self, cluster, tmp_path):
>>>>>>> origin/main
        """
        test unable to execute transaction once grant is revoked
        """
        spend_limit = 200
        transaction_coins = 100
<<<<<<< HEAD
        granter_address = astra_temp.cosmos_cli(0).address("community")
        grantee_address = astra_temp.cosmos_cli(0).address("other_partner")
        receiver_address = astra_temp.cosmos_cli(0).address("treasury")
        granter_initial_balance = astra_temp.cosmos_cli(0).balance(granter_address)
        receiver_initial_balance = astra_temp.cosmos_cli(0).balance(receiver_address)

        grants = query_command(
            astra_temp, AUTHZ, GRANTS, granter_address, grantee_address
=======
        granter_address = cluster.address("community")
        grantee_address = cluster.address("other_partner")
        receiver_address = cluster.address("treasury")
        granter_initial_balance = cluster.balance(granter_address)
        receiver_initial_balance = cluster.balance(receiver_address)

        grants = query_command(
            cluster, AUTHZ, GRANTS, granter_address, grantee_address
>>>>>>> origin/main
        )
        assert len(grants["grants"]) == 0

        grant_authorization(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            grantee_address,
            AUTHORIZATION_SEND,
            granter_address,
            spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
        )
        grants_after_authorization = query_command(
<<<<<<< HEAD
            astra_temp, AUTHZ, GRANTS, granter_address, grantee_address
        )
        assert len(grants_after_authorization["grants"]) == 1
        assert grants_after_authorization["grants"][0]["authorization"]["spend_limit"][0] == {
=======
            cluster, AUTHZ, GRANTS, granter_address, grantee_address
        )
        assert len(grants_after_authorization["grants"]) == 1
        assert grants_after_authorization["grants"][0]["authorization"]["spend_limit"][
            0
        ] == {
>>>>>>> origin/main
            "denom": AASTRA_DENOM,
            "amount": "%s" % spend_limit,
        }

        revoke_authorization(
<<<<<<< HEAD
            astra_temp, grantee_address, SEND_MSG_TYPE_URL, granter_address
        )
        assert (
            len(query_command(astra_temp, AUTHZ, GRANTS, granter_address, grantee_address)["grants"]) == 0
        )
        generated_tx_txt = tmp_path / "generated_tx.txt"
        generated_tx_msg = transfer(
            astra_temp,
=======
            cluster, grantee_address, SEND_MSG_TYPE_URL, granter_address
        )
        assert (
            len(
                query_command(
                    cluster, AUTHZ, GRANTS, granter_address, grantee_address
                )["grants"]
            )
            == 0
        )
        generated_tx_txt = tmp_path / "generated_tx.txt"
        generated_tx_msg = transfer(
            cluster,
>>>>>>> origin/main
            granter_address,
            receiver_address,
            "%s%s" % (transaction_coins, AASTRA_DENOM),
            GENERATE_ONLY,
        )
        with open(generated_tx_txt, "w") as opened_file:
            json.dump(generated_tx_msg, opened_file)

        with pytest.raises(Exception, match=r".*authorization not found.*"):
<<<<<<< HEAD
            exec_tx_by_grantee(astra_temp, generated_tx_txt, grantee_address)
        assert astra_temp.cosmos_cli(0).balance(granter_address) == granter_initial_balance
        assert astra_temp.cosmos_cli(0).balance(receiver_address) == receiver_initial_balance


    def test_generic_authorization_grant(self, astra_temp, tmp_path):
=======
            exec_tx_by_grantee(cluster, generated_tx_txt, grantee_address)
        assert cluster.balance(granter_address) == granter_initial_balance
        assert cluster.balance(receiver_address) == receiver_initial_balance

    def test_generic_authorization_grant(self, cluster, tmp_path):
>>>>>>> origin/main
        """
        test grant authorization with generic authorization with send msg type
        """
        delegate_coins = 1000000
<<<<<<< HEAD
        validator_address = astra_temp.cosmos_cli(0).validators()[0]["operator_address"]
        granter_address = astra_temp.cosmos_cli(0).address("community")
        grantee_address = astra_temp.cosmos_cli(0).address("other_partner")
        granter_initial_reward_amount = query_total_reward_amount(
            astra_temp, granter_address, validator_address
        )

        delegate_amount(
            astra_temp,
=======
        validator_address = cluster.validators()[0]["operator_address"]
        granter_address = cluster.address("community")
        grantee_address = cluster.address("other_partner")
        granter_initial_reward_amount = query_total_reward_amount(
            cluster, granter_address, validator_address
        )

        delegate_amount(
            cluster,
>>>>>>> origin/main
            validator_address,
            "%s%s" % (delegate_coins, AASTRA_DENOM),
            granter_address,
            broadcast_mode=BLOCK_BROADCASTING,
        )
        # wait for some reward
<<<<<<< HEAD
        wait_for_new_blocks(astra_temp.cosmos_cli(0), 2)
        granter_balance_bef_rewards_withdrawal = astra_temp.cosmos_cli(0).balance(granter_address)

        grant_authorization(
            astra_temp,
=======
        wait_for_new_blocks(cluster, 2)
        granter_balance_bef_rewards_withdrawal = cluster.balance(granter_address)

        grant_authorization(
            cluster,
>>>>>>> origin/main
            grantee_address,
            AUTHORIZATION_GENERIC,
            granter_address,
            msg_type=WITHDRAW_DELEGATOR_REWARD_TYPE_URL,
        )

        generated_tx_txt = tmp_path / "generated_tx.txt"
        generated_tx_msg = withdraw_all_rewards(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            granter_address,
            GENERATE_ONLY,
        )
        with open(generated_tx_txt, "w") as opened_file:
            json.dump(generated_tx_msg, opened_file)
        exec_tx_by_grantee(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            generated_tx_txt,
            grantee_address,
            broadcast_mode=BLOCK_BROADCASTING,
        )
<<<<<<< HEAD
        wait_for_new_blocks(astra_temp.cosmos_cli(0), 1)

        assert(
            astra_temp.cosmos_cli(0).balance(granter_address) - granter_balance_bef_rewards_withdrawal
=======
        wait_for_new_blocks(cluster, 1)

        assert (
            cluster.balance(granter_address)
            - granter_balance_bef_rewards_withdrawal
>>>>>>> origin/main
            > granter_initial_reward_amount
        )

        # teardown
        revoke_authorization(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            grantee_address,
            WITHDRAW_DELEGATOR_REWARD_TYPE_URL,
            granter_address,
        )
<<<<<<< HEAD
        assert(
            len(query_command(astra_temp, AUTHZ, GRANTS, granter_address, grantee_address)["grants"]) == 0
        )


    def test_execute_delegate_to_allowed_validator(self, astra_temp, tmp_path):
=======
        assert (
            len(
                query_command(
                    cluster, AUTHZ, GRANTS, granter_address, grantee_address
                )["grants"]
            )
            == 0
        )

    def test_execute_delegate_to_allowed_validator(self, cluster, tmp_path):
>>>>>>> origin/main
        """
        test execute delegate to allowed validator should succeed
        test execute delegate to other validators should fail
        """
        # test execute delegate to allowed validator should succeed
        spend_limit = 200
        delegate_coins = 100
<<<<<<< HEAD
        allowed_validator_address = astra_temp.cosmos_cli(0).validators()[0]["operator_address"]
        another_validator_address = astra_temp.cosmos_cli(0).validators()[1]["operator_address"]
        granter_address = astra_temp.cosmos_cli(0).address("community")
        grantee_address = astra_temp.cosmos_cli(0).address("other_partner")
        validator_initial_delegation_amount = int(
            query_delegation_amount(astra_temp, granter_address, allowed_validator_address)["amount"]
        )

        grant_authorization(
            astra_temp,
=======
        allowed_validator_address = cluster.validators()[0]["operator_address"]
        another_validator_address = cluster.validators()[1]["operator_address"]
        granter_address = cluster.address("community")
        grantee_address = cluster.address("other_partner")
        validator_initial_delegation_amount = int(
            query_delegation_amount(
                cluster, granter_address, allowed_validator_address
            )["amount"]
        )

        grant_authorization(
            cluster,
>>>>>>> origin/main
            grantee_address,
            AUTHORIZATION_DELEGATE,
            granter_address,
            spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
            allowed_validators=allowed_validator_address,
        )
        generated_delegate_txt = tmp_path / "generated_delegate.txt"
        generated_delegate_msg1 = delegate_amount(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            allowed_validator_address,
            "%s%s" % (delegate_coins, AASTRA_DENOM),
            granter_address,
            GENERATE_ONLY,
        )
        with open(generated_delegate_txt, "w") as opened_file:
            json.dump(generated_delegate_msg1, opened_file)
<<<<<<< HEAD
        exec_tx_by_grantee(astra_temp, generated_delegate_txt, grantee_address)

        assert query_delegation_amount(
            astra_temp, granter_address, allowed_validator_address) == {
                "denom": AASTRA_DENOM,
                "amount": str(validator_initial_delegation_amount + delegate_coins),
            }

        # test execute delegate to other validators not in allowed validators should fail # noqa: E501
        another_validator_initial_delegation_amount = int(
            query_delegation_amount(astra_temp, granter_address, another_validator_address)["amount"]
        )
        generated_delegate_msg2 = delegate_amount(
            astra_temp,
=======
        exec_tx_by_grantee(cluster, generated_delegate_txt, grantee_address)

        assert query_delegation_amount(
            cluster, granter_address, allowed_validator_address
        ) == {
            "denom": AASTRA_DENOM,
            "amount": str(validator_initial_delegation_amount + delegate_coins),
        }

        # test execute delegate to other validators not in allowed validators should fail # noqa: E501
        another_validator_initial_delegation_amount = int(
            query_delegation_amount(
                cluster, granter_address, another_validator_address
            )["amount"]
        )
        generated_delegate_msg2 = delegate_amount(
            cluster,
>>>>>>> origin/main
            another_validator_address,
            "%s%s" % (delegate_coins, AASTRA_DENOM),
            granter_address,
            GENERATE_ONLY,
        )
        with open(generated_delegate_txt, "w") as opened_file:
            json.dump(generated_delegate_msg2, opened_file)

        with pytest.raises(Exception, match=r".*unauthorized.*"):
<<<<<<< HEAD
            exec_tx_by_grantee(astra_temp, generated_delegate_txt, grantee_address)
        assert query_delegation_amount(
            astra_temp, granter_address, another_validator_address) == {
                "denom": AASTRA_DENOM,
                "amount": str(another_validator_initial_delegation_amount),
            }

        # teardown
        revoke_authorization(
            astra_temp, grantee_address, DELEGATE_MSG_TYPE_URL, granter_address
        )
        assert(
            len(query_command(astra_temp, AUTHZ, GRANTS, granter_address, grantee_address)["grants"]) == 0
        )


    def test_unable_to_execute_delegate_to_deny_validator(self, astra_temp, tmp_path):
=======
            exec_tx_by_grantee(cluster, generated_delegate_txt, grantee_address)
        assert query_delegation_amount(
            cluster, granter_address, another_validator_address
        ) == {
            "denom": AASTRA_DENOM,
            "amount": str(another_validator_initial_delegation_amount),
        }

        # teardown
        revoke_authorization(
            cluster, grantee_address, DELEGATE_MSG_TYPE_URL, granter_address
        )
        assert (
            len(
                query_command(
                    cluster, AUTHZ, GRANTS, granter_address, grantee_address
                )["grants"]
            )
            == 0
        )

    def test_unable_to_execute_delegate_to_deny_validator(self, cluster, tmp_path):
>>>>>>> origin/main
        """
        test execute delegate to deny validator should fail
        """
        spend_limit = 200
        delegate_coins = 100
<<<<<<< HEAD
        deny_validator_address = astra_temp.cosmos_cli(0).validators()[0]["operator_address"]
        granter_address = astra_temp.cosmos_cli(0).address("community")
        grantee_address = astra_temp.cosmos_cli(0).address("other_partner")

        # test execute delegate to deny validator should fail
        grant_authorization(
            astra_temp,
=======
        deny_validator_address = cluster.validators()[0]["operator_address"]
        granter_address = cluster.address("community")
        grantee_address = cluster.address("other_partner")

        # test execute delegate to deny validator should fail
        grant_authorization(
            cluster,
>>>>>>> origin/main
            grantee_address,
            AUTHORIZATION_DELEGATE,
            granter_address,
            spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
            deny_validators=deny_validator_address,
        )
        generated_delegate_txt = tmp_path / "generated_delegate.txt"
        generated_delegate_msg1 = delegate_amount(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            deny_validator_address,
            "%s%s" % (delegate_coins, AASTRA_DENOM),
            granter_address,
            GENERATE_ONLY,
        )
        with open(generated_delegate_txt, "w") as opened_file:
            json.dump(generated_delegate_msg1, opened_file)

        with pytest.raises(Exception, match=r".*unauthorized.*"):
<<<<<<< HEAD
            exec_tx_by_grantee(astra_temp, generated_delegate_txt, grantee_address)

        # teardown
        revoke_authorization(
            astra_temp, grantee_address, DELEGATE_MSG_TYPE_URL, granter_address
        )
        assert(
            len(query_command(astra_temp, AUTHZ, GRANTS, granter_address, grantee_address)["grants"]) == 0
        )


    def test_execute_all_staking_operations(self, astra_temp, tmp_path):
=======
            exec_tx_by_grantee(cluster, generated_delegate_txt, grantee_address)

        # teardown
        revoke_authorization(
            cluster, grantee_address, DELEGATE_MSG_TYPE_URL, granter_address
        )
        assert (
            len(
                query_command(
                    cluster, AUTHZ, GRANTS, granter_address, grantee_address
                )["grants"]
            )
            == 0
        )

    def test_execute_all_staking_operations(self, cluster, tmp_path):
>>>>>>> origin/main
        """
        test execute delegate, unbond, redelegate by grantee
        """
        spend_limit = 200
        delegate_coins = 100
        unbond_coins = 50
        redelegate_coins = 30
<<<<<<< HEAD
        validator1_address = astra_temp.cosmos_cli(0).validators()[0]["operator_address"]
        validator2_address = astra_temp.cosmos_cli(0).validators()[1]["operator_address"]
        granter_address = astra_temp.cosmos_cli(0).address("community")
        grantee_address = astra_temp.cosmos_cli(0).address("other_partner")
        validator1_initial_deligation_amount = int(
            query_delegation_amount(astra_temp, granter_address, validator1_address)["amount"]
        )
        validator2_initial_deligation_amount = int(
            query_delegation_amount(astra_temp, granter_address, validator2_address)["amount"]
=======
        validator1_address = cluster.validators()[0]["operator_address"]
        validator2_address = cluster.validators()[1]["operator_address"]
        granter_address = cluster.address("community")
        grantee_address = cluster.address("other_partner")
        validator1_initial_deligation_amount = int(
            query_delegation_amount(cluster, granter_address, validator1_address)[
                "amount"
            ]
        )
        validator2_initial_deligation_amount = int(
            query_delegation_amount(cluster, granter_address, validator2_address)[
                "amount"
            ]
>>>>>>> origin/main
        )

        # test execute delegate
        grant_authorization(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            grantee_address,
            AUTHORIZATION_DELEGATE,
            granter_address,
            spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
            allowed_validators=validator1_address,
        )
        generated_delegate_txt = tmp_path / "generated_delegate.txt"
        generated_delegate_msg = delegate_amount(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            validator1_address,
            "%s%s" % (delegate_coins, AASTRA_DENOM),
            granter_address,
            GENERATE_ONLY,
        )
        with open(generated_delegate_txt, "w") as opened_file:
            json.dump(generated_delegate_msg, opened_file)
<<<<<<< HEAD
        exec_tx_by_grantee(astra_temp, generated_delegate_txt, grantee_address)

        assert query_delegation_amount(astra_temp, granter_address, validator1_address) == {
=======
        exec_tx_by_grantee(cluster, generated_delegate_txt, grantee_address)

        assert query_delegation_amount(
            cluster, granter_address, validator1_address
        ) == {
>>>>>>> origin/main
            "denom": AASTRA_DENOM,
            "amount": str(validator1_initial_deligation_amount + delegate_coins),
        }

        # test execute unbond
        grant_authorization(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            grantee_address,
            AUTHORIZATION_UNBOND,
            granter_address,
            spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
            allowed_validators=validator1_address,
        )
        generated_unbond_txt = tmp_path / "generated_unbond.txt"
        generated_unbond_msg = unbond_amount(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            validator1_address,
            "%s%s" % (unbond_coins, AASTRA_DENOM),
            granter_address,
            GENERATE_ONLY,
        )
        with open(generated_unbond_txt, "w") as opened_file:
            json.dump(generated_unbond_msg, opened_file)
<<<<<<< HEAD
        rsp = exec_tx_by_grantee(astra_temp, generated_unbond_txt, grantee_address)
        wait_for_block_time(
            astra_temp.cosmos_cli(0),
            isoparse(parse_events(rsp["logs"])["unbond"]["completion_time"]) + timedelta(seconds=1),
        )

        assert query_delegation_amount(astra_temp, granter_address, validator1_address) == {
=======
        rsp = exec_tx_by_grantee(cluster, generated_unbond_txt, grantee_address)
        wait_for_block_time(
            cluster,
            isoparse(parse_events(rsp["logs"])["unbond"]["completion_time"])
            + timedelta(seconds=1),
        )

        assert query_delegation_amount(
            cluster, granter_address, validator1_address
        ) == {
>>>>>>> origin/main
            "denom": AASTRA_DENOM,
            "amount": str(
                validator1_initial_deligation_amount + delegate_coins - unbond_coins
            ),
        }

        # test execute redelegate
        grant_authorization(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            grantee_address,
            AUTHORIZATION_REDELEGATE,
            granter_address,
            spend_limit="%s%s" % (spend_limit, AASTRA_DENOM),
            allowed_validators=validator2_address,
        )
        generated_redelegate_txt = tmp_path / "generated_redelegate.txt"
        generated_redelegate_msg = redelegate_amount(
<<<<<<< HEAD
            astra_temp,
=======
            cluster,
>>>>>>> origin/main
            validator1_address,
            validator2_address,
            "%s%s" % (redelegate_coins, AASTRA_DENOM),
            granter_address,
            GENERATE_ONLY,
        )
        with open(generated_redelegate_txt, "w") as opened_file:
            json.dump(generated_redelegate_msg, opened_file)
<<<<<<< HEAD
        exec_tx_by_grantee(astra_temp, generated_redelegate_txt, grantee_address)

        assert query_delegation_amount(
            astra_temp, granter_address, validator1_address) == {
                "denom": AASTRA_DENOM,
                "amount": str(
                    validator1_initial_deligation_amount + delegate_coins - unbond_coins - redelegate_coins
                ),
            }
        assert query_delegation_amount(astra_temp, granter_address, validator2_address) == {
=======
        exec_tx_by_grantee(cluster, generated_redelegate_txt, grantee_address)

        assert query_delegation_amount(
            cluster, granter_address, validator1_address
        ) == {
            "denom": AASTRA_DENOM,
            "amount": str(
                validator1_initial_deligation_amount
                + delegate_coins
                - unbond_coins
                - redelegate_coins
            ),
        }
        assert query_delegation_amount(
            cluster, granter_address, validator2_address
        ) == {
>>>>>>> origin/main
            "denom": AASTRA_DENOM,
            "amount": str(validator2_initial_deligation_amount + redelegate_coins),
        }

        # teardown
        revoke_authorization(
<<<<<<< HEAD
            astra_temp, grantee_address, DELEGATE_MSG_TYPE_URL, granter_address
        )
        revoke_authorization(
            astra_temp, grantee_address, UNBOND_MSG_TYPE_URL, granter_address
        )
        revoke_authorization(
            astra_temp, grantee_address, REDELEGATE_MSG_TYPE_URL, granter_address
        )
        assert(
            len(query_command(astra_temp, AUTHZ, GRANTS, granter_address, grantee_address)["grants"]) == 0
        )
=======
            cluster, grantee_address, DELEGATE_MSG_TYPE_URL, granter_address
        )
        revoke_authorization(
            cluster, grantee_address, UNBOND_MSG_TYPE_URL, granter_address
        )
        revoke_authorization(
            cluster, grantee_address, REDELEGATE_MSG_TYPE_URL, granter_address
        )
        assert (
            len(
                query_command(
                    cluster, AUTHZ, GRANTS, granter_address, grantee_address
                )["grants"]
            )
            == 0
        )    
>>>>>>> origin/main
