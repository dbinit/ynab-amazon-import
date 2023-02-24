package main

import (
	"flag"
	"log"
	"strings"

	"github.com/dbinit/ynab-amazon-import/client"
	"github.com/dbinit/ynab-amazon-import/client/budgets"
	httptransport "github.com/go-openapi/runtime/client"
)

var (
	token   = flag.String("token", "", "YNAB personal access token")
	budget  = flag.String("budget", "", "YNAB budget name")
	account = flag.String("account", "", "YNAB account name")
)

func main() {
	flag.Parse()

	var missing []string
	for _, n := range []string{"token", "budget", "account"} {
		if f := flag.Lookup(n); f == nil || f.Value.String() == "" {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		log.Fatalf("Required flag(s): %v.", missing)
	}

	auth := httptransport.BearerToken(*token)

	params := budgets.NewGetBudgetsParams().WithIncludeAccounts(boolPtr(true))
	budgets, err := client.Default.Budgets.GetBudgets(params, auth)
	if err != nil || budgets == nil {
		log.Fatalf("GetBudgets(): %v.", err)
	}

	var budgetID, accountID string
	for _, b := range budgets.Payload.Data.Budgets {
		if b == nil || b.ID == nil || b.Name == nil || !strings.EqualFold(*b.Name, *budget) {
			continue
		}
		log.Printf("Budget %q found with ID %q.", *b.Name, b.ID)
		budgetID = b.ID.String()
		for _, a := range b.Accounts {
			if a == nil || a.ID == nil || a.Name == nil || !strings.EqualFold(*a.Name, *account) {
				continue
			}
			log.Printf("Account %q found with ID %q.", *a.Name, a.ID)
			accountID = a.ID.String()
			break
		}
		break
	}
	if budgetID == "" {
		log.Fatalf("Budget %q not found.", *budget)
	}
	if accountID == "" {
		log.Fatalf("Account %q not found in budget %q.", *account, *budget)
	}
}

func boolPtr(v bool) *bool { return &v }
