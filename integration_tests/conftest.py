from pathlib import Path

import pytest

from .network import setup_astra, setup_geth
from .utils import cluster_fixture


def pytest_configure(config):
    config.addinivalue_line("markers", "slow: marks tests as slow")
    config.addinivalue_line("markers", "ledger: marks tests as ledger hardware test")
    config.addinivalue_line("markers", "grpc: marks grpc tests")
    config.addinivalue_line("markers", "upgrade: marks upgrade tests")
    config.addinivalue_line("markers", "normal: marks normal tests")
    config.addinivalue_line("markers", "ibc: marks ibc tests")
    config.addinivalue_line("markers", "byzantine: marks byzantine tests")
    config.addinivalue_line("markers", "gov: marks gov tests")


@pytest.fixture(scope="session")
def astra(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    yield from setup_astra(path, 26650)


@pytest.fixture(scope="session")
def geth(tmp_path_factory):
    path = tmp_path_factory.mktemp("geth")
    yield from setup_geth(path, 8545)


@pytest.fixture(scope="session", params=["astra", "geth"])
def cluster(request, astra, geth):
    """
    run on both cronos and geth
    """
    provider = request.param
    if provider == "astra":
        yield astra
    elif provider == "geth":
        yield geth
    else:
        raise NotImplementedError

# @pytest.fixture(scope="session")
# def cluster(worker_index, tmp_path_factory):
#     "default cluster fixture"
#     yield from cluster_fixture(
#         Path(__file__).parent / "configs/astra-devnet.yaml",
#         worker_index,
#         # tmp_path_factory.mktemp("data"),
#         Path(__file__).parent.parent / "data",
#         None,
#         None,
#         "astrad"
#     )


@pytest.fixture(scope="session")
def suspend_capture(pytestconfig):
    "used for pause in testing"

    class SuspendGuard:
        def __init__(self):
            self.capmanager = pytestconfig.pluginmanager.getplugin("capturemanager")

        def __enter__(self):
            self.capmanager.suspend_global_capture(in_=True)

        def __exit__(self, _1, _2, _3):
            self.capmanager.resume_global_capture()

    yield SuspendGuard()
