package ndau

import (
	"encoding/base64"
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
)

func makeVKs(t *testing.T, keys ...string) []signature.PublicKey {
	vks := make([]signature.PublicKey, 0, len(keys))
	for _, ks := range keys {
		vk, err := signature.ParsePublicKey(ks)
		require.NoError(t, err)
		vks = append(vks, *vk)
	}
	return vks
}

func Test_ndacc2gihhrj6rhe3v2jx5k6gqpedy878eaxn35j4tvcdirq_History(t *testing.T) {
	app, _ := initApp(t)

	node1, err := address.Validate("ndarw5i7rmqtqstw4mtnchmfvxnrq4k3e2ytsyvsc7nxt2y7")
	require.NoError(t, err)
	modify(t, node1.String(), app, func(ad *backing.AccountData) {
		ad.ValidationKeys = makeVKs(t,
			"npuba8jadtbbeamn89h5zgr5cmjggcwkchbsgqhf5m7zb58xe7rwqwvzif23ebfqz4wh224ve2qw",
			"npuba8jadtbbeabmk869zakhpzmiv2xvzc7yyxrzcmfu6eqbw9ttyi9bwrcpiz7jqki9pwsw7vsp",
			"npuba8jadtbbebivxyxnve83n7rwdmdzg3k3mpv7ed9y5jptgsnd5qf3uu9fx7sbddf63b636s3i",
			"npuba8jadtbbed6uj93t6c8hn72bt4ypw2rxx6zmfpcfkqmmxxt5m2e7ydit3gtfpt4quxzfcmkr",
		)
	})
	err = app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Nodes[node1.String()] = backing.Node{
			Active: true,
		}
		return st, nil
	})
	require.NoError(t, err)

	node2, err := address.Validate("ndam75fnjn7cdues7ivi7ccfq8f534quieaccqibrvuzhqxa")
	require.NoError(t, err)
	modify(t, node2.String(), app, func(ad *backing.AccountData) {
		ad.ValidationKeys = makeVKs(t,
			"npuba8jadtbbea97bcz4v2c4gtcntx53cgjpv92hscm95gg2m6tntwysawxkahkse4bcdpp5dm24",
			"npuba8jadtbbeabmk869zakhpzmiv2xvzc7yyxrzcmfu6eqbw9ttyi9bwrcpiz7jqki9pwsw7vsp",
			"npuba8jadtbbebivxyxnve83n7rwdmdzg3k3mpv7ed9y5jptgsnd5qf3uu9fx7sbddf63b636s3i",
			"npuba8jadtbbed6uj93t6c8hn72bt4ypw2rxx6zmfpcfkqmmxxt5m2e7ydit3gtfpt4quxzfcmkr",
		)
	})
	err = app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Nodes[node2.String()] = backing.Node{
			Active: true,
		}
		return st, nil
	})
	require.NoError(t, err)

	ts := math.Timestamp(576201600000000)
	// create the account
	// from https://github.com/oneiro-ndev/genesis/blob/master/pkg/etl/transform.go
	modify(t, "ndacc2gihhrj6rhe3v2jx5k6gqpedy878eaxn35j4tvcdirq", app, func(ad *backing.AccountData) {
		ad.Balance = 1000 * constants.NapuPerNdau
		ad.LastEAIUpdate = ts
		ad.LastWAAUpdate = ts
		ad.CurrencySeatDate = &ts
		ad.Lock = backing.NewLock(
			math.Year+(2*math.Month)+(22*math.Day),
			eai.DefaultLockBonusEAI,
		)
		ad.Lock.Notify(ts, 0)
		ad.RecourseSettings.Period = math.Hour
	})

	addr, err := address.Validate("ndacc2gihhrj6rhe3v2jx5k6gqpedy878eaxn35j4tvcdirq")
	require.NoError(t, err)
	err = app.UpdateStateImmediately(app.Delegate(addr, node1))
	require.NoError(t, err)

	// set the EAI overtime system var above what we need
	overtime, err := math.Duration(10 * math.Year).MarshalMsg(nil)
	require.NoError(t, err)
	context := ddc(t).with(func(svs map[string][]byte) {
		svs[sv.EAIOvertime] = overtime
	})

	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQm8wcHHPdEembIgJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxAqNzaWeRkgHEQI2kYXg5Kd6LW6XKDDups73s8gwKWA70LGOYkY0Rr7IR/lP8FDuXAbtCpy6BZ4oaSe+3CKLAdNEzAQH0bpafpQM=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(610888325348024))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(113065871044), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQRWImtXZ3Eem+3AJCrBEAAq5UcmFuc2FjdGFibGVJRAKsVHJhbnNhY3RhYmxlhaN0Z3SR2TBuZGFjYzJnaWhocmo2cmhlM3Yyang1azZncXBlZHk4NzhlYXhuMzVqNHR2Y2RpcnGja2V5k5ICxEohAllO1Ze1+SGCy0Sa5Fkch5Lyp6bSrsttqDGGycXtYtZQBcukjAAAAAHaVtwbo/R0t1uoEe4QJ3n8zM7XI/CfxGiC2Ol788tm/pIBxCEgjhJKJhJ8yMWecbsI2SZhZt3rFradW1SMiC9vwIEo4kCSAcQhIL82ERxzKiFdsKcokKyJjiAqOCAHwVqOpc3Y+2vYMC+3o3ZhbMQEoAAgiKNzZXEio3NpZ5GSAcRA6jtG/hPKD7ip1y2gqVAtMzPq+qEa9r092wJGWOzMz39suEm0JlnXRSj510wGJJNesxPaEEe9U4ZeJi/xtHpgBg==")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(611174224000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(113061821044), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQMhzT7nw2EemQQwJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxCqNzaWeRkgHEQJ5CDw7bFI2QHk0U7rVQz0LDJM//Y1bIQjGCjhAFGEGTAhNP4Ba5z1j+RtiFml3Zy8ujmimkPDYHTR/41aeV+As=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(611805982000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(113312425269), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQ7o/waoPOEemGwwJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxC6NzaWeRkgHEQCpd0bHrutqX8eulZISTrsbqVjxsiX1W4uYN11ggI6jBzTZaeZa9usJDW2jOcbEgjqBwqL6bG7TYJ/V3oJoIYwI=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(612641044000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(113644548794), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQg3nnAYPPEemGwwJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxDKNzaWeRkgHEQIEWKbbSkrLhw/WSkRQIZJiqomRjCQe46loN4TKdwtpk4Bc34MkdFxJOd970ok7Vim6PKx1lfqHz3xzzKve6bwI=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(612641280000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(113644642818), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQkPRcr4vuEemUYAJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxDaNzaWeRkgHEQMhK/LCcGcCPnya1ROPP+oXGI2Y/3gWqkGsUk44ru5mIHpdMDFzL6XMzvBMeh7YLz5ngBUsmLOuT/xxDSinLiA0=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(613534215000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114000867544), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQnZAuVZEtEemHxwJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxDqNzaWeRkgHEQPOjrN6j0KJj7z1cFfaiIG2AU3kQZuR1wws3RVtLGqcN4ANO9xNW7gf0fdfAy9ZdweQIHVko/CXYBkvFNjVOfwE=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(614111252000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114231639863), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQJXXqwpaeEemL9gJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxD6NzaWeRkgHEQN1pW9YFCWq0w6tCiYt/yOoSQid3MMUiBdoA9PVdoTjji1yI72knkVxUuGEtrKhK2kqNwVF0G5UsuknyIVgbHg0=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(614709075000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114471219722), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQCYLczpazEemL9gJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxEKNzaWeRkgHEQFOAed/mSIX6voxanbmz4SNBI8H9VmTmjXBng/KvZgxFdElmUJVq4yHfMG5p3Q748pOqN66oXjThAqQz3zMLtAA=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(614718115000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114474845516), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQLXtsSJg2EemSQAJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxEaNzaWeRkgHEQIMEZN/MaMZek8MoiDAPOCXIWHQodEjOO4izlsZVG3vwQOMBKyYjJKDhuadfBkigMCT1FxKTm18DTKaayipgkA8=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(614884462000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114535843598), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQUuFrMJjTEemxtgJCrBEAAq5UcmFuc2FjdGFibGVJRAesVHJhbnNhY3RhYmxlhKN0Z3SR2TBuZGFjYzJnaWhocmo2cmhlM3Yyang1azZncXBlZHk4NzhlYXhuMzVqNHR2Y2RpcnGjcGVy0wAABxJ9t8AAo3NlcSOjc2lnkZICxEYwRAIgYPPCxkVG66tgHDP0sfOTkfrM27nYEa3MnwmyRhKeIAMCIFeViEUKISiOR74hSSvelhC74zf763CwbXCnhS41eCy4")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(614951944000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114535343598), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQb0biXpjTEem72QJCrBEAAq5UcmFuc2FjdGFibGVJRAisVHJhbnNhY3RhYmxlg6N0Z3SR2TBuZGFjYzJnaWhocmo2cmhlM3Yyang1azZncXBlZHk4NzhlYXhuMzVqNHR2Y2RpcnGjc2VxJKNzaWeRkgLERzBFAiEA+dvItQxUguULk8x6iRFtKBUtiwgkySrpfFhY1mJYTYECICf4hJkN0t5LLBWSiraHIhYftlFgipRFXwknYLUT06sm")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(614952114000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114534843598), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQJKIcLZm9EemMrAJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxEqNzaWeRkgHEQFG1fv5sok5a+DdGVs/AKBJQDloEWuUk52e4e+eimhGCFVK8Q8GVmtMDjWsYm+wMbmY02a1+rCF+bB3tblv3aQQ=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(615052374000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114568895887), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQdkiDMaKFEemyCgJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxE6NzaWeRkgHEQPYvpXW8fAyunyIxZUCoDDA7AKhY5aHYFer9vRRzJ9qg8xffFEMBJs5VkZPVVv6UZDiWGKlkLEePp4pqwEHtnQ0=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(616018135000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114897499778), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQPNDIH6fPEemGiAJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxFKNzaWeRkgHEQL31YbAcbM4ib/qECYeMLfsSS3bqVntgOM3HbqjAs4DrUQMD8HdjTkVmsps6hZyF2nO/EjFIzPwVS5vSPqG0/Ac=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(616599467000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(115095734748), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQuboddKzBEem65AJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxFaNzaWeRkgHEQG8Q07XIqueNTh376BUcZNU2EAJcPT3ehZy34tFO4bbiYytrqR91fTA/HcVQwTS/gF8gOt9N+BiiLEBsIC7GUwY=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(617143559000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(115281579031), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQV5Gt1K1TEem65AJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFydzVpN3JtcXRxc3R3NG10bmNobWZ2eG5ycTRrM2UyeXRzeXZzYzdueHQyeTejc2VxFqNzaWeRkgHEQGiav06uCjJLbBfJvbjvRvqvYjT9G3ABMR6TLKmA9tqvYKeEBtiIy9qxHgCxWEu8OA7fx3VpfsUGkHJmKm7HkQ8=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(617206002000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(115302923970), acct.Balance)
	}
}

