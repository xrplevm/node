const { expect } = require('chai')
const hre = require('hardhat')
const { findEvent, waitWithTimeout, RETRY_DELAY_FUNC} = require('../common')

describe('ERC20 Precompile', function () {
    let erc20, erc20BurnContract, erc20Burn0Contract, owner, spender, recipient
    const GAS_LIMIT = 1_000_000 // skip gas estimation for simplicity

    const ERC20_PRECOMPILE_ADDRESS = '0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE'

    const ERC20_BURN_ABI = [
        'function mint(address to, uint256 amount) external returns (bool)',
        'function burn(uint256 amount) external',
    ]

    // Get the ERC20 precompile contract instance
    const ERC20_BURN0_ABI = [
        'function mint(address to, uint256 amount) external returns (bool)',
        'function burn(address from, uint256 amount) external',
    ]

    before(async function () {
        [owner, spender, recipient] = await hre.ethers.getSigners()
        erc20 = await hre.ethers.getContractAt(
            'IERC20Metadata',
            '0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE'
        )

        erc20BurnContract = new hre.ethers.Contract(ERC20_PRECOMPILE_ADDRESS, ERC20_BURN_ABI, owner)
        erc20Burn0Contract = new hre.ethers.Contract(ERC20_PRECOMPILE_ADDRESS, ERC20_BURN0_ABI, owner)
    })

    it('should return the name', async function () {
        const name = await erc20.name()
        expect(name).to.contain('Token')
    })

    it('should return the symbol', async function () {
        const symbol = await erc20.symbol()
        expect(symbol).to.contain('TOKEN')
    })

    it('should return the decimals', async function () {
        const decimals = await erc20.decimals()
        expect(decimals).to.equal(18)
    })

    it('should return the total supply', async function () {
        const totalSupply = await erc20.totalSupply()
        expect(totalSupply).to.be.gt(0n)
    })

    it('should return the balance of the owner', async function () {
        const balance = await erc20.balanceOf(owner.address)
        expect(balance).to.be.gt(0n)
    })

    it('should return zero allowance by default', async function () {
        const allowance = await erc20.allowance(owner.address, spender.address)
        expect(allowance).to.equal(0n)
    })


    it('should return the contract owner address', async function () {
        const ownerAddr = await erc20.owner()
        expect(ownerAddr).to.equal(owner.address)
    })

    it('should transfer tokens', async function () {
        const amount = hre.ethers.parseEther('1')
        const prev   = await erc20.balanceOf(spender.address)

        const tx = await erc20.connect(owner).transfer(spender.address, amount)
        const receipt = await waitWithTimeout(tx, 20000, RETRY_DELAY_FUNC)

        const transferEvent = findEvent(receipt.logs, erc20.interface, 'Transfer')
        expect(transferEvent, 'Transfer event must be emitted').to.exist
        expect(transferEvent.args.from).to.equal(owner.address)
        expect(transferEvent.args.to).to.equal(spender.address)
        expect(transferEvent.args.value).to.equal(amount)

        const after = await erc20.balanceOf(spender.address)
        expect(after - prev).to.equal(amount)
    })

    it('should transfer tokens using transferFrom', async function () {
        const amount = hre.ethers.parseEther('0.5')

        // owner gives spender permission to move amount
        const approvalTx = await erc20.
            connect(owner)
            .approve(spender.address, amount, {gasLimit: GAS_LIMIT})
        const approvalReceipt = await waitWithTimeout(approvalTx, 20000, RETRY_DELAY_FUNC)


        const approvalEvent = findEvent(approvalReceipt.logs, erc20.interface, 'Approval')
        expect(approvalEvent, 'Approval event must be emitted').to.exist
        expect(approvalEvent.args.owner).to.equal(owner.address)
        expect(approvalEvent.args.spender).to.equal(spender.address)
        expect(approvalEvent.args.value).to.equal(amount)

        // record pre-transfer balances and allowance
        const prevBalance    = await erc20.balanceOf(recipient.address)
        const prevAllowance  = await erc20.allowance(owner.address, spender.address)

        // spender pulls from owner → recipient
        const tx = await erc20
            .connect(spender)
            .transferFrom(owner.address, recipient.address, amount, {gasLimit: GAS_LIMIT})
        const receipt = await waitWithTimeout(tx, 20000, RETRY_DELAY_FUNC)

        const transferEvent = findEvent(receipt.logs, erc20.interface, 'Transfer')
        expect(transferEvent, 'Transfer event must be emitted').to.exist
        expect(transferEvent.args.from).to.equal(owner.address)
        expect(transferEvent.args.to).to.equal(recipient.address)
        expect(transferEvent.args.value).to.equal(amount)

        // post-transfer checks
        const afterBalance   = await erc20.balanceOf(recipient.address)
        const afterAllowance = await erc20.allowance(owner.address, spender.address)

        // recipient should gain exactly `amount`
        expect(afterBalance - prevBalance).to.equal(amount)

        // allowance should have decreased by `amount`
        expect(afterAllowance).to.equal(prevAllowance - amount)
    })


    describe('mint', function () {
        it('should revert if the caller is not the contract owner', async function () {
            const mintAmount = hre.ethers.parseEther('100')

            // Connect as spender (non-owner) and mint - this should revert
            const spenderContract = erc20.connect(spender)

            // Mint tokens as non-owner spender to recipient - should revert
            await expect(spenderContract.mint(recipient.address, mintAmount))
                .to.be.reverted
        })

        it('should mint tokens to the recipient if the caller is the contract owner', async function () {
            const mintAmount = hre.ethers.parseEther('100')

            // Connect as owner and mint
            const contractOwner = erc20.connect(owner)

            // Get initial balance
            const initialBalance = await erc20.balanceOf(recipient.address)

            // Mint tokens as owner
            const tx = await contractOwner.mint(recipient.address, mintAmount)
            const receipt = await waitWithTimeout(tx, 20000, RETRY_DELAY_FUNC)

            expect(tx).to.not.be.reverted

            // Check Transfer event was emitted
            const transferEvent = findEvent(receipt.logs, erc20.interface, 'Transfer')
            expect(transferEvent, 'Transfer event must be emitted').to.exist
            expect(transferEvent.args.from).to.equal('0x0000000000000000000000000000000000000000') // Zero address for minting
            expect(transferEvent.args.to).to.equal(recipient.address)
            expect(transferEvent.args.value).to.equal(mintAmount)

            // Check new balance
            const newBalance = await erc20.balanceOf(recipient.address)
            expect(newBalance).to.equal(initialBalance + mintAmount)
        })
    })

    describe('burn', function () {
        it('should burn tokens from the caller', async function () {
            const mintAmount = hre.ethers.parseEther('100')
            const burnAmount = hre.ethers.parseEther('50')

            // First mint some tokens to owner (use owner for this test to avoid conflicts)
            const mintTx = await erc20BurnContract.mint(spender.address, mintAmount)
            const mintReceipt = await waitWithTimeout(mintTx, 20000, RETRY_DELAY_FUNC)

            const spenderContract = new hre.ethers.Contract(ERC20_PRECOMPILE_ADDRESS, ERC20_BURN_ABI, spender)
            // Get initial balance
            const initialBalance = await erc20.balanceOf(spender.address)

            // Owner burns their own tokens
            const burnTx = await spenderContract.burn(burnAmount, {gasPrice: 0})
            const burnReceipt = await waitWithTimeout(burnTx, 20000, RETRY_DELAY_FUNC)

            // Check Transfer event was emitted
            const transferEvent = findEvent(burnReceipt.logs, erc20.interface, 'Transfer')
            expect(transferEvent, 'Transfer event must be emitted').to.exist
            expect(transferEvent.args.from).to.equal(spender.address)
            expect(transferEvent.args.to).to.equal('0x0000000000000000000000000000000000000000') // Zero address for burning
            expect(transferEvent.args.value).to.equal(burnAmount)

            // Check new balance
            const newBalance = await erc20.balanceOf(spender.address)
            expect(newBalance).to.equal(initialBalance - burnAmount)
        })
    })

    describe('burn0', function () {
        it('should revert if the caller is not the contract owner', async function () {
            const burnAmount = hre.ethers.parseEther('10')

            // Connect as spender (non-owner) and attempt to burn from recipient - this should revert
            const contractAsSpender = new hre.ethers.Contract(ERC20_PRECOMPILE_ADDRESS, ERC20_BURN0_ABI, spender)

            // Attempt to burn tokens from recipient as non-owner spender - should revert
            await expect(contractAsSpender.burn(recipient.address, burnAmount))
                .to.be.reverted
        })

        it('should allow owner to burn tokens from any address', async function () {
            const mintAmount = hre.ethers.parseEther('100')
            const burnAmount = hre.ethers.parseEther('30')

            // First mint some tokens to spender
            const mintTx = await erc20Burn0Contract.mint(spender.address, mintAmount)
            const mintReceipt = await waitWithTimeout(mintTx, 20000, RETRY_DELAY_FUNC)

            // Get initial balance of spender
            const initialBalance = await erc20.balanceOf(spender.address)

            // Owner burns tokens from spender's account
            const burnTx = await erc20Burn0Contract.burn(spender.address, burnAmount)
            const burnReceipt = await waitWithTimeout(burnTx, 20000, RETRY_DELAY_FUNC)

            expect(burnTx).to.not.be.reverted

            // Check Transfer event was emitted
            const transferEvent = findEvent(burnReceipt.logs, erc20.interface, 'Transfer')
            expect(transferEvent, 'Transfer event must be emitted').to.exist
            expect(transferEvent.args.from).to.equal(spender.address)
            expect(transferEvent.args.to).to.equal('0x0000000000000000000000000000000000000000') // Zero address for burning
            expect(transferEvent.args.value).to.equal(burnAmount)

            // Check new balance
            const newBalance = await erc20.balanceOf(spender.address)
            expect(newBalance).to.equal(initialBalance - burnAmount)
        })

        it('should revert when trying to burn more than available balance', async function () {
            // Get current balance of spender
            const currentBalance = await erc20.balanceOf(spender.address)
            const burnAmount = currentBalance + hre.ethers.parseEther('1') // Try to burn more than available

            // Owner attempts to burn more tokens than spender has - should revert
            await expect(erc20Burn0Contract.burn(spender.address, burnAmount))
                .to.be.reverted
        })
    })

    describe('burnFrom', function () {
        it('should allow any caller to burn from account with allowance', async function () {
            const mintAmount = hre.ethers.parseEther('100')
            const burnAmount = hre.ethers.parseEther('50')

            // First mint some tokens to spender
            const mintTx = await erc20.connect(owner).mint(spender.address, mintAmount)
            const mintReceipt = await waitWithTimeout(mintTx, 20000, RETRY_DELAY_FUNC)

            // Spender approves recipient to spend tokens
            const contractAsSpender = erc20.connect(spender)
            const approveTx = await contractAsSpender.approve(recipient.address, burnAmount)
            const approveReceipt = await waitWithTimeout(approveTx, 20000, RETRY_DELAY_FUNC)

            // Get initial balance and allowance
            const initialBalance = await erc20.balanceOf(spender.address)
            const initialAllowance = await erc20.allowance(spender.address, recipient.address)

            // Connect as recipient (non-owner) and burnFrom - this should succeed with allowance
            const contractAsRecipient = erc20.connect(recipient)

            const burnFromTx = await contractAsRecipient.burnFrom(spender.address, burnAmount)
            const burnFromReceipt = await waitWithTimeout(burnFromTx, 20000, RETRY_DELAY_FUNC)

            // Check Transfer event was emitted
            const transferEvent = findEvent(burnFromReceipt.logs, erc20.interface, 'Transfer')
            expect(transferEvent, 'Transfer event must be emitted').to.exist
            expect(transferEvent.args.from).to.equal(spender.address)
            expect(transferEvent.args.to).to.equal('0x0000000000000000000000000000000000000000') // Zero address for burning
            expect(transferEvent.args.value).to.equal(burnAmount)

            // Check new balance
            const newBalance = await erc20.balanceOf(spender.address)
            expect(newBalance).to.equal(initialBalance - burnAmount)

            // Check allowance was reduced
            const newAllowance = await erc20.allowance(spender.address, recipient.address)
            expect(newAllowance).to.equal(0)
        })

        it('should burn tokens from the specified account with allowance', async function () {
            const mintAmount = hre.ethers.parseEther('200') // Use different amount to avoid conflicts
            const burnAmount = hre.ethers.parseEther('75')  // Use different amount to avoid conflicts

            // First mint some tokens to recipient (use recipient for this test)
            const mintTx = await erc20.connect(owner).mint(recipient.address, mintAmount)
            const mintReceipt = await waitWithTimeout(mintTx, 20000, RETRY_DELAY_FUNC)

            // Recipient approves spender to spend tokens (different direction than first burnFrom test)
            const contractAsRecipient = erc20.connect(recipient)
            const approveTx = await contractAsRecipient.approve(spender.address, burnAmount)
            const approveReceipt = await waitWithTimeout(approveTx, 20000, RETRY_DELAY_FUNC)

            // Get initial balance and allowance
            const initialBalance = await erc20.balanceOf(recipient.address)
            const initialAllowance = await erc20.allowance(recipient.address, spender.address)

            // Spender burns tokens from recipient's account
            const contractAsSpender = erc20.connect(spender)
            const burnFromTx = await contractAsSpender.burnFrom(recipient.address, burnAmount)
            const burnFromReceipt = await waitWithTimeout(burnFromTx, 20000, RETRY_DELAY_FUNC)

            // Check Transfer event was emitted
            const transferEvent = findEvent(burnFromReceipt.logs, erc20.interface, 'Transfer')
            expect(transferEvent, 'Transfer event must be emitted').to.exist
            expect(transferEvent.args.from).to.equal(recipient.address)
            expect(transferEvent.args.to).to.equal('0x0000000000000000000000000000000000000000') // Zero address for burning
            expect(transferEvent.args.value).to.equal(burnAmount)

            // Check new balance
            const newBalance = await erc20.balanceOf(recipient.address)
            expect(newBalance).to.equal(initialBalance - burnAmount)

            // Check allowance was reduced
            const newAllowance = await erc20.allowance(recipient.address, spender.address)
            expect(newAllowance).to.equal(0)
        })
    })

    describe('transferOwnership', function () {
        it('should revert if the caller is not the contract owner', async function () {
            // Connect as spender (non-owner) and attempt to transfer ownership - this should revert
            const contractAsSpender = erc20.connect(spender)

            // Attempt to transfer ownership as non-owner spender to recipient - should revert
            await expect(contractAsSpender.transferOwnership(recipient.address))
                .to.be.reverted
        })

        it('should transfer ownership when called by the current owner', async function () {
            // Get initial owner
            const initialOwner = await erc20.owner()
            expect(initialOwner).to.equal(owner.address)

            // Connect as owner and transfer ownership
            const contractAsOwner = erc20.connect(owner)

            // Transfer ownership to spender
            const tx = await contractAsOwner.transferOwnership(spender.address)
            const receipt = await waitWithTimeout(tx, 20000, RETRY_DELAY_FUNC)

            expect(tx).to.not.be.reverted

            // Check ownership has changed
            const newOwner = await erc20.owner()
            expect(newOwner).to.equal(spender.address)
        })
    })


})
