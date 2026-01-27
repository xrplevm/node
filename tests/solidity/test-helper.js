const fs = require('fs')
const path = require('path')
const { spawn } = require('child_process')
const yargs = require('yargs/yargs')
const { hideBin } = require('yargs/helpers')

const logger = {
  warn: (msg) => console.error(`WARN: ${msg}`),
  err: (msg) => console.error(`ERR: ${msg}`),
  info: (msg) => console.log(`INFO: ${msg}`)
}

function panic (errMsg) {
  logger.err(errMsg)
  process.exit(-1)
}

// Function to extract EVMChainID from Go config file
function extractChainIDFromGo(goFilePath) {
  try {
    if (!fs.existsSync(goFilePath)) {
      logger.warn(`Go config file not found at ${goFilePath}, using default chain ID: 1440002`)
      return 1440002
    }

    const goFileContent = fs.readFileSync(goFilePath, 'utf8')

    // Look for DefaultEVMChainID = number
    const chainIdMatch = goFileContent.match(/DefaultEVMChainID\s*=\s*(\d+)/)

    if (chainIdMatch) {
      const chainId = parseInt(chainIdMatch[1], 10)
      logger.info(`Extracted DefaultEVMChainID from Go config: ${chainId}`)
      return chainId
    }

    logger.warn('DefaultEVMChainID not found in Go file, using default: 1440002')
    return 1440002
  } catch (error) {
    logger.warn(`Error reading Go config file: ${error.message}, using default: 1440002`)
    return 1440002
  }
}

// Function to update Hardhat config with the extracted chain ID
function updateHardhatConfig(chainId, hardhatConfigPath) {
  try {
    if (!fs.existsSync(hardhatConfigPath)) {
      logger.warn(`Hardhat config not found at ${hardhatConfigPath}`)
      return
    }

    let configContent = fs.readFileSync(hardhatConfigPath, 'utf8')

    // Find the cosmos network block and update chainId within it
    const cosmosBlockRegex = /cosmos:\s*{([^{}]*(?:{[^{}]*}[^{}]*)*)}/
    const match = configContent.match(cosmosBlockRegex)

    if (match) {
      const cosmosBlock = match[1]
      const chainIdRegex = /chainId:\s*\d+/

      if (chainIdRegex.test(cosmosBlock)) {
        const updatedCosmosBlock = cosmosBlock.replace(chainIdRegex, `chainId: ${chainId}`)
        const updatedContent = configContent.replace(cosmosBlockRegex, `cosmos: {${updatedCosmosBlock}}`)

        fs.writeFileSync(hardhatConfigPath, updatedContent)
        logger.info(`Updated Hardhat config with chainId: ${chainId}`)
      } else {
        logger.warn('chainId not found in cosmos network block')
        logger.info('Cosmos block content:', cosmosBlock)
      }
    } else {
      logger.warn('Could not find cosmos network block in Hardhat config')
      logger.info('Please check if your Hardhat config has a cosmos network configuration')

      // Show available network blocks for debugging
      const networkMatches = configContent.match(/\w+:\s*{[^{}]*}/g)
      if (networkMatches) {
        logger.info('Found network blocks:', networkMatches)
      }
    }
  } catch (error) {
    logger.warn(`Error updating Hardhat config: ${error.message}`)
  }
}

// Function to create backup of Hardhat config
function backupHardhatConfig(hardhatConfigPath) {
  const backupPath = hardhatConfigPath + '.backup'
  try {
    if (fs.existsSync(hardhatConfigPath)) {
      fs.copyFileSync(hardhatConfigPath, backupPath)
      logger.info(`Created backup: ${backupPath}`)
      return backupPath
    }
  } catch (error) {
    logger.warn(`Error creating backup: ${error.message}`)
  }
  return null
}

// Function to restore Hardhat config from backup
function restoreHardhatConfig(hardhatConfigPath, backupPath) {
  try {
    if (backupPath && fs.existsSync(backupPath)) {
      fs.copyFileSync(backupPath, hardhatConfigPath)
      fs.unlinkSync(backupPath) // Remove backup file
      logger.info('Restored original Hardhat config')
    }
  } catch (error) {
    logger.warn(`Error restoring config: ${error.message}`)
  }
}

// Function to sync configuration from Go to Hardhat
function syncConfiguration() {
  // Adjust these paths based on your project structure
  const goConfigPath = path.join(__dirname, '../../server/config/config.go')
  const hardhatConfigPath = path.join(__dirname, './suites/precompiles/hardhat.config.js')

  logger.info('Syncing configuration from Go to Hardhat...')

  // Create backup before modifying
  const backupPath = backupHardhatConfig(hardhatConfigPath)

  const chainId = extractChainIDFromGo(goConfigPath)
  updateHardhatConfig(chainId, hardhatConfigPath)

  return { hardhatConfigPath, backupPath }
}

