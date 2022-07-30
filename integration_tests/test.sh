#!/bin/sh
pytest -m normal -vv
pytest -m gov -vv
pytest -m authz -vv
pytest -m authz_execute -vv
pytest -m staking -vv
pytest -m vesting -vv
pytest -m byzantine -vv
pytest -m eip1559 -vv
pytest -m reward -vv