func Test_ndadyrd7u7kyjkq9nwcz3rgyi3m6fyeexwwi6hy6giwby7wy_History(t *testing.T) {
	app, _ := initApp(t)

	node1, err := address.Validate("ndarw5i7rmqtqstw4mtnchmfvxnrq4k3e2ytsyvsc7nxt2y7")
	require.NoError(t, err)
	modify(t, node1.String(), app, func(ad *backing.AccountData) {
		ad.ValidationKeys = makeVKs(t,
			"npuba8jadtbbeamn89h5zgr5cmjggcwkchbsgqhf5m7zb58xe7rwqwvzif23ebfqz4wh224ve2qw",
			"npuba8jadtbbeabmk869zakhpzmiv2xvzc7yyxrzcmfu6eqbw9ttyi9bwrcpiz7jqki9pwsw7vsp",
			"npuba8jadtbbebivxyxnve83n7rwdmdzg3k3mpv7ed9y5jptgsnd5qf3uu9fx7sbddf63b636s3i",
			"npuba8jadtbbed6uj93t6c8hn72bt4ypw2rxx6zmfpcfkqmmxxt5m2e7ydit3gtfpt4quxzfcmkr",
		)
	})
	err = app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Nodes[node1.String()] = backing.Node{
			Active: true,
		}
		return st, nil
	})
	require.NoError(t, err)

	node2, err := address.Validate("ndam75fnjn7cdues7ivi7ccfq8f534quieaccqibrvuzhqxa")
	require.NoError(t, err)
	modify(t, node2.String(), app, func(ad *backing.AccountData) {
		ad.ValidationKeys = makeVKs(t,
			"npuba8jadtbbea97bcz4v2c4gtcntx53cgjpv92hscm95gg2m6tntwysawxkahkse4bcdpp5dm24",
			"npuba8jadtbbeabmk869zakhpzmiv2xvzc7yyxrzcmfu6eqbw9ttyi9bwrcpiz7jqki9pwsw7vsp",
			"npuba8jadtbbebivxyxnve83n7rwdmdzg3k3mpv7ed9y5jptgsnd5qf3uu9fx7sbddf63b636s3i",
			"npuba8jadtbbed6uj93t6c8hn72bt4ypw2rxx6zmfpcfkqmmxxt5m2e7ydit3gtfpt4quxzfcmkr",
		)
	})
	err = app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Nodes[node2.String()] = backing.Node{
			Active: true,
		}
		return st, nil
	})
	require.NoError(t, err)

	ts := math.Timestamp(576201600000000)
	// create the account
	// from https://github.com/oneiro-ndev/genesis/blob/master/pkg/etl/transform.go
	modify(t, "ndadyrd7u7kyjkq9nwcz3rgyi3m6fyeexwwi6hy6giwby7wy", app, func(ad *backing.AccountData) {
		ad.Balance = 1000 * constants.NapuPerNdau
		ad.LastEAIUpdate = ts
		ad.LastWAAUpdate = ts
		ad.CurrencySeatDate = &ts
		ad.Lock = backing.NewLock(
			math.Year+(2*math.Month)+(22*math.Day),
			eai.DefaultLockBonusEAI,
		)
		ad.Lock.Notify(ts, 0)
		ad.RecourseSettings.Period = math.Hour
	})

	addr, err := address.Validate("ndadyrd7u7kyjkq9nwcz3rgyi3m6fyeexwwi6hy6giwby7wy")
	require.NoError(t, err)
	err = app.UpdateStateImmediately(app.Delegate(addr, node2))
	require.NoError(t, err)

	// set the EAI overtime system var above what we need
	overtime, err := math.Duration(10 * math.Year).MarshalMsg(nil)
	require.NoError(t, err)
	context := ddc(t).with(func(svs map[string][]byte) {
		svs[sv.EAIOvertime] = overtime
	})

	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQojkXznPdEembIgJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxAqNzaWeRkgHEQPMfAyW/6CNAc9Y/G7YlR7kf3fSPcKUJO46ZEXK3LbfsKB6lJJxVc7s7T6BNSx67vT8mcc/Ilga9+dD98EpcTgk=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(610888336000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(113065875436), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQaZR7W3Z3Eem+3AJCrBEAAq5UcmFuc2FjdGFibGVJRAKsVHJhbnNhY3RhYmxlhaN0Z3SR2TBuZGFkeXJkN3U3a3lqa3E5bndjejNyZ3lpM202ZnllZXh3d2k2aHk2Z2l3Ynk3d3mja2V5k5ICxEohA8Kch0GOOzKpOu5IKycy67vdcAYDE+QN6TujEcwEoaEYBXiyJwAAAAO5aHPYT++eMeHEEbA9B4pGmND7yLiiLrjxcCklLy1lcZIBxCEgDX5hLjA6AqB2RKurpLgivvn+HIF3qalYi+ulgFToDzOSAcQhIL82ERxzKiFdsKcokKyJjiAqOCAHwVqOpc3Y+2vYMC+3o3ZhbMQEoAAgiKNzZXFJo3NpZ5GSAcRA2Mz/Y0xlKGxp6Z/LLcnP7sQvhBZ+d3Y4HNaIq9rv+tx8APJ0Stbd3oZVGfwd1moCswDN/mrSZDPDn5naUNluBA==")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(611174284000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(113061825436), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQ8rllnHw2EemQQwJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxCqNzaWeRkgHEQGWZ8Hnyb99gFln/+pMW+VHlV+6eeLAtn8ip08jPYaF8RiV8WWHaCgnA76VnfMmanzCat/o4ADLeKFv1sdenkQU=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(611806304000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(113312533243), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQ8vYSBYPOEemGwwJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxC6NzaWeRkgHEQLlbdJdO6m0cawJNHgm1QU1uH/R0Fjqup3j3FicXR3oQUCngfEb3byn2ywSqeUtZ7aLkRs4rGKaWLnwiE5Pe2wU=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(612641247000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(113644609690), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQh8fdtoPPEemGwwJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxDKNzaWeRkgHEQF/g5qTL9azeJnYixmw/q5S1Oqs4X0/h64QJw9wdwKb/dpLpB/KXMRXyfClKPnoRV2gDyz++dep6FB5c+7X/Wgw=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(612641496000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(113644709052), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQlVAoGovuEemUYAJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxDaNzaWeRkgHEQE72KXnnp51vjWjqtCPyfbg+2lW9gQUuCp+iNlCXjpqxegH5bKguohkgNe1iv9yKmvIDO2yjClU/Lqo9XyUhLQE=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(613534443000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114000938578), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQocEJ9ZEtEemHxwJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxDqNzaWeRkgHEQDG88Z6u2Yd6wJMNtJgYSUh7H9+AAB7DQJokdzOKcYCpY087mgQQkFy4YRnnT3YTtZt8gS8NR8ZaFEEw5iR9NQ4=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(614111277000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114231630301), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQKml47JaeEemL9gJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxD6NzaWeRkgHEQOylnQDoI20gw5qcvx6s488xqsfEo4p8+iFLmXKzT0ofXKC/CE0sR+YeOQg0vPyU63Rkb5CYZm6MUq5Jdi2paQk=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(614709365000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114471316145), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQDh6V85azEemL9gJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxEKNzaWeRkgHEQEhHaHWIT3+Ud5rTMudt9cRUy0L6gWYdODLWOy4ErkLVCy0XfAITfyGF0wy9jdlZxXRC6UwHZGjuq+hHK8BOQwI=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(614718338000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114474915280), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQMbqV7Zg2EemSQAJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxEaNzaWeRkgHEQG0VbwlNsoHtBffgAMMNGz4f5iRsBEOZXfX5gPm2tW+8L2sxETwcRSUegn9/uH5FU3ueyymSlqqGgB3MxUhIUAs=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(614884613000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114535870536), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQKSwVVZm9EemMrAJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxEqNzaWeRkgHEQIUFR0icsvFvloidmR/nApShK/m/qnJj94gGvgXHrt7Mh89Lm2Z2ydQ8hautl8Ruo4/IT5nli7vvJEybEJPeXAE=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(615052532000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114606941224), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQghVxVKKFEemNvAJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxE6NzaWeRkgHEQIf1jf1paDrz0Sl0e7DylBUcpk1hn1CWeoAKmMcHauArH2xoeNk32rbqzzSY0AjN+4RR1WhnocxXXZ9VcZahggE=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(616018190000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(114977026651), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQSPxITafPEemLswJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxFKNzaWeRkgHEQCXgXSUnTBx3pDm70ToBASxA1b8Li5NybS8C37irrGWdln8F1JMMcJH5DLWnLkYUHwXQpAu+DLm/y5G5vv5DQgg=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(616599633000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(115529348351), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQ35d/IaiFEem7twJCrBEAAq5UcmFuc2FjdGFibGVJRAesVHJhbnNhY3RhYmxlhKN0Z3SR2TBuZGFkeXJkN3U3a3lqa3E5bndjejNyZ3lpM202ZnllZXh3d2k2aHk2Z2l3Ynk3d3mjcGVy0wAABxJ9t8AAo3NlcUqjc2lnkZICxEYwRAIgLieoYcLVd6DC9lf48TacTFliz9tvCW29MIh5T04xo9sCIEvGt1EpcmQdFZQbEnyUq56DaqH/YZLPirOOeDsoUsZB")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(616677905000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(115528848351), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQ4YWFG6iFEemWKwJCrBEAAq5UcmFuc2FjdGFibGVJRAisVHJhbnNhY3RhYmxlg6N0Z3SR2TBuZGFkeXJkN3U3a3lqa3E5bndjejNyZ3lpM202ZnllZXh3d2k2aHk2Z2l3Ynk3d3mjc2VxS6NzaWeRkgLERzBFAiEA3uAjXlRkWjOjzywD494+BCNlFRM7e5E9WFs2FMHFnrcCIGQj44m/L7QU4q7qx4DnEkLRAwPCZwG3g3Nni1E8TPjX")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(616678057000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(115528348351), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQvjsqx6zBEem65AJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxFaNzaWeRkgHEQFc5Gxuu34tyWs/0jro+3kvVL031JJseLNBZCxnVqccHLp7xEI6S1BIhXJFFCat/18LDB4L6vVPO9ftEDRDIWgQ=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(617143572000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(115687929080), acct.Balance)
	}
	{
		data, err := base64.StdEncoding.DecodeString("g6VOb25jZcQQY/xeaa1TEemJwQJCrBEAAq5UcmFuc2FjdGFibGVJRAasVHJhbnNhY3RhYmxlg6Nub2SR2TBuZGFtNzVmbmpuN2NkdWVzN2l2aTdjY2ZxOGY1MzRxdWllYWNjcWlicnZ1emhxeGGjc2VxFqNzaWeRkgHEQEoQXAJs13Aml8sFbSJdfRKtOChbLI7tPDr/r8uhSvtuEmIFwt25g1EHQ0zXR7RZZguMj8tZvmPmBW3BYk5uGQU=")
		require.NoError(t, err)
		tx, err := metatx.Unmarshal(data, TxIDs)
		require.NoError(t, err)
		resp, _ := deliverTxContext(t, app, tx, context.at(617206127000000))
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		acct, _ := app.getAccount(addr)
		require.Equal(t, math.Ndau(115709387814), acct.Balance)
	}
}
