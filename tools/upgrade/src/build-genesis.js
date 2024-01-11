const fs = require("fs");

const VOTING_POWER = 1;
const ADDRESS = "ethm1lntwj7nnpsg55mv0635ycr9vzke0syyg68jp2l";
const EVM_ADDRESS = "0xfCD6E97a730C114a6D8fd4684c0CaC15b2F81088";
const PUB_KEY = {
    "@type": "/ethermint.crypto.v1.ethsecp256k1.PubKey",
    "key": "Al8vAj8DfDp+X4Z2gJHNDisFYffp7U01+0ZhSC+9blLt"
};
const VALIDATOR_ADDRESS = "ethmvaloper1lntwj7nnpsg55mv0635ycr9vzke0syyg4hcdjz";
const VALIDATOR_CONS_ADDRESS = "ethmvalcons1l09acuuu7zv7n7zp4npwqdjn3ancpjfegtmjl6";
const VALIDATOR_CONS_PUB_KEY = {
    "@type": "/cosmos.crypto.ed25519.PubKey",
    "key": "vf/n1Yy7GYZ6IymjfjEhJL0SrRqBF5Ns3dhAxHXAkQU="
};
const VALIDATOR_ID = "FBCBDC739CF099E9F841ACC2E036538F6780C939";
const XRP_SUPPLY = "99999998174940445117533027764";
//"1000000000000000000000"

const GOV_VOTING_PERIOD = "60s";
const GOV_MIN_DEPOSIT = {
    amount: "1",
    denom: "axrp",
};

const DEFAULT_POWER_REDUCTION = 10**6;
const STAKING_BONDED_POOL_ADDRESS = "ethm1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3w48d64";
const DISTRIBUTION_MODULE_ADDRESS = "ethm1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8u3272a";

const gs = JSON.parse(fs.readFileSync(process.argv[2]).toString());

/**
 * GOV
 */
gs.app_state.gov.deposit_params.min_deposit = [GOV_MIN_DEPOSIT];
gs.app_state.gov.voting_params.voting_period = GOV_VOTING_PERIOD;

/**
 * AUTH
 */
gs.app_state.auth.accounts.push({
    "@type": "/ethermint.types.v1.EthAccount",
    "base_account": {
        "account_number": "0",
        "address": ADDRESS,
        "pub_key": PUB_KEY,
        "sequence": "1"
    },
    "code_hash": "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
});

/**
 * BANK
 */
for (const bankEntry of gs.app_state.bank.balances) {
    if (bankEntry.address === STAKING_BONDED_POOL_ADDRESS) {
        bankEntry.coins[0].amount = (VOTING_POWER * DEFAULT_POWER_REDUCTION).toString();
    } else if (bankEntry.address === DISTRIBUTION_MODULE_ADDRESS) {
        bankEntry.coins[0].amount = "0";
    } else {
        for (const coin of bankEntry.coins) {
            if (coin.denom === "apoa") {
                coin.amount = "0";
            }
        }
    }
}

gs.app_state.bank.balances.push({
    address: ADDRESS,
    coins: [{
        amount: "1000000000000000000000",
        denom: "axrp",
    }]
})
for (const bankEntry of gs.app_state.bank.supply) {
    if (bankEntry.denom === "apoa") {
        bankEntry.amount = (VOTING_POWER * DEFAULT_POWER_REDUCTION).toString();
    }
    if (bankEntry.denom === "axrp") {
        bankEntry.amount = XRP_SUPPLY;

    }
}

/**
 * DISTRIBUTION
 */
gs.app_state.distribution.fee_pool.community_pool = [];
gs.app_state.distribution.delegator_starting_infos = [{
    "delegator_address": ADDRESS,
    "starting_info": {
        "height": "0",
        "previous_period": "1",
        "stake": `${VOTING_POWER * DEFAULT_POWER_REDUCTION}.000000000000000000`
    },
    "validator_address": VALIDATOR_ADDRESS
}];
gs.app_state.distribution.outstanding_rewards = [{
    "outstanding_rewards": [],
    "validator_address": VALIDATOR_ADDRESS,
}];
gs.app_state.distribution.validator_accumulated_commissions = [{
    "accumulated": {
        "commission": []
    },
    "validator_address": VALIDATOR_ADDRESS,
}];
gs.app_state.distribution.validator_current_rewards = [{
    "rewards": {
        "period": "2",
        "rewards": []
    },
    "validator_address": VALIDATOR_ADDRESS,
}];
gs.app_state.distribution.validator_historical_rewards = [{
    "period": "1",
    "rewards": {
        "cumulative_reward_ratio": [],
        "reference_count": 11
    },
    "validator_address": VALIDATOR_ADDRESS,
}];

/**
 * EVM
 */
gs.app_state.evm.accounts.push({
    "address": EVM_ADDRESS,
    "code": "",
    "storage": []
})

/**
 * SLASHING
 */
gs.app_state.slashing.missed_blocks = [{
    "address": VALIDATOR_CONS_ADDRESS,
    "missed_blocks": []
}];
gs.app_state.slashing.signing_infos = [{
    "address": VALIDATOR_CONS_ADDRESS,
    "validator_signing_info": {
        "address": VALIDATOR_CONS_ADDRESS,
        "index_offset": "2",
        "jailed_until": "1970-01-01T00:00:00Z",
        "missed_blocks_counter": "0",
        "start_height": "0",
        "tombstoned": false
    }
}];

/**
 * STAKING
 */
gs.app_state.staking.delegations = [{
    "delegator_address": ADDRESS,
    "shares": `${VOTING_POWER * DEFAULT_POWER_REDUCTION}.000000000000000000`,
    "validator_address": VALIDATOR_ADDRESS,
}];
gs.app_state.staking.last_total_power = (VOTING_POWER).toString();
gs.app_state.staking.last_validator_powers = [{
    "address": VALIDATOR_ADDRESS,
    "power": VOTING_POWER.toString(),
}];
gs.app_state.staking.validators = [{
    "commission": {
        "commission_rates": {
            "max_change_rate": "0.010000000000000000",
            "max_rate": "0.200000000000000000",
            "rate": "0.100000000000000000"
        },
        "update_time": "2024-01-10T16:51:51.626423Z"
    },
    "consensus_pubkey": VALIDATOR_CONS_PUB_KEY,
    "delegator_shares": `${VOTING_POWER * DEFAULT_POWER_REDUCTION}.000000000000000000`,
    "description": {
        "details": "",
        "identity": "",
        "moniker": "mynode",
        "security_contact": "",
        "website": ""
    },
    "jailed": false,
    "min_self_delegation": VOTING_POWER.toString(),
    "operator_address": VALIDATOR_ADDRESS,
    "status": "BOND_STATUS_BONDED",
    "tokens": `${VOTING_POWER * DEFAULT_POWER_REDUCTION}`,
    "unbonding_height": "0",
    "unbonding_time": "1970-01-01T00:00:00Z"
}];

/**
 * GENERAL
 */
gs.validators = [{
    "address": VALIDATOR_ID,
    "name": "mynode",
    "power": VOTING_POWER.toString(),
    "pub_key": {
        "type": "tendermint/PubKeyEd25519",
        "value": VALIDATOR_CONS_PUB_KEY.key
    },
}];

fs.writeFileSync(process.argv[3], JSON.stringify(gs));
