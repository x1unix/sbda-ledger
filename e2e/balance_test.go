package e2e

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/x1unix/sbda-ledger/internal/model/loan"
	"github.com/x1unix/sbda-ledger/internal/model/user"
	"github.com/x1unix/sbda-ledger/pkg/ledger"
)

func TestBalance(t *testing.T) {
	require.NoError(t, TruncateData(), "failed to truncate data before test")

	// For example, Alice, Bob, and Charlie are in the same group, "Friends".
	alice := mustCreateUser(t, "alice", "alice@mail.com")
	bob := mustCreateUser(t, "bob", "bob@mail.com")
	charlie := mustCreateUser(t, "charlie", "charlie@mail.com")

	gFriends, err := Client.CreateGroup("friend", alice.Token)
	require.NoError(t, err, "failed to create Friends group")

	require.NoError(t, Client.AddGroupMembers(gFriends.ID, alice.Token, bob.User.ID, charlie.User.ID),
		"failed to add members to a group Friends")

	// hey are out for dinner, and Alice pays the bill.
	// The bill has to be split equally between them, so she enters a new expense named "Pizza!"
	// of €42 to the "Friends" group.
	require.NoError(t, Client.AddGroupExpense(gFriends.ID, 4200, alice.Token), "failed to add an pizza expense")

	// When Alice checks her balance, the system will let her know that Bob owes her €14 and Charlie owes her €14 as well.
	b, err := Client.Balance(alice.Token)
	require.NoError(t, err, "failed to get Alice's balance")
	expectAliceBalance := map[string]int64{
		bob.User.ID:     1400,
		charlie.User.ID: 1400,
	}

	// Check Alice balance
	aliceBalance := balanceListToMap(b)
	require.Equal(t, expectAliceBalance, aliceBalance, "unexpected alice balance from response")
	checkDatabaseAndCacheBalance(t, alice.User.ID, expectAliceBalance)

	// Check balance of Bob and Charlie
	expectedRestBalance := map[string]int64{
		alice.User.ID: -1400,
	}
	for _, u := range []*ledger.LoginResponse{bob, charlie} {
		b, err = Client.Balance(u.Token)
		require.NoError(t, err, "failed to get balance of", u.User.Name, u.User.ID)

		balanceMap := balanceListToMap(b)
		require.Equalf(t, expectedRestBalance, balanceMap, "unexpected %s's balance from response", u.User.Name)
		checkDatabaseAndCacheBalance(t, u.User.ID, expectedRestBalance)
	}

	// The day after, Alice and Bob are out for a coffee, and Bob pays the bill, €8.
	// He enters the new expense in the system.

	// Create a group for a second test story.
	gCoffee, err := Client.CreateGroup("hot coffee", alice.Token)
	require.NoError(t, err, "failed to create a coffee group")
	require.NoError(t, Client.AddGroupMembers(gCoffee.ID, alice.Token, bob.User.ID))

	// Put €8 bill for Bob and Alice
	require.NoError(t, Client.AddGroupExpense(gCoffee.ID, 800, bob.Token))

	// When Alice checks her balance again, the system will let her know
	// that Bob owes her €10, being the simplified debit of €14 he owed from
	// the pizza minus €4 he borrowed when paying for the coffee.
	expectAliceBalance = map[string]int64{
		bob.User.ID:     1000,
		charlie.User.ID: 1400,
	}
	b, err = Client.Balance(alice.Token)
	require.NoError(t, err, "failed to get Alice's balance")
	aliceBalance = balanceListToMap(b)
	require.Equal(t, expectAliceBalance, aliceBalance, "unexpected alice balance from response")
	checkDatabaseAndCacheBalance(t, alice.User.ID, expectAliceBalance)

	// Check Bob's balance.
	// Bob's balance state also should be synced in cache after an update.
	b, err = Client.Balance(bob.Token)
	require.NoError(t, err, "failed to get balance of bob", bob.User.ID)
	expectedBobBalance := map[string]int64{
		alice.User.ID: -1000,
	}
	balanceMap := balanceListToMap(b)
	require.Equalf(t, expectedBobBalance, balanceMap, "unexpected %s's balance from response", bob.User.Name)
	checkDatabaseAndCacheBalance(t, bob.User.ID, expectedBobBalance)
}

func balanceListToMap(l []ledger.Balance) map[string]int64 {
	out := make(map[string]int64, len(l))
	for _, v := range l {
		out[v.UserID] = v.Balance
	}
	return out
}

func checkDatabaseAndCacheBalance(t *testing.T, uid string, expect map[string]int64) {
	checkCacheBalance(t, uid, expect)
	checkCacheBalance(t, uid, expect)
}

func checkDatabaseBalance(t *testing.T, uid string, expect map[string]int64) {
	var out []loan.Balance
	const query = "SELECT user_id, SUM(amount) as balance FROM (" +
		"SELECT debtor_id AS user_id, amount FROM loans WHERE lender_id = $1" +
		" UNION ALL " +
		"SELECT lender_id AS user_id, amount * -1 FROM loans WHERE debtor_id = $1" +
		") balance GROUP BY user_id"
	err := DB.Select(&out, query, uid)
	require.NoError(t, err, "failed to calculate balance of user", uid)
	got := make(map[string]int64, len(out))
	for _, v := range out {
		got[user.IDToString(v.UserID)] = v.Balance
	}
	require.Equal(t, expect, got, "mismatch between DB and expected balance")
}

func checkCacheBalance(t *testing.T, uid string, expect map[string]int64) {
	// assert that cache flag is set
	exits, err := Redis.Exists(context.Background(), "cached:"+uid).Result()
	require.NoError(t, err, "failed to check if cache flag is set")
	if exits == 0 {
		t.Fatal("cache flag is unset for user", uid)
	}

	// compare cached values with expected
	kv, err := Redis.HGetAll(context.Background(), "balance:"+uid).Result()
	require.NoError(t, err, "failed to get keys from Redis")
	got := make(map[string]int64, len(kv))
	for debtorID, val := range kv {
		balance, err := strconv.ParseInt(val, 10, 64)
		require.NoError(t, err, "failed to parse balance value for debtor", balance)
		got[debtorID] = balance
	}
	require.Equal(t, expect, got, "mismatch between Redis and expected balance")
}
