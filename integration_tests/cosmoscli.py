import configparser
import enum
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys
import tempfile
from time import sleep

import bech32
import jsonmerge
from dateutil.parser import isoparse
from pystarport.utils import build_cli_args_safe, format_doc_string, interact
from pystarport import ports
import tomlkit
from .utils import DEFAULT_GAS_PRICE, SUPERVISOR_CONFIG_FILE, DEFAULT_GAS

COMMON_PROG_OPTIONS = {
    # redirect to supervisord's stdout, easier to collect all logs
    "autostart": "true",
    "autorestart": "true",
    "redirect_stderr": "true",
    "startsecs": "3",
}


class ModuleAccount(enum.Enum):
    FeeCollector = "fee_collector"
    Mint = "mint"
    Gov = "gov"
    Distribution = "distribution"
    BondedPool = "bonded_tokens_pool"
    NotBondedPool = "not_bonded_tokens_pool"
    IBCTransfer = "transfer"


@format_doc_string(
    options=",".join(v.value for v in ModuleAccount.__members__.values())
)
def module_address(name):
    """
    get address of module accounts
    :param name: name of module account, values: {options}
    """
    data = hashlib.sha256(ModuleAccount(name).value.encode()).digest()[:20]
    return bech32.bech32_encode("astra", bech32.convertbits(data, 8, 5))


def home_dir(data_dir, i):
    return data_dir / f"node{i}"


class ChainCommand:
    def __init__(self, cmd):
        self.cmd = cmd

    def __call__(self, cmd, *args, stdin=None, **kwargs):
        "execute astrad"
        args = " ".join(build_cli_args_safe(cmd, *args, **kwargs))
        result = ""
        tried = 0
        # trying to interact with chain 10 times when socket error occurs
        while result == "":
            tried += 1
            if tried == 10:
                return interact(f"{self.cmd} {args}", input=stdin)
            try:
                result = interact(f"{self.cmd} {args}", input=stdin)
            except:
                result = ""
                sleep(0.5)
        return result


