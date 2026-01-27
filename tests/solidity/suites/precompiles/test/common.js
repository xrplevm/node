// Common constants and helper utilities for precompile tests

const STAKING_PRECOMPILE_ADDRESS = '0x0000000000000000000000000000000000000800'
const BECH32_PRECOMPILE_ADDRESS = '0x0000000000000000000000000000000000000400'
const DISTRIBUTION_PRECOMPILE_ADDRESS = '0x0000000000000000000000000000000000000801'
const BANK_PRECOMPILE_ADDRESS = '0x0000000000000000000000000000000000000804'
const GOV_PRECOMPILE_ADDRESS = '0x0000000000000000000000000000000000000805'
const SLASHING_PRECOMPILE_ADDRESS = '0x0000000000000000000000000000000000000806'
const P256_PRECOMPILE_ADDRESS = '0x0000000000000000000000000000000000000100'
const WERC20_ADDRESS = '0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE'

// Default gas limits used across tests
const DEFAULT_GAS_LIMIT = 1_000_000
const LARGE_GAS_LIMIT = 10_000_000

const RETRY_DELAY_FUNC  = (attempt) => 500 * Math.pow(2, attempt)


function waitWithTimeout(txn, timeoutMs, retryDelayFn = (attempt) => 1000) {
    return new Promise((resolve, reject) => {
        const deadlineTimer = setTimeout(() => {
            reject(new Error(`Txn wait failed after timeout of ${timeoutMs}ms.`));
        }, timeoutMs);

        let attempt = 0;

        function attemptCall() {
            Promise.resolve()
                .then(() => txn.wait())
                .then((result) => {
                    clearTimeout(deadlineTimer);
                    resolve(result);
                })
                .catch(() => {
                    attempt++;
                    const delay = retryDelayFn(attempt);
                    setTimeout(attemptCall, delay);
                });
        }

        attemptCall();
    });
}

// Helper to convert the raw tuple returned by staking.validator() into an object
function parseValidator (raw) {
    return {
        operatorAddress: raw[0],
        consensusPubkey: raw[1],
        jailed: raw[2],
        status: raw[3],
        tokens: raw[4],
        delegatorShares: raw[5],
        description: raw[6],
        unbondingHeight: raw[7],
        unbondingTime: raw[8],
        commission: raw[9],
        minSelfDelegation: raw[10]
    }
}

// Utility to parse logs and return the first matching event by name
function findEvent (logs, iface, eventName) {
    for (const log of logs) {
        try {
            const parsed = iface.parseLog(log)
            if (parsed && parsed.name === eventName) {
                return parsed
            }
        } catch {
            // ignore logs that do not match the contract interface
        }
    }
    return null
}

module.exports = {
    STAKING_PRECOMPILE_ADDRESS,
    BECH32_PRECOMPILE_ADDRESS,
    DISTRIBUTION_PRECOMPILE_ADDRESS,
    BANK_PRECOMPILE_ADDRESS,
    GOV_PRECOMPILE_ADDRESS,
    SLASHING_PRECOMPILE_ADDRESS,
    P256_PRECOMPILE_ADDRESS,
    WERC20_ADDRESS,
    DEFAULT_GAS_LIMIT,
    LARGE_GAS_LIMIT,
    RETRY_DELAY_FUNC,
    parseValidator,
    findEvent,
    waitWithTimeout
}