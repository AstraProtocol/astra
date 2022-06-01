import json
import os
import signal
import subprocess
from pathlib import Path

import web3
from pystarport import ports

from .cosmoscli import CosmosCLI
from .utils import wait_for_port, wait_for_block


class Astra:
    def __init__(self, base_dir):
        self._w3 = None
        self.base_dir = base_dir
        self.enable_auto_deployment = True
        self._use_websockets = False

    def copy(self):
        return Astra(self.base_dir)

    @property
    def w3_http_endpoint(self, i=0):
        port = ports.evmrpc_port(self.base_port(i))
        # port = 8545
        return f"http://localhost:{port}"

    @property
    def w3_ws_endpoint(self, i=0):
        port = ports.evmrpc_ws_port(self.base_port(i))
        # port = 8546
        return f"ws://localhost:{port}"

    @property
    def w3(self, i=0):
        if self._w3 is None:
            if self._use_websockets:
                self._w3 = web3.Web3(
                    web3.providers.WebsocketProvider(self.w3_ws_endpoint)
                )
            else:
                self._w3 = web3.Web3(web3.providers.HTTPProvider(self.w3_http_endpoint))
        return self._w3

    def base_port(self, i):
        config = json.loads((self.base_dir / "config.json").read_text())
        return config["validators"][i]["base_port"]

    def node_rpc(self, i):
        return "tcp://127.0.0.1:%d" % ports.rpc_port(self.base_port(i))

    def cosmos_cli(self, i=0):
        return CosmosCLI(self.base_dir / f"node{i}", self.node_rpc(i), "astrad")

    def use_websocket(self, use=True):
        self._w3 = None
        self._use_websockets = use


def setup_astra(path, base_port, cfg=None, enable_auto_deployment=True):
    if cfg == None:
        cfg = Path(__file__).parent / "configs/default.yaml"
    yield from setup_custom_astra(path, base_port, cfg)     


def setup_custom_astra(path, base_port, config, post_init=None, chain_binary=None):
    cmd = [
        "pystarport",
        "init",
        "--config",
        config,
        "--data",
        path,
        "--base_port",
        str(base_port),
        "--no_remove",
    ]
    if chain_binary is not None:
        cmd = cmd[:1] + ["--cmd", chain_binary] + cmd[1:]
    print(*cmd)
    subprocess.run(cmd, check=True)
    if post_init is not None:
        post_init(path, base_port, config)
    proc = subprocess.Popen(
        ["pystarport", "start", "--data", path, "--quiet"],
        preexec_fn=os.setsid,
    )
    try:
        wait_for_port(ports.evmrpc_port(base_port))
        wait_for_port(ports.evmrpc_ws_port(base_port))

        cluster = Astra(path / "astra_777-1")
        wait_for_port(ports.rpc_port(base_port))
        # wait for the first block generated before start testing
        wait_for_block(cluster.cosmos_cli(0), 2)

        yield cluster
    finally:
        os.killpg(os.getpgid(proc.pid), signal.SIGTERM)
        # proc.terminate()
        proc.wait()       