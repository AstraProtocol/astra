#!/bin/bash
rm -rf .testnets/
killall screen

# start a testnet
./build/astrad testnet init-files --keyring-backend=test

# change staking denom to ubaby
cat .testnets/node0/astrad/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="aastra"' > .testnets/node0/astrad/config/tmp_genesis.json && mv .testnets/node0/astrad/config/tmp_genesis.json .testnets/node0/astrad/config/genesis.json

# update crisis variable to ubaby
cat .testnets/node0/astrad/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="aastra"' > .testnets/node0/astrad/config/tmp_genesis.json && mv .testnets/node0/astrad/config/tmp_genesis.json .testnets/node0/astrad/config/genesis.json

# udpate gov genesis
cat .testnets/node0/astrad/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="aastra"' > .testnets/node0/astrad/config/tmp_genesis.json && mv .testnets/node0/astrad/config/tmp_genesis.json .testnets/node0/astrad/config/genesis.json

# update mint genesis
cat .testnets/node0/astrad/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="aastra"' > .testnets/node0/astrad/config/tmp_genesis.json && mv .testnets/node0/astrad/config/tmp_genesis.json .testnets/node0/astrad/config/genesis.json

cat .testnets/node0/astrad/config/genesis.json | jq '.app_state["slashing"]["params"]["signed_blocks_window"]="15"' > .testnets/node0/astrad/config/tmp_genesis.json && mv .testnets/node0/astrad/config/tmp_genesis.json .testnets/node0/astrad/config/genesis.json

# change app.toml values

# validator 1
sed -i -E 's|swagger = false|swagger = true|g' .testnets/node0/astrad/config/app.toml
sed -i -E 's|enabled-unsafe-cors = false|enabled-unsafe-cors = true|g' .testnets/node0/astrad/config/app.toml


# validator2
sed -i -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1316|g' .testnets/node1/astrad/config/app.toml
sed -i -E 's|0.0.0.0:9090|0.0.0.0:9088|g' .testnets/node1/astrad/config/app.toml
sed -i -E 's|0.0.0.0:9091|0.0.0.0:9089|g' .testnets/node1/astrad/config/app.toml
sed -i -E 's|0.0.0.0:8545|0.0.0.0:8547|g' .testnets/node1/astrad/config/app.toml
sed -i -E 's|0.0.0.0:8546|0.0.0.0:8548|g' .testnets/node1/astrad/config/app.toml
sed -i -E 's|swagger = false|swagger = true|g' .testnets/node1/astrad/config/app.toml

# validator3
sed -i -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1315|g' .testnets/node2/astrad/config/app.toml
sed -i -E 's|0.0.0.0:9090|0.0.0.0:9086|g' .testnets/node2/astrad/config/app.toml
sed -i -E 's|0.0.0.0:9091|0.0.0.0:9087|g' .testnets/node2/astrad/config/app.toml
sed -i -E 's|0.0.0.0:8545|0.0.0.0:8549|g' .testnets/node2/astrad/config/app.toml
sed -i -E 's|0.0.0.0:8546|0.0.0.0:8550|g' .testnets/node2/astrad/config/app.toml
sed -i -E 's|swagger = false|swagger = true|g' .testnets/node2/astrad/config/app.toml

# validator4
sed -i -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1314|g' .testnets/node3/astrad/config/app.toml
sed -i -E 's|0.0.0.0:9090|0.0.0.0:9084|g' .testnets/node3/astrad/config/app.toml
sed -i -E 's|0.0.0.0:9091|0.0.0.0:9085|g' .testnets/node3/astrad/config/app.toml
sed -i -E 's|0.0.0.0:8545|0.0.0.0:8551|g' .testnets/node3/astrad/config/app.toml
sed -i -E 's|0.0.0.0:8546|0.0.0.0:8552|g' .testnets/node3/astrad/config/app.toml
sed -i -E 's|swagger = false|swagger = true|g' .testnets/node3/astrad/config/app.toml

# change config.toml values

# validator1
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' .testnets/node0/astrad/config/config.toml
sed -i -E 's|tcp://0.0.0.0:40003|tcp://0.0.0.0:40000|g' .testnets/node0/astrad/config/config.toml
sed -i -E 's|tcp://0.0.0.0:50003|tcp://0.0.0.0:50000|g' .testnets/node0/astrad/config/config.toml
# validator2
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26655|g' .testnets/node1/astrad/config/config.toml
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' .testnets/node1/astrad/config/config.toml
# validator3
sed -i -E 's|tcp://0.0.0.0:40003|tcp://0.0.0.0:40001|g' .testnets/node1/astrad/config/config.toml
sed -i -E 's|tcp://0.0.0.0:50003|tcp://0.0.0.0:50001|g' .testnets/node1/astrad/config/config.toml

sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26652|g' .testnets/node2/astrad/config/config.toml
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' .testnets/node2/astrad/config/config.toml
sed -i -E 's|tcp://0.0.0.0:40003|tcp://0.0.0.0:40002|g' .testnets/node2/astrad/config/config.toml
sed -i -E 's|tcp://0.0.0.0:50003|tcp://0.0.0.0:50002|g' .testnets/node2/astrad/config/config.toml

# validator 4
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26651|g' .testnets/node3/astrad/config/config.toml
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' .testnets/node3/astrad/config/config.toml

sed -i -E 's|timeout_commit = "5s"|timeout_commit = "1s"|g' .testnets/node0/astrad/config/config.toml
sed -i -E 's|timeout_commit = "5s"|timeout_commit = "1s"|g' .testnets/node1/astrad/config/config.toml
sed -i -E 's|timeout_commit = "5s"|timeout_commit = "1s"|g' .testnets/node2/astrad/config/config.toml
sed -i -E 's|timeout_commit = "5s"|timeout_commit = "1s"|g' .testnets/node3/astrad/config/config.toml

# copy validator1 genesis file to validator2-3
cp .testnets/node0/astrad/config/genesis.json .testnets/node1/astrad/config/genesis.json
cp .testnets/node0/astrad/config/genesis.json .testnets/node2/astrad/config/genesis.json
cp .testnets/node0/astrad/config/genesis.json .testnets/node3/astrad/config/genesis.json
rm .testnets/node3/astrad/config/priv_validator_key.json

echo "start all four validators"
#screen -S validator1 -d -m astrad start --home=.testnets/node0/astrad
#screen -S validator2 -d -m astrad start --home=.testnets/node1/astrad
# screen -S validator3 -d -m astrad start --home=.testnets/node2/astrad

#echo $(astrad keys show node0 -a --keyring-backend=test --home=.testnets/node0/astrad)
#echo $(astrad keys show node1 -a --keyring-backend=test --home=.testnets/node1/astrad)
#echo $(astrad keys show node2 -a --keyring-backend=test --home=.testnets/node2/astrad)