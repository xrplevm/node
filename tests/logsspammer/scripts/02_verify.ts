import path = require("path");
import {ethers} from "hardhat";
import * as fs from "fs";
import {LogSpammer} from "../typechain-types";
import {hexlify, Signer} from "ethers";

const INIT_BLOCK = 16428781;

const LOCALNET_RPC_URL = "http://localhost:26657";
const DEVNET_RPC_URL = "http://cosmos.xrplevm.org:26657"

async function main() {
    for(let block = INIT_BLOCK; block < INIT_BLOCK + 20; block++) {
        const b = Number(block);
        checkBlock(b);
    }
}

async function checkBlock(block: number) {
    let res = "block not found";
    while (res == "block not found") {
        res = await requestBlock(block);
        if (res != "block not found") {
            console.log("new block", block);
            compareResWithFuture(block, res);
        }
    }
}
async function compareResWithFuture(height: number, pastRes: any) {
    await sleep(1000);
    const res = await requestBlock(height);
    if (pastRes !== res) {
        console.log("Different response for height: ", height);
    } else {
        console.log("Block ok: ", height);
    }
}

async function requestBlock(height: number) {
    const res = await fetch(`${LOCALNET_RPC_URL}/block_results?height=${height}`);
    if (res.status >= 300) {
        return "block not found";
    }
    const json = await res.json();
    return JSON.stringify(json);
}

async function sleep(ms: number) {
    return new Promise(res => setTimeout(res, ms));
}

main().catch((e) => {
    console.error(e);
    process.exit(1);
});
