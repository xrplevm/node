import path = require("path");
import {ethers} from "hardhat";
import * as fs from "fs";

export const LOCALNET_RPC_URL = "http://localhost:8545";
export const DEVNET_RPC_URL = "https://rpc.devnet.xrplevm.org";

export const DEVNET_PRIVATE_KEY = "";

const pathOutputJson = path.join(__dirname, "./deploy_output.json");

async function main() {
    const provider = new ethers.JsonRpcProvider(DEVNET_RPC_URL);
    const signer = new ethers.Wallet(DEVNET_PRIVATE_KEY, provider);

    console.log("Deploying LogSpammer with the account:", signer.address);
    const LogSpammerFactory = await ethers.getContractFactory("LogSpammer", signer);
    const logSpammer = (await LogSpammerFactory.deploy());
    await logSpammer.waitForDeployment();
    console.log("LogSpammer deployed at:", await logSpammer.getAddress());

    const outputJson = {
        logSpammerAddress: await logSpammer.getAddress(),
    };

    fs.writeFileSync(pathOutputJson, JSON.stringify(outputJson, null, 1));
}

main().catch((e) => {
    console.error(e);
    process.exit(1);
});
