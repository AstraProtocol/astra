#!/bin/sh
pytest -m normal -vv
pytest -m gov -vv
pytest -m authz -vv
pytest -m staking -vv
pytest -m vesting -vv
pytest -m byzantine -vv 