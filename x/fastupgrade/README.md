# x/fastupgrade

Fast-tracks single-message minor software upgrades once bonded validators holding at least `threshold` of bonded tokens vote Yes. This README contains the full operator contract.

The v12.0.0 upgrade enables the module with `timelock_blocks` 150, `threshold` 0.85, and `postpone_blocks` 10. If gov's params are incompatible with that threshold at the upgrade, the module stays disabled and the node logs `x/fastupgrade stays disabled`. Gov can then enable it with `MsgUpdateParams`. A generic genesis has it disabled. An exported genesis keeps the exported params, including `enabled`, but holds no timer and no completed-upgrade markers. Neither chain can start the fast path until x/upgrade completes a real upgrade on it.

## What qualifies

A proposal is eligible when all of these hold:

- It has exactly one message, a `MsgSoftwareUpgrade`.
- The plan is valid for x/upgrade: a positive height that is not below the current height, no `time`, and no `upgraded_client_state`.
- The plan name is strict `vMAJOR.MINOR.PATCH`: digits only, no leading zeros, no suffix.
- The plan has the same major version and a higher minor version than the last completed upgrade. The last completed upgrade is the one x/upgrade completed at the greatest height, not the highest version.
- The plan name is not an upgrade that already ran.

Anything else, including patch-only and major upgrades, uses stock governance.

## Operator flow