function checkTestEnv () {
  const argv = yargs(hideBin(process.argv))
      .usage('Usage: $0 [options] <tests>')
      .example('$0 --network cosmos', 'run all tests using cosmos evm network')
      .example(
          '$0 --network cosmos --allowTests=test1,test2',
          'run only test1 and test2 using cosmos network'
      )
      .help('h')
      .alias('h', 'help')
      .describe('network', 'set which network to use: ganache|cosmos')
      .describe(
          'batch',
          'set the test batch in parallelized testing. Format: %d-%d'
      )
      .describe('allowTests', 'only run specified tests. Separated by comma.')
      .boolean('verbose-log')
      .describe('verbose-log', 'print exrpd output, default false').argv

  if (!fs.existsSync(path.join(__dirname, './node_modules'))) {
    panic(
        'node_modules not existed. Please run `yarn install` before running tests.'
    )
  }
  const runConfig = {}

  // Check test network
  if (!argv.network) {
    runConfig.network = 'cosmos'
  } else {
    if (argv.network !== 'cosmos' && argv.network !== 'ganache') {
      panic('network is invalid. Must be ganache or cosmos')
    } else {
      runConfig.network = argv.network
    }
  }

  if (argv.batch) {
    const [toRunBatch, allBatches] = argv.batch
        .split('-')
        .map((e) => Number(e))

    console.log([toRunBatch, allBatches])
    if (!toRunBatch || !allBatches) {
      panic('bad batch input format')
    }

    if (toRunBatch > allBatches) {
      panic('test batch number is larger than batch counts')
    }

    if (toRunBatch <= 0 || allBatches <= 0) {
      panic('test batch number or batch counts must be non-zero values')
    }

    runConfig.batch = {}
    runConfig.batch.this = toRunBatch
    runConfig.batch.all = allBatches
  }

  // only test
  runConfig.onlyTest = argv.allowTests
      ? argv.allowTests.split(',')
      : undefined
  runConfig.verboseLog = !!argv['verbose-log']

  logger.info(`Running on network: ${runConfig.network}`)
  return runConfig
}

function loadTests (runConfig) {
  let validTests = []
  fs.readdirSync(path.join(__dirname, 'suites')).forEach((dirname) => {
    const dirStat = fs.statSync(path.join(__dirname, 'suites', dirname))
    if (!dirStat.isDirectory) {
      logger.warn(`${dirname} is not a directory. Skip this test suite.`)
      return
    }

    const needFiles = ['package.json', 'test']
    for (const f of needFiles) {
      if (!fs.existsSync(path.join(__dirname, 'suites', dirname, f))) {
        logger.warn(
            `${dirname} does not contains file/dir: ${f}. Skip this test suite.`
        )
        return
      }
    }

    // test package.json
    try {
      const testManifest = JSON.parse(
          fs.readFileSync(
              path.join(__dirname, 'suites', dirname, 'package.json'),
              'utf-8'
          )
      )
      const needScripts = ['test-ganache', 'test-cosmos']
      for (const s of needScripts) {
        if (Object.keys(testManifest.scripts).indexOf(s) === -1) {
          logger.warn(
              `${dirname} does not have test script: \`${s}\`. Skip this test suite.`
          )
          return
        }
      }
    } catch (error) {
      logger.warn(
          `${dirname} test package.json load failed. Skip this test suite.`
      )
      logger.err(error)
      return
    }
    validTests.push(dirname)
  })

  if (runConfig.onlyTest) {
    validTests = validTests.filter((t) => runConfig.onlyTest.indexOf(t) !== -1)
  }

  if (runConfig.batch) {
    const chunkSize = Math.ceil(validTests.length / runConfig.batch.all)
    const toRunTests = validTests.slice(
        (runConfig.batch.this - 1) * chunkSize,
        runConfig.batch.this === runConfig.batch.all
            ? undefined
            : runConfig.batch.this * chunkSize
    )
    return toRunTests
  } else {
    return validTests
  }
}

function performTestSuite ({ testName, network }) {
  const cmd = network === 'ganache' ? 'test-ganache' : 'test-cosmos'
  return new Promise((resolve, reject) => {
    const testProc = spawn('yarn', [cmd], {
      cwd: path.join(__dirname, 'suites', testName)
    })

    testProc.stdout.pipe(process.stdout)
    testProc.stderr.pipe(process.stderr)

    testProc.on('close', (code) => {
      if (code === 0) {
        console.log('end')
        resolve()
      } else {
        reject(new Error(`Test: ${testName} exited with error code ${code}`))
      }
    })
  })
}

