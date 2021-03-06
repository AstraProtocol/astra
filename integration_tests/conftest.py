from pathlib import Path

import pytest

from .utils import DEFAULT_BASE_PORT, cluster_fixture
from .network import setup_astra


def pytest_configure(config):
    config.addinivalue_line("markers", "normal: marks normal tests")
    config.addinivalue_line("markers", "authz: marks authz tests")
    config.addinivalue_line("markers", "byzantine: marks byzantine tests")
    config.addinivalue_line("markers", "gov: marks gov tests")
    config.addinivalue_line("markers", "staking: marks staking tests")
    config.addinivalue_line("markers", "vesting: marks vesting tests")


@pytest.fixture(scope="session")
def worker_index():
    # match = re.search(r"\d+", worker_id)
    # return int(match[0]) if match is not None else 0
    return 0


@pytest.fixture(scope="session")
def cluster(worker_index, tmp_path_factory):
    "default cluster fixture"
    yield from cluster_fixture(
        Path(__file__).parent / "configs/default.yaml",
        worker_index,
        # tmp_path_factory.mktemp("data"),
        Path(__file__).parent.parent / "data",
        None,
        None,
        "astrad"
    )


@pytest.fixture(scope="session")
def astra(tmp_path_factory):
    path = tmp_path_factory.mktemp("astra")
    yield from setup_astra(path, DEFAULT_BASE_PORT)


@pytest.fixture(scope="session")
def suspend_capture(pytestconfig):
    """
    used to pause in testing
    Example:
    ```
    def test_simple(suspend_capture):
        with suspend_capture:
            # read user input
            print(input())
    ```
    """

    class SuspendGuard:
        def __init__(self):
            self.capmanager = pytestconfig.pluginmanager.getplugin("capturemanager")

        def __enter__(self):
            self.capmanager.suspend_global_capture(in_=True)

        def __exit__(self, _1, _2, _3):
            self.capmanager.resume_global_capture()

    yield SuspendGuard()