1. Before voting, install the two-binary bundle for the upgrade (`exrpd-bundle_<current>-<target>`, for example `exrpd-bundle_12.0.0-12.1.0`) and restart the service. See [Install a new bundle file](../../docs/bundle.md#install-a-new-bundle-file). The bundle then switches the node to the target binary by itself at the upgrade height.
2. Vote Yes with a normal `MsgVote` signed by the validator's operator account: the account address (`ethm1...`) with the same bytes as the validator's `ethmvaloper1...` address. Votes from any other account do not count for the gate. A Yes vote means "this node is ready". A vote through the EVM gov precompile counts in the same way while that precompile is active.
3. In the block where Yes power reaches the threshold and gov's own tally would pass, the module starts a timer. The `fastupgrade_started` event and `exrpd q fastupgrade timer` give the upgrade height H, which is that block's height plus `timelock_blocks`. A proposal whose original plan height is below H-1 does not start a timer, because that height would be in the past at the final check.
4. Until H-1, any validator can change its vote. A No vote reduces Yes power but does not block the fast path if Yes power still meets the threshold.
5. At H-1 the module rechecks everything. If the checks pass, it rewrites the plan height to H and ends the vote. Gov executes the proposal in that same block, and the chain halts for the upgrade at H. If any check fails, the timer is dropped and the proposal continues under its normal deadline with its original plan. The proposal's own deadline always wins: a timer can start when that deadline comes before H-1, and it is then dropped in the block of the deadline.
6. If the checks pass but another proposal ends in block H-1, the module does not close the vote. It moves H later by `postpone_blocks`, announces the new H with `fastupgrade_postponed`, and repeats step 5 at the new H-1.

## Queries and transactions

```bash
exrpd q fastupgrade params -o json   # {"params":{"enabled":true,"timelock_blocks":"150","threshold":"0.850000000000000000","postpone_blocks":"10"}}
exrpd q fastupgrade timer -o json    # {} with no timer, else {"timer":{"proposal_id":"N","upgrade_height":"H"}}
```

The CLI's JSON output omits fields with zero values, so `enabled` is absent when it is false. REST: `/exrp/fastupgrade/params`, `/exrp/fastupgrade/timer`.

Only gov can change params, through a `MsgUpdateParams` proposal. The module has no transaction command of its own. Create the proposal with `exrpd tx gov draft-proposal` and choose `/fastupgrade.MsgUpdateParams` as the message type, or write the JSON below. Submit it with `exrpd tx gov submit-proposal proposal.json --from <key> ...`. `authority` is the gov module address from `exrpd q auth module-account gov`:

```json
{
  "messages": [{
    "@type": "/fastupgrade.MsgUpdateParams",
    "authority": "<gov module address>",
    "params": {"enabled": true, "timelock_blocks": "150", "threshold": "0.850000000000000000", "postpone_blocks": "10"}
  }],
  "deposit": "<deposit>",
  "title": "Update fastupgrade params",
  "summary": "..."
}
```

All params must be supplied:

- `timelock_blocks` must be in [2, 100000].
- `threshold` must be in (0, 1] with at most 18 decimals.
- `postpone_blocks` must be in [1, 100000].
- With `enabled` true, `threshold` must stay compatible with gov's params: greater than `threshold` and `expedited_threshold`, and at least `quorum` and `1 - veto_threshold`. With `enabled` false, this is not checked, so gov can always turn the module off.

A new threshold applies at the next check, including a running timer's H-1 check. A new timelock affects only timers started later. A new `postpone_blocks` applies at the next postponement. To turn the module off, set `enabled` to false; a running timer is dropped at the next block.

## Events

All events are emitted in EndBlock.

| Event | Attributes | Meaning |
| --- | --- | --- |
| `fastupgrade_started` | `proposal_id`, `plan_name`, `plan_info`, `upgrade_height` | Timer started; H announced. |
| `fastupgrade_not_started` | `proposal_id`, `reason` | Validators reached the threshold, but a later start check failed. |
| `fastupgrade_dropped` | `proposal_id`, `upgrade_height`, `reason` | Timer ended without firing; stock gov continues. |
| `fastupgrade_postponed` | `proposal_id`, `upgrade_height` | Another proposal ended in block H-1; `upgrade_height` is the new H. |
| `fastupgrade_early_close` | `proposal_id`, `plan_name`, `upgrade_height` | Vote closed at H-1; gov executes the proposal in this block. |

Reasons:

| Reason | Meaning |
| --- | --- |
| `incompatible_with_gov` | Our threshold no longer fits gov's params. The `fast upgrade threshold is incompatible with gov params` log names the conflict. Raise our threshold or restore gov's params. |
| `gate_below_threshold` | Validator Yes power dropped below the threshold. |
| `gov_tally_rejected` | Gov's tally would not pass. Unexpected without delegations from other accounts; investigate. |
| `tally_below_threshold` | Gov's Yes count is below the threshold. Possible with delegations from other accounts. Also possible with a weighted validator vote when Yes power is within one token unit of the threshold, because gov truncates its Yes count. |
| `natural_deadline` | The proposal's own deadline arrived before the vote closed early. The timer is dropped in the block of the deadline, and stock gov decides with the original height. If gov converts an expedited proposal to a regular one there, it deletes all votes: a new Yes vote starts a new timer. |
| `not_eligible` | The proposal is no longer eligible: another upgrade completed during the timelock, or the chain passed the plan's original height. |
| `plan_pending` | Another upgrade plan is scheduled. |
| `disabled` | The module was disabled. |
| `not_voting` | The proposal is no longer in its voting period, for example because the proposer cancelled it. |
| `missed_window` | The H-1 check did not run. The timer never fires after its own H-1. |
| `internal_error` | EndBlock hit an internal error or a panic while a timer was running. The `fast upgrade end block failed` log has the error. |

After `fastupgrade_not_started` or `fastupgrade_dropped`, the module does not check the proposal again on its own. To retry, a bonded validator must send its Yes vote again.

While another upgrade plan is scheduled, the module skips the start with no event. It does not check the proposal again after that plan is gone. To retry, a bonded validator must send its Yes vote again.

If EndBlock hits an internal error or a panic, it discards its own changes and logs `fast upgrade end block failed`. If a timer was running, it drops it with `fastupgrade_dropped` and reason `internal_error`. Without a timer it emits no event. Stock gov then decides the proposal.

Gov runs after this module in the same block. A proposal that gov executes in block H-1 can change gov's params, schedule an upgrade plan, or remove a validator after the module's check, so gov's own tally of an early-closed proposal could then fail. The module therefore postpones H when any other proposal ends in block H-1, and checks again at the new H-1. This can repeat until the proposal's own deadline, after which stock gov decides with the original plan.

Any account can cause a postponement with a proposal that ends in block H-1, including one that fails. A proposal submitted after the upgrade proposal cannot end sooner than gov's `expedited_voting_period` later, so it cannot postpone a final check that comes before that. The timelock is always much shorter than `expedited_voting_period`, because a fast path that takes as long as an expedited vote has no purpose. Vote soon after submission.

Logs, under `module=x/fastupgrade`: `fast upgrade timer started`, `fast upgrade timer not started`, `fast upgrade timer dropped`, `fast upgrade timer postponed`, `fast upgrade vote closed early`.

## State

- Params: KV store `fastupgrade`. This is the only exported genesis state.
- Timer: KV store `fastupgrade`. At most one exists.
- Nominations: KV store `fastupgrade`. It holds the proposals that bonded validators voted Yes on in the current block. EndBlock clears it in every block.

## Tests

```bash
GOWORK=off go test -tags=test ./x/fastupgrade/...
GOWORK=off go test -tags=test ./tests/integration -run 'TestIntegrationTestSuite/TestFastUpgrade'
```
