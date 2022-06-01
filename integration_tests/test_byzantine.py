import sys
import time
from pathlib import Path

import pytest
from integration_tests.network import setup_astra
from .utils import DEFAULT_BASE_PORT

MAX_SLEEP_SEC = 600

pytestmark = pytest.mark.byzantine


@pytest.fixture(scope="class")
def astra_byzantine(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    cfg = Path(__file__).parent / "configs/byzantine.yaml"
    yield from setup_astra(path, DEFAULT_BASE_PORT, cfg)
    

class TestByzantineModule:

    def test_byzantine(self, astra_byzantine):
        """
        - 3 nodes
        - node0 has more than 2/3 voting powers
        - stop node2
        - copy node1's validator key to node2
        - start node2
        - check node1 & node2 jailed
        """
        assert len(astra_byzantine.cosmos_cli(0).validators()) == 3
        from_node = 1
        to_node = 2
        val_addr_byzantine = astra_byzantine.cosmos_cli(from_node).address("validator", bech="val")
        val_addr_slash = astra_byzantine.cosmos_cli(to_node).address("validator", bech="val")
        tokens_byzantine_before = int((astra_byzantine.cosmos_cli(0).validator(val_addr_byzantine))["tokens"])
        tokens_slash_before = int((astra_byzantine.cosmos_cli(0).validator(val_addr_slash))["tokens"])
        astra_byzantine.cosmos_cli(0).stop_node(to_node)
        astra_byzantine.cosmos_cli(0).copy_validator_key(from_node, to_node)
        astra_byzantine.cosmos_cli(0).start_node(to_node)

        # it may take 30s to finish the loop
        i = 0
        while i < MAX_SLEEP_SEC:
            time.sleep(1)
            sys.stdout.write(".")
            sys.stdout.flush()
            i += 1
            val1 = astra_byzantine.cosmos_cli(0).validator(val_addr_byzantine)
            if val1["jailed"]:
                break
        assert val1["jailed"]
        assert val1["status"] == "BOND_STATUS_UNBONDING"
        print("\n{}s waiting for node 1 jailed".format(i))

        # it may take 2min to finish the loop
        i = 0
        while i < MAX_SLEEP_SEC:
            time.sleep(1)
            i += 1
            sys.stdout.write(".")
            sys.stdout.flush()
            val2 = astra_byzantine.cosmos_cli(0).validator(val_addr_slash)
            if val2["jailed"]:
                break
        assert val2["jailed"]
        assert val2["status"] == "BOND_STATUS_UNBONDING"
        print("\n{}s waiting for node 2 jailed".format(i))

        tokens_byzantine_after = int((astra_byzantine.cosmos_cli(0).validator(val_addr_byzantine))["tokens"])
        tokens_slash_after = int((astra_byzantine.cosmos_cli(0).validator(val_addr_slash))["tokens"])
        assert tokens_byzantine_before * 0.95 == tokens_byzantine_after
        assert tokens_slash_before * 0.99 == tokens_slash_after