async function performTests ({ allTests, runConfig }) {
  if (allTests.length === 0) {
    panic('No tests are found or all invalid!')
  }

  for (const currentTestName of allTests) {
    logger.info(`Start test: ${currentTestName}`)
    await performTestSuite({
      testName: currentTestName,
      network: runConfig.network
    })
  }

  logger.info(`${allTests.length} test suites passed!`)
}

function setupNetwork ({ runConfig, timeout }) {
  if (runConfig.network !== 'cosmos') {
    // no need to start ganache. Truffle will start it
    return
  }

  // Spawn the cosmos evm process

  const spawnPromise = new Promise((resolve, reject) => {
    const serverStartedLog = 'Starting JSON-RPC server'
    const serverStartedMsg = 'exrpd started'

    const scriptPath = path.join(__dirname, 'init-node.sh');  // → "init-node.sh"

    const osdProc = spawn(scriptPath, ['-y'], {
      cwd: __dirname,
      stdio: ['ignore', 'pipe', 'pipe'],  // <-- stdout/stderr streams
    })

    logger.info(`Starting exrpd process... timeout: ${timeout}ms`)
    if (runConfig.verboseLog) {
      osdProc.stdout.pipe(process.stdout)
      osdProc.stderr.pipe(process.stderr)
    }


    osdProc.stdout.on('data', (d) => {
      const oLine = d.toString()
      if (runConfig.verboseLog) {
        process.stdout.write(oLine)
      }

      if (oLine.indexOf(serverStartedLog) !== -1) {
        logger.info(serverStartedMsg)
        resolve(osdProc)
      }
    })

    osdProc.stderr.on('data', (d) => {
      const oLine = d.toString()
      if (runConfig.verboseLog) {
        process.stdout.write(oLine)
      }

      if (oLine.indexOf(serverStartedLog) !== -1) {
        logger.info(serverStartedMsg)
        resolve(osdProc)
      }
    })
  })

  const timeoutPromise = new Promise((resolve, reject) => {
    setTimeout(() => reject(new Error('Start exrpd timeout!')), timeout)
  })
  return Promise.race([spawnPromise, timeoutPromise])
}

async function main () {
  // Sync configuration before running tests
  const configPaths = syncConfiguration()

  let proc = null

  try {
    const runConfig = checkTestEnv()
    const allTests = loadTests(runConfig)

    console.log(`Running Tests: ${allTests.join()}`)

    proc = await setupNetwork({ runConfig, timeout: 200000 })

    // sleep for 20s to wait blocks being produced
    //
    // TODO: this should be handled more gracefully, i.e. check for block height
    await new Promise((resolve) => setTimeout(resolve, 20000))

    await performTests({ allTests, runConfig })

    logger.info('Tests completed successfully!')
  } catch (error) {
    logger.err(`Test execution failed: ${error.message}`)
    throw error
  } finally {
    // Always restore the original config, even if tests fail
    if (configPaths) {
      restoreHardhatConfig(configPaths.hardhatConfigPath, configPaths.backupPath)
    }

    if (proc) {
      proc.kill()
    }
  }

  process.exit(0)
}

// Add handler to exit the program when UnhandledPromiseRejection
process.on('unhandledRejection', (e) => {
  console.error(e)

  // Try to restore config if possible
  const hardhatConfigPath = path.join(__dirname, 'hardhat.config.js')
  const backupPath = hardhatConfigPath + '.backup'
  if (fs.existsSync(backupPath)) {
    try {
      fs.copyFileSync(backupPath, hardhatConfigPath)
      fs.unlinkSync(backupPath)
      logger.info('Restored original Hardhat config after error')
    } catch (restoreError) {
      logger.warn(`Could not restore config: ${restoreError.message}`)
    }
  }

  process.exit(-1)
})

// Handle SIGINT (Ctrl+C) to restore config
process.on('SIGINT', () => {
  logger.info('Received SIGINT, cleaning up...')

  const hardhatConfigPath = path.join(__dirname, 'hardhat.config.js')
  const backupPath = hardhatConfigPath + '.backup'
  if (fs.existsSync(backupPath)) {
    try {
      fs.copyFileSync(backupPath, hardhatConfigPath)
      fs.unlinkSync(backupPath)
      logger.info('Restored original Hardhat config')
    } catch (restoreError) {
      logger.warn(`Could not restore config: ${restoreError.message}`)
    }
  }

  process.exit(0)
})

main()