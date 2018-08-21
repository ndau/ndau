package main

import (
	cli "github.com/jawher/mow.cli"
)

func getAccount(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Command(
			"list",
			"list known accounts",
			getAccountList(verbose),
		)

		cmd.Command(
			"new",
			"create a new account",
			getAccountNew(verbose),
		)

		cmd.Command(
			"claim",
			"claim this account on the blockchain",
			getAccountClaim(verbose),
		)

		cmd.Command(
			"validation",
			"change the account's validation",
			getAccountValidation(verbose),
		)

		cmd.Command(
			"query",
			"query the ndau chain about this account",
			getAccountQuery(verbose),
		)

		cmd.Command(
			"change-settlement-period",
			"change the settlement period for outbound transfers from this account",
			getAccountChangeSettlement(verbose),
		)

		cmd.Command(
			"delegate",
			"delegate EAI calculation to a node",
			getAccountDelegate(verbose),
		)

		cmd.Command(
			"credit-eai",
			"credit EAI for accounts which have delegated to this one",
			getAccountCreditEAI(verbose),
		)

		cmd.Command(
			"lock",
			"lock this account with a specified notice period",
			getLock(verbose),
		)

		cmd.Command(
			"notify",
			"notify that this account should be unlocked once its notice period expires",
			getNotify(verbose),
		)

		cmd.Command(
			"set-rewards-target",
			"set the rewards target for this account",
			getSetRewardsDestination(verbose),
		)
	}
}