class CosmosCLI:
    "the apis to interact with wallet and blockchain"

    def __init__(
            self,
            data_dir,
            node_rpc,
            cmd,
    ):
        self.data_dir = data_dir
        self._genesis = json.loads(
            (self.data_dir / "config" / "genesis.json").read_text()
        )
        self.chain_id = self._genesis["chain_id"]
        self.node_rpc = node_rpc
        self.cmd = cmd
        self.raw = ChainCommand(cmd)
        self.config = json.load((self.data_dir / "../" / "config.json").open())
        self.output = None
        self.error = None

    def reload_supervisor(self):
        subprocess.run(
            [
                sys.executable,
                "-msupervisor.supervisorctl",
                "-c",
                Path(self.data_dir / "../../") / SUPERVISOR_CONFIG_FILE,
                "update",
            ],
            check=True,
        )

    def node_id(self):
        "get tendermint node id"
        output = self.raw("tendermint", "show-node-id", home=self.data_dir)
        return output.decode().strip()

    def base_port(self, i):
        return self.config["validators"][i]["base_port"]

    def get_node_rpc(self, i):
        "rpc url of i-th node"
        return "tcp://127.0.0.1:%d" % ports.rpc_port(self.base_port(i))

    def delete_account(self, name):
        "delete wallet account in node's keyring"
        return self.raw(
            "keys",
            "delete",
            name,
            "-y",
            "--force",
            home=self.data_dir,
            output="json",
            keyring_backend="test",
        )

    def create_account(self, name, mnemonic=None):
        "create new keypair in node's keyring"
        if mnemonic is None:
            output = self.raw(
                "keys",
                "add",
                name,
                home=self.data_dir,
                output="json",
                keyring_backend="test",
            )
        else:
            output = self.raw(
                "keys",
                "add",
                name,
                "--recover",
                home=self.data_dir,
                output="json",
                keyring_backend="test",
                stdin=mnemonic.encode() + b"\n",
            )
        return json.loads(output)

    def create_account_specific_node(self, name, mnemonic=None, i=0):
        "create new keypair in node's keyring"
        data_path = self.data_dir / "../"
        if mnemonic is None:
            output = self.raw(
                "keys",
                "add",
                name,
                home=data_path / str("node" + str(i)),
                output="json",
                keyring_backend="test",
            )
        else:
            output = self.raw(
                "keys",
                "add",
                name,
                "--recover",
                home=data_path / str("node" + str(i)),
                output="json",
                keyring_backend="test",
                stdin=mnemonic.encode() + b"\n",
            )
        return json.loads(output)

    def init(self, moniker):
        "the node's config is already added"
        return self.raw(
            "init",
            moniker,
            chain_id=self.chain_id,
            home=self.data_dir,
        )

    def init_new_node(self, moniker, i):
        "the node's config is already added"
        data_path = self.data_dir / "../"
        return self.raw(
            "init",
            moniker,
            chain_id=self.chain_id,
            home=data_path / str("node" + str(i)),
        )

    def home(self, i):
        "home directory of i-th node"
        return home_dir(Path(self.data_dir).parent, i)

    def validate_genesis(self):
        return self.raw("validate-genesis", home=self.data_dir)

    def add_genesis_account(self, addr, coins, **kwargs):
        return self.raw(
            "add-genesis-account",
            addr,
            coins,
            home=self.data_dir,
            output="json",
            **kwargs,
        )

    def gentx(self, name, coins, min_self_delegation=1, pubkey=None):
        return self.raw(
            "gentx",
            name,
            coins,
            min_self_delegation=str(min_self_delegation),
            home=self.data_dir,
            chain_id=self.chain_id,
            keyring_backend="test",
            pubkey=pubkey,
        )

    def collect_gentxs(self, gentx_dir):
        return self.raw("collect-gentxs", gentx_dir, home=self.data_dir)

    def status(self):
        return json.loads(self.raw("status", node=self.node_rpc))

    def block_height(self):
        return int(self.status()["SyncInfo"]["latest_block_height"])

    def block_time(self):
        return isoparse(self.status()["SyncInfo"]["latest_block_time"])

    def balances(self, addr):
        return json.loads(
            self.raw("query", "bank", "balances", addr, home=self.data_dir)
        )["balances"]

    def balance(self, addr, denom="aastra"):
        denoms = {coin["denom"]: int(coin["amount"]) for coin in self.balances(addr)}
        return denoms.get(denom, 0)

    def query_tx(self, tx_type, tx_value):
        tx = self.raw(
            "query",
            "tx",
            "--type",
            tx_type,
            tx_value,
            home=self.data_dir,
            chain_id=self.chain_id,
            node=self.node_rpc,
        )
        return json.loads(tx)

    def query_all_txs(self, addr):
        txs = self.raw(
            "query",
            "txs-all",
            addr,
            home=self.data_dir,
            keyring_backend="test",
            chain_id=self.chain_id,
            node=self.node_rpc,
        )
        return json.loads(txs)

    def distribution_params(self):
        return json.loads(
            self.raw(
                "query",
                "distribution",
                "params",
                output="json",
                node=self.node_rpc,
            )
        )

    def distribution_commission(self, addr):
        coin = json.loads(
            self.raw(
                "query",
                "distribution",
                "commission",
                addr,
                output="json",
                node=self.node_rpc,
            )
        )["commission"][0]
        return float(coin["amount"])

    def distribution_community(self):
        coin = json.loads(
            self.raw(
                "query",
                "distribution",
                "community-pool",
                output="json",
                node=self.node_rpc,
            )
        )["pool"]
        if len(coin) > 0:
            return float(coin[0]["amount"])

        return float(0.0)

    def distribution_reward(self, delegator_addr):
        coin = json.loads(
            self.raw(
                "query",
                "distribution",
                "rewards",
                delegator_addr,
                output="json",
                node=self.node_rpc,
            )
        )["total"][0]
        return float(coin["amount"])

    def address(self, name, bech="acc"):
        output = self.raw(
            "keys",
            "show",
            name,
            "-a",
            home=self.data_dir,
            keyring_backend="test",
            bech=bech,
        )
        return output.strip().decode()

    def account(self, addr):
        return json.loads(
            self.raw(
                "query", "auth", "account", addr, output="json", node=self.node_rpc
            )
        )

    def total_supply(self):
        return json.loads(
            self.raw("query", "bank", "total", output="json", node=self.node_rpc)
        )

    def validator(self, addr):
        return json.loads(
            self.raw(
                "query",
                "staking",
                "validator",
                addr,
                output="json",
                node=self.node_rpc,
            )
        )

    def validators(self):
        return json.loads(
            self.raw(
                "query", "staking", "validators", output="json", node=self.node_rpc
            )
        )["validators"]

    def staking_params(self):
        return json.loads(
            self.raw("query", "staking", "params", output="json", node=self.node_rpc)
        )

    def total_minted_provision(self):
        return float(json.loads(
            self.raw("query", "mint", "total-minted-provision", output="json", node=self.node_rpc)
        )["amount"])

    def block_provisions(self):
        return json.loads(
            self.raw("query", "mint", "block-provision", output="json", node=self.node_rpc)
        )["amount"]

    def staking_pool(self, bonded=True):
        return int(
            json.loads(
                self.raw("query", "staking", "pool", output="json", node=self.node_rpc)
            )["bonded_tokens" if bonded else "not_bonded_tokens"]
        )

    def get_inflation_params(self, height=None):
        if height:
            return json.loads(
                self.raw("query", "inflation", "params", "--height", height, output="json", node=self.node_rpc)
            )
        return json.loads(
            self.raw("query", "inflation", "params", output="json", node=self.node_rpc)
        )

    def get_inflation_rate(self, height=None):
        if height:
            return json.loads(
                self.raw("query", "inflation", "inflation-rate", "--height", height, output="json", node=self.node_rpc)
            )
        return json.loads(
            self.raw("query", "inflation", "inflation-rate", output="json", node=self.node_rpc)
        )

    def get_epoch_mint_provision(self):
        res = self.raw("query", "inflation", "epoch-mint-provision", node=self.node_rpc).decode()
        res = res[:-6]
        if res[-1] == 'a':
            res = res[:-1]
        return int(float(res))

    def get_circulating_supply(self):
        res = self.raw("query", "inflation", "circulating-supply", node=self.node_rpc).decode()
        res = res[:-6]
        if res[-1] == 'a':
            res = res[:-1]
        return int(float(res))

    def get_inflation_epoch_identifier(self):
        inflation_period_info = self.get_inflation_period()
        return inflation_period_info["epoch_identifier"]

    def get_inflation_period(self):
        return json.loads(
            self.raw("query", "inflation", "period", output="json", node=self.node_rpc)
        )

    def get_epoch_infos(self):
        return json.loads(
            self.raw("query", "epochs", "epoch-infos", output="json", node=self.node_rpc)
        )

    def get_epoch_duration(self, epoch_identifier="day"):
        epoch_infos = self.get_epoch_infos()
        if epoch_infos["epochs"]:
            for epoch in epoch_infos["epochs"]:
                if epoch["identifier"] == epoch_identifier:
                    return int(epoch["duration"])
        return None

    def get_current_epoch(self, epoch_identifier="day"):
        res = json.loads(
            self.raw("query", "epochs", "current-epoch", epoch_identifier, output="json", node=self.node_rpc)
        )
        if res:
            return int(res["current_epoch"])

        return None

    def transfer(self, from_, to, coins, generate_only=False, fees=None, **kwargs):
        if fees is None:
            kwargs.setdefault("gas_prices", DEFAULT_GAS_PRICE)
        return json.loads(
            self.raw(
                "tx",
                "bank",
                "send",
                from_,
                to,
                coins,
                "-y",
                "--generate-only" if generate_only else None,
                home=self.data_dir,
                fees=fees,
                **kwargs,
            )
        )

    def get_delegated_amount(self, which_addr):
        return json.loads(
            self.raw(
                "query",
                "staking",
                "delegations",
                which_addr,
                home=self.data_dir,
                chain_id=self.chain_id,
                node=self.node_rpc,
                output="json",
            )
        )

    def delegate_amount(self, to_addr, amount, from_addr, gas_price=None):
        if gas_price is None:
            return json.loads(
                self.raw(
                    "tx",
                    "staking",
                    "delegate",
                    to_addr,
                    amount,
                    "-y",
                    home=self.data_dir,
                    from_=from_addr,
                    keyring_backend="test",
                    chain_id=self.chain_id,
                    node=self.node_rpc,
                )
            )
        else:
            return json.loads(
                self.raw(
                    "tx",
                    "staking",
                    "delegate",
                    to_addr,
                    amount,
                    "-y",
                    home=self.data_dir,
                    from_=from_addr,
                    keyring_backend="test",
                    chain_id=self.chain_id,
                    node=self.node_rpc,
                    gas_prices=gas_price,
                )
            )

    # to_addr: astraclcl1...  , from_addr: astra1...
    def unbond_amount(self, to_addr, amount, from_addr):

        return json.loads(
            self.raw(
                "tx",
                "staking",
                "unbond",
                to_addr,
                amount,
                "-y",
                home=self.data_dir,
                from_=from_addr,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
                gas=DEFAULT_GAS
            )
        )

    # to_validator_addr: astracncl1...  ,  from_from_validator_addraddr: astracl1...
    def redelegate_amount(
            self, to_validator_addr, from_validator_addr, amount, from_addr
    ):
        return json.loads(
            self.raw(
                "tx",
                "staking",
                "redelegate",
                from_validator_addr,
                to_validator_addr,
                amount,
                "-y",
                home=self.data_dir,
                from_=from_addr,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    # from_delegator can be account name or address
    def withdraw_all_rewards(self, from_delegator):
        return json.loads(
            self.raw(
                "tx",
                "distribution",
                "withdraw-all-rewards",
                "-y",
                from_=from_delegator,
                home=self.data_dir,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def make_multisig(self, name, signer1, signer2):
        self.raw(
            "keys",
            "add",
            name,
            multisig=f"{signer1},{signer2}",
            multisig_threshold="2",
            home=self.data_dir,
            keyring_backend="test",
        )

    def sign_multisig_tx(self, tx_file, multi_addr, signer_name):
        return json.loads(
            self.raw(
                "tx",
                "sign",
                tx_file,
                from_=signer_name,
                multisig=multi_addr,
                home=self.data_dir,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def sign_batch_multisig_tx(
            self, tx_file, multi_addr, signer_name, account_number, sequence_number
    ):
        r = self.raw(
            "tx",
            "sign-batch",
            "--offline",
            tx_file,
            account_number=account_number,
            sequence=sequence_number,
            from_=signer_name,
            multisig=multi_addr,
            home=self.data_dir,
            keyring_backend="test",
            chain_id=self.chain_id,
            node=self.node_rpc,
        )
        return r.decode("utf-8")

    def encode_signed_tx(self, signed_tx):
        return self.raw(
            "tx",
            "encode",
            signed_tx,
        )

    def sign_single_tx(self, tx_file, signer_name):
        return json.loads(
            self.raw(
                "tx",
                "sign",
                tx_file,
                from_=signer_name,
                home=self.data_dir,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def combine_multisig_tx(self, tx_file, multi_name, signer1_file, signer2_file):
        return json.loads(
            self.raw(
                "tx",
                "multisign",
                tx_file,
                multi_name,
                signer1_file,
                signer2_file,
                home=self.data_dir,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def combine_batch_multisig_tx(
            self, tx_file, multi_name, signer1_file, signer2_file
    ):
        r = self.raw(
            "tx",
            "multisign-batch",
            tx_file,
            multi_name,
            signer1_file,
            signer2_file,
            home=self.data_dir,
            keyring_backend="test",
            chain_id=self.chain_id,
            node=self.node_rpc,
        )
        return r.decode("utf-8")

    def broadcast_tx(self, tx_file, **kwargs):
        kwargs.setdefault("broadcast_mode", "block")
        kwargs.setdefault("output", "json")
        return json.loads(
            self.raw("tx", "broadcast", tx_file, node=self.node_rpc, **kwargs)
        )

    def broadcast_tx_json(self, tx, **kwargs):
        with tempfile.NamedTemporaryFile("w") as fp:
            json.dump(tx, fp)
            fp.flush()
            return self.broadcast_tx(fp.name)

    def unjail(self, addr):
        return json.loads(
            self.raw(
                "tx",
                "slashing",
                "unjail",
                "-y",
                from_=addr,
                home=self.data_dir,
                node=self.node_rpc,
                keyring_backend="test",
                chain_id=self.chain_id,
            )
        )

    def create_validator(
            self,
            amount,
            moniker=None,
            commission_max_change_rate="0.01",
            commission_rate="0.1",
            commission_max_rate="0.2",
            min_self_delegation="1",
            identity="",
            website="",
            security_contact="",
            details="",
    ):
        """MsgCreateValidator
        create the node with create_node before call this"""
        pubkey = (
                "'"
                + (
                    self.raw(
                        "tendermint",
                        "show-validator",
                        home=self.data_dir,
                    )
                        .strip()
                        .decode()
                )
                + "'"
        )

        return json.loads(
            self.raw(
                "tx",
                "staking",
                "create-validator",
                "-y",
                from_=self.address("validator"),
                amount=amount,
                pubkey=pubkey,
                min_self_delegation=min_self_delegation,
                # commision
                commission_rate=commission_rate,
                commission_max_rate=commission_max_rate,
                commission_max_change_rate=commission_max_change_rate,
                # description
                moniker=moniker,
                identity=identity,
                website=website,
                security_contact=security_contact,
                details=details,
                # basic
                home=self.data_dir,
                node=self.node_rpc,
                keyring_backend="test",
                chain_id=self.chain_id,
                gas=DEFAULT_GAS
            )
        )

    def edit_validator(
            self,
            commission_rate=None,
            moniker=None,
            identity=None,
            website=None,
            security_contact=None,
            details=None,
    ):
        """MsgEditValidator"""
        options = dict(
            commission_rate=commission_rate,
            # description
            identity=identity,
            website=website,
            security_contact=security_contact,
            details=details,
        )
        options["new-moniker"] = moniker
        return json.loads(
            self.raw(
                "tx",
                "staking",
                "edit-validator",
                "-y",
                from_=self.address("validator"),
                home=self.data_dir,
                node=self.node_rpc,
                keyring_backend="test",
                chain_id=self.chain_id,
                **{k: v for k, v in options.items() if v is not None},
            )
        )

    def gov_propose(self, proposer, kind, proposal, **kwargs):
        kwargs.setdefault("gas_prices", DEFAULT_GAS_PRICE)
        if kind == "software-upgrade":
            return json.loads(
                self.raw(
                    "tx",
                    "gov",
                    "submit-proposal",
                    kind,
                    proposal["name"],
                    "-y",
                    from_=proposer,
                    # content
                    title=proposal.get("title"),
                    description=proposal.get("description"),
                    upgrade_height=proposal.get("upgrade-height"),
                    upgrade_time=proposal.get("upgrade-time"),
                    upgrade_info=proposal.get("upgrade-info"),
                    deposit=proposal.get("deposit"),
                    # basic
                    home=self.data_dir,
                    **kwargs,
                )
            )
        elif kind == "cancel-software-upgrade":
            return json.loads(
                self.raw(
                    "tx",
                    "gov",
                    "submit-proposal",
                    kind,
                    "-y",
                    from_=proposer,
                    # content
                    title=proposal.get("title"),
                    description=proposal.get("description"),
                    deposit=proposal.get("deposit"),
                    # basic
                    home=self.data_dir,
                    **kwargs,
                )
            )
        else:
            with tempfile.NamedTemporaryFile("w") as fp:
                json.dump(proposal, fp)
                fp.flush()
                return json.loads(
                    self.raw(
                        "tx",
                        "gov",
                        "submit-proposal",
                        kind,
                        fp.name,
                        "-y",
                        from_=proposer,
                        # basic
                        home=self.data_dir,
                        **kwargs,
                    )
                )

    def gov_vote(self, voter, proposal_id, option, **kwargs):
        kwargs.setdefault("gas_prices", DEFAULT_GAS_PRICE)
        return json.loads(
            self.raw(
                "tx",
                "gov",
                "vote",
                proposal_id,
                option,
                "-y",
                from_=voter,
                home=self.data_dir,
                **kwargs,
            )
        )

    def gov_deposit(self, depositor, proposal_id, amount):
        return json.loads(
            self.raw(
                "tx",
                "gov",
                "deposit",
                proposal_id,
                amount,
                "-y",
                from_=depositor,
                home=self.data_dir,
                node=self.node_rpc,
                keyring_backend="test",
                chain_id=self.chain_id,
            )
        )

    def query_proposals(self, depositor=None, limit=None, status=None, voter=None):
        return json.loads(
            self.raw(
                "query",
                "gov",
                "proposals",
                depositor=depositor,
                count_total=limit,
                status=status,
                voter=voter,
                output="json",
                node=self.node_rpc,
            )
        )

    def query_proposal(self, proposal_id):
        return json.loads(
            self.raw(
                "query",
                "gov",
                "proposal",
                proposal_id,
                output="json",
                node=self.node_rpc,
            )
        )

    def query_tally(self, proposal_id):
        return json.loads(
            self.raw(
                "query",
                "gov",
                "tally",
                proposal_id,
                output="json",
                node=self.node_rpc,
            )
        )

    def ibc_transfer(
            self,
            from_,
            to,
            amount,
            channel,  # src channel
            target_version,  # chain version number of target chain
            i=0,
    ):
        return json.loads(
            self.raw(
                "tx",
                "ibc-transfer",
                "transfer",
                "transfer",  # src port
                channel,
                to,
                amount,
                "-y",
                # FIXME https://github.com/cosmos/cosmos-sdk/issues/8059
                "--absolute-timeouts",
                from_=from_,
                home=self.data_dir,
                node=self.node_rpc,
                keyring_backend="test",
                chain_id=self.chain_id,
                packet_timeout_height=f"{target_version}-10000000000",
                packet_timeout_timestamp=0,
            )
        )

    def export(self):
        return self.raw("export", home=self.data_dir)

    def unsaferesetall(self):
        return self.raw("unsafe-reset-all")

    def create_nft(self, from_addr, denomid, denomname, schema, fees):
        return json.loads(
            self.raw(
                "tx",
                "nft",
                "issue",
                denomid,
                "-y",
                fees=fees,
                name=denomname,
                schema=schema,
                home=self.data_dir,
                from_=from_addr,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def query_nft(self, denomid):
        return json.loads(
            self.raw(
                "query",
                "nft",
                "denom",
                denomid,
                output="json",
                home=self.data_dir,
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def query_denom_by_name(self, denomname):
        return json.loads(
            self.raw(
                "query",
                "nft",
                "denom-by-name",
                denomname,
                output="json",
                home=self.data_dir,
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def create_nft_token(self, from_addr, to_addr, denomid, tokenid, uri, fees):
        return json.loads(
            self.raw(
                "tx",
                "nft",
                "mint",
                denomid,
                tokenid,
                "-y",
                uri=uri,
                recipient=to_addr,
                home=self.data_dir,
                from_=from_addr,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def query_nft_token(self, denomid, tokenid):
        return json.loads(
            self.raw(
                "query",
                "nft",
                "token",
                denomid,
                tokenid,
                output="json",
                home=self.data_dir,
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def burn_nft_token(self, from_addr, denomid, tokenid):
        return json.loads(
            self.raw(
                "tx",
                "nft",
                "burn",
                denomid,
                tokenid,
                "-y",
                from_=from_addr,
                keyring_backend="test",
                home=self.data_dir,
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def edit_nft_token(self, from_addr, denomid, tokenid, newuri, newname):
        return json.loads(
            self.raw(
                "tx",
                "nft",
                "edit",
                denomid,
                tokenid,
                "-y",
                from_=from_addr,
                uri=newuri,
                name=newname,
                keyring_backend="test",
                home=self.data_dir,
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def transfer_nft_token(self, from_addr, to_addr, denomid, tokenid):
        return json.loads(
            self.raw(
                "tx",
                "nft",
                "transfer",
                to_addr,
                denomid,
                tokenid,
                "-y",
                from_=from_addr,
                keyring_backend="test",
                home=self.data_dir,
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def set_delegate_keys(self, val_addr, acc_addr, eth_addr, signature, **kwargs):
        """
        val_addr: astra validator address
        acc_addr: orchestrator's astra address
        eth_addr: orchestrator's ethereum address
        """
        kwargs.setdefault("gas_prices", DEFAULT_GAS_PRICE)
        return json.loads(
            self.raw(
                "tx",
                "gravity",
                "set-delegate-keys",
                val_addr,
                acc_addr,
                eth_addr,
                signature,
                "-y",
                home=self.data_dir,
                **kwargs,
            )
        )

    def query_gravity_params(self):
        return json.loads(self.raw("query", "gravity", "params", home=self.data_dir))

    def query_signer_set_txs(self):
        return json.loads(
            self.raw("query", "gravity", "signer-set-txs", home=self.data_dir)
        )

    def query_signer_set_tx(self, nonce):
        return json.loads(
            self.raw(
                "query", "gravity", "signer-set-tx", str(nonce), home=self.data_dir
            )
        )

    def query_latest_signer_set_tx(self):
        return json.loads(
            self.raw("query", "gravity", "latest-signer-set-tx", home=self.data_dir)
        )

    def send_to_ethereum(self, receiver, coins, fee, **kwargs):
        kwargs.setdefault("gas_prices", DEFAULT_GAS_PRICE)
        return json.loads(
            self.raw(
                "tx",
                "gravity",
                "send-to-ethereum",
                receiver,
                coins,
                fee,
                "-y",
                home=self.data_dir,
                **kwargs,
            )
        )

    def query_contract_by_denom(self, denom: str):
        "query contract by denom"
        return json.loads(
            self.raw(
                "query",
                "astra",
                "contract-by-denom",
                denom,
                home=self.data_dir,
            )
        )

    def gov_propose_token_mapping_change(self, denom, contract, **kwargs):
        kwargs.setdefault("gas_prices", DEFAULT_GAS_PRICE)
        return json.loads(
            self.raw(
                "tx",
                "gov",
                "submit-proposal",
                "token-mapping-change",
                denom,
                contract,
                "-y",
                home=self.data_dir,
                **kwargs,
            )
        )

    def update_token_mapping(self, denom, contract, **kwargs):
        kwargs.setdefault("gas_prices", DEFAULT_GAS_PRICE)
        return json.loads(
            self.raw(
                "tx",
                "astra",
                "update-token-mapping",
                denom,
                contract,
                "-y",
                home=self.data_dir,
                **kwargs,
            )
        )

    def build_evm_tx(self, raw_tx: str, **kwargs):
        return json.loads(
            self.raw(
                "tx",
                "evm",
                "raw",
                raw_tx,
                "-y",
                "--generate-only",
                home=self.data_dir,
                **kwargs,
            )
        )

    def transfer_tokens(self, from_, to, amount, **kwargs):
        return json.loads(
            self.raw(
                "tx",
                "astra",
                "transfer-tokens",
                from_,
                to,
                amount,
                "-y",
                home=self.data_dir,
                **kwargs,
            )
        )

    def start_node(self, i):
        subprocess.run(
            [
                sys.executable,
                "-msupervisor.supervisorctl",
                "-c",
                Path(self.data_dir / "../../") / SUPERVISOR_CONFIG_FILE,
                "start",
                "{}-node{}".format(self.chain_id, i),
            ]
        )

    def stop_node(self, i):
        subprocess.run(
            [
                sys.executable,
                "-msupervisor.supervisorctl",
                "-c",
                Path(self.data_dir / "../../") / SUPERVISOR_CONFIG_FILE,
                "stop",
                "{}-node{}".format(self.chain_id, i),
            ]
        )

    def copy_validator_key(self, from_node, to_node):
        "Copy the validtor file in from_node to to_node"
        from_key_file = "{}/node{}/config/priv_validator_key.json".format(
            Path(self.data_dir / "../"), from_node
        )
        to_key_file = "{}/node{}/config/priv_validator_key.json".format(
            Path(self.data_dir / "../"), to_node
        )
        with open(from_key_file, "r") as f:
            key = f.read()
        with open(to_key_file, "w") as f:
            f.write(key)

    def nodes_len(self):
        "find how many 'node{i}' sub-directories"
        data_path = Path(self.data_dir / "../")
        return len(
            [p for p in data_path.iterdir() if re.match(r"^node\d+$", p.name)]
        )

    def create_node(
            self,
            base_port=None,
            moniker=None,
            hostname="localhost",
            statesync=False,
            mnemonic=None,
    ):
        """create new node in the data directory,
        process information is written into supervisor config
        start it manually with supervisor commands
        :return: new node index and config
        """
        i = self.nodes_len()

        # default configs
        if base_port is None:
            # use the node0's base_port + i * 10 as default base port for new ndoe
            base_port = self.config["validators"][0]["base_port"] + i * 10
        if moniker is None:
            moniker = f"node{i}"

        # add config
        assert len(self.config["validators"]) == i
        self.config["validators"].append(
            {
                "base_port": base_port,
                "hostname": hostname,
                "moniker": moniker,
            }
        )
        (Path(self.data_dir).parent / "config.json").write_text(json.dumps(self.config))

        # init home directory
        self.init_new_node(self.config["validators"][i]["moniker"], i)
        home = self.home(i)
        (home / "config/genesis.json").unlink()
        (home / "config/genesis.json").symlink_to("../../genesis.json")
        (home / "config/client.toml").write_text(
            tomlkit.dumps(
                {
                    "chain-id": self.chain_id,
                    "keyring-backend": "test",
                    "output": "json",
                    "node": self.get_node_rpc(i),
                    "broadcast-mode": "block",
                }
            )
        )
        # use p2p peers from node0's config
        node0 = tomlkit.parse((self.data_dir / "../node0/config/config.toml").read_text())

        def custom_edit_tm(doc):
            if statesync:
                info = self.status()["SyncInfo"]
                doc["statesync"].update(
                    {
                        "enable": True,
                        "rpc_servers": ",".join(self.get_node_rpc(i) for i in range(2)),
                        "trust_height": int(info["latest_block_height"]),
                        "trust_hash": info["latest_block_hash"],
                        "temp_dir": str(Path(self.data_dir).parent),
                        "discovery_time": "5s",
                    }
                )

        edit_tm_cfg(
            home / "config/config.toml",
            base_port,
            node0["p2p"]["persistent_peers"],
            {},
            custom_edit=custom_edit_tm,
        )
        edit_app_cfg(home / "config/app.toml", base_port, {})

        # create validator account
        self.create_account_specific_node("validator", mnemonic, i)

        # add process config into supervisor
        path = Path(self.data_dir).parent / SUPERVISOR_CONFIG_FILE
        ini = configparser.RawConfigParser()
        ini.read_file(path.open())
        chain_id = self.chain_id
        prgname = f"{chain_id}-node{i}"
        section = f"program:{prgname}"
        ini.add_section(section)
        ini[section].update(
            dict(
                COMMON_PROG_OPTIONS,
                command=f"{self.cmd} start --home %(here)s/node{i}",
                autostart="false",
                stdout_logfile=f"%(here)s/node{i}.log",
            )
        )
        with path.open("w") as fp:
            ini.write(fp)
        self.reload_supervisor()
        return i


def edit_tm_cfg(path, base_port, peers, config, *, custom_edit=None):
    "field name changed after tendermint 0.35, support both flavours."
    doc = tomlkit.parse(open(path).read())
    doc["mode"] = "validator"
    # tendermint is start in process, not needed
    # doc['proxy_app'] = 'tcp://127.0.0.1:%d' % abci_port(base_port)
    rpc = doc["rpc"]
    rpc["laddr"] = "tcp://0.0.0.0:%d" % ports.rpc_port(base_port)
    rpc["pprof_laddr"] = rpc["pprof-laddr"] = "localhost:%d" % (
        ports.pprof_port(base_port),
    )
    rpc["timeout_broadcast_tx_commit"] = rpc["timeout-broadcast-tx-commit"] = "30s"
    rpc["grpc_laddr"] = rpc["grpc-laddr"] = "tcp://0.0.0.0:%d" % (
        ports.grpc_port_tx_only(base_port),
    )
    p2p = doc["p2p"]
    # p2p["use-legacy"] = True
    p2p["laddr"] = "tcp://0.0.0.0:%d" % ports.p2p_port(base_port)
    p2p["persistent_peers"] = p2p["persistent-peers"] = peers
    p2p["addr_book_strict"] = p2p["addr-book-strict"] = False
    p2p["allow_duplicate_ip"] = p2p["allow-duplicate-ip"] = True
    doc["consensus"]["timeout_commit"] = doc["consensus"]["timeout-commit"] = "1s"
    patch_toml_doc(doc, config)
    if custom_edit is not None:
        custom_edit(doc)
    open(path, "w").write(tomlkit.dumps(doc))


def edit_app_cfg(path, base_port, app_config):
    default_patch = {
        "api": {
            "enable": True,
            "swagger": True,
            "enable-unsafe-cors": True,
            "address": "tcp://0.0.0.0:%d" % ports.api_port(base_port),
        },
        "grpc": {
            "address": "0.0.0.0:%d" % ports.grpc_port(base_port),
        },
        "pruning": "nothing",
        "state-sync": {
            "snapshot-interval": 5,
            "snapshot-keep-recent": 10,
        },
        "minimum-gas-prices": "0basecro",
    }

    app_config = format_value(
        app_config,
        {
            "EVMRPC_PORT": ports.evmrpc_port(base_port),
            "EVMRPC_PORT_WS": ports.evmrpc_ws_port(base_port),
        },
    )

    doc = tomlkit.parse(open(path).read())
    doc["grpc-web"] = {}
    doc["grpc-web"]["address"] = "0.0.0.0:%d" % ports.grpc_web_port(base_port)
    patch_toml_doc(doc, jsonmerge.merge(default_patch, app_config))
    open(path, "w").write(tomlkit.dumps(doc))


def patch_toml_doc(doc, patch):
    for k, v in patch.items():
        if isinstance(v, dict):
            patch_toml_doc(doc[k], v)
        else:
            doc[k] = v


def format_value(v, ctx):
    if isinstance(v, str):
        return v.format(**ctx)
    elif isinstance(v, dict):
        return {k: format_value(vv, ctx) for k, vv in v.items()}
    else:
        return v
