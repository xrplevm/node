import * as path from "path";
import {ethers} from "hardhat";
import * as fs from "fs";
import {LogSpammer} from "../typechain-types";
import {Provider, Signer} from "ethers";
import {log} from "node:util";

const pathOutputJson = path.join(__dirname, "./deploy_output.json");

const TXS = 10_000;

const FOO = 12345;
const BAR = 54321;

async function spam(logSpammer: LogSpammer, signer: Signer) {
    console.log("Spamming logs...");
    let nonce = await signer.getNonce();
    for (let i = 0; i < TXS; i++) {
        sendSpamTx(logSpammer, nonce)
        await sleep(200);
        nonce++;
    }
}

async function sendSpamTx(logSpammer: LogSpammer, nonce: number) {
    const n = randomIntFromInterval(10, 20);
    let data = new Uint8Array(randomIntFromInterval(32, 32 * 4));
    data = crypto.getRandomValues(data);
    try {
        const res = await logSpammer.spam(n, FOO, BAR, data, {
            nonce: nonce,
        })
        console.log(`Spammed ${res.nonce} - ${res.hash}`);
    } catch (err) {
        await sleep(200);
        sendSpamTx(logSpammer, nonce);
    }
}

async function noise(address: string, provider: Provider) {
    console.log("initiating noise");
    while (true) {
        provider.getCode(address).then(console.log)
        provider.getBalance(address).then(console.log)
        provider.getBlockNumber().then(console.log)
        provider.getLogs({ address: address }).then(console.log)
        await sleep(1)
    }
}

async function main() {
    const {logSpammerAddress} = JSON.parse(fs.readFileSync(pathOutputJson).toString());


    const provider = new ethers.JsonRpcProvider("http://78.47.240.16:8545");
    // const provider = new ethers.JsonRpcProvider("https://rpc.devnet.xrplevm.org");
    // const provider = new ethers.JsonRpcProvider("http://65.108.250.207:8545");
    const signer = new ethers.Wallet("", provider);
    const LogSpammerFactory = await ethers.getContractFactory("LogSpammer", signer);
    const logSpammer = LogSpammerFactory.attach(logSpammerAddress) as unknown as LogSpammer;

    // noise(logSpammerAddress, provider);

    while(true) {
        try {
            await spam(logSpammer, signer);
        } catch (e) {
            console.log("cathced");
            console.log(e);
        }
    }
}


function randomIntFromInterval(min: number, max: number) {
    return Math.floor(Math.random() * (max - min + 1) + min);
}

async function sleep(ms: number) {
    return new Promise(res => setTimeout(res, ms));
}

main().catch((e) => {
    console.error(e);
    process.exit(1);
});
