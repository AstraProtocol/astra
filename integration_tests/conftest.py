from pathlib import Path

import pytest

from .utils import cluster_fixture

#import re


@pytest.fixture(scope="session")
def worker_index():
    #match = re.search(r"\d+", worker_id)
    #return int(match[0]) if match is not None else 0
    return 0


@pytest.fixture(scope="session")
def cluster(worker_index, tmp_path_factory):
    "default cluster fixture"
    yield from cluster_fixture(
        Path(__file__).parent / "configs/default.yaml",
        worker_index,
        tmp_path_factory.mktemp("data"),
        #Path(__file__).parent.parent / "data",
        None,
        None,
        "astrad"
    )


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
