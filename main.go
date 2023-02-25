package main

//go:generate swagger generate client -f spec-v1-swagger.json --additional-initialism=OK --additional-initialism=YNAB -O createTransaction -O getBudgets -M Account -M AccountType -M BudgetSummary -M BudgetSummaryResponse -M CurrencyFormat -M DateFormat -M ErrorDetail -M ErrorResponse -M LoanAccountPeriodicValue -M PostTransactionsWrapper -M SaveSubTransaction -M SaveTransaction -M SaveTransactionsResponse -M SaveTransactionWithOptionalFields -M SubTransaction -M TransactionDetail -M TransactionSummary

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dbinit/ynab-amazon-import/client"
	"github.com/dbinit/ynab-amazon-import/client/budgets"
	"github.com/dbinit/ynab-amazon-import/client/transactions"
	"github.com/dbinit/ynab-amazon-import/models"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

var (
	token   = flag.String("token", "", "YNAB personal access token")
	budget  = flag.String("budget", "", "YNAB budget name")
	account = flag.String("account", "", "YNAB account name")
	orders  = flag.String("orders", "", "Amazon orders CSV file")
	items   = flag.String("items", "", "Amazon items CSV file")
)

const (
	// Default payee.
	payeeName = "Amazon"

	// Amazon CSV date format.
	dateFormat = "01/02/06"

	// Amazon URL prefix for order details.
	orderURL = "https://amzn.com/order-details/?orderID="

	// CSV column names.
	orderDate       = "Order Date"
	orderID         = "Order ID"
	shippingCharge  = "Shipping Charge"
	totalPromotions = "Total Promotions"
	totalCharged    = "Total Charged"
	title           = "Title"
	seller          = "Seller"
	itemTotal       = "Item Total"
)

func main() {
	flag.Parse()

	// Make sure required flags are provided.
	var missing []string
	for _, n := range []string{"token", "budget", "account", "orders", "items"} {
		if f := flag.Lookup(n); f == nil || f.Value.String() == "" {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		log.Fatalf("missing required flag(s): %v", missing)
	}

	authInfo := httptransport.BearerToken(*token)
	budgetID, accountID, err := budgetAccount(*budget, *account, authInfo)
	if err != nil {
		log.Fatal(err)
	}

	otm, err := orderTransactions(*orders, accountID)
	if err != nil {
		log.Fatal(err)
	}

	itm, err := itemTransactions(*items, accountID)
	if err != nil {
		log.Fatal(err)
	}

	for oid, s := range itm {
		if t, ok := otm[oid]; ok {
			// Append the subtransactions to an existing transaction.
			t.Subtransactions = append(t.Subtransactions, s.Subtransactions...)
			continue
		}
		// Copy the full item transaction if an order transaction is missing.
		otm[oid] = s
	}

	data := &models.PostTransactionsWrapper{}
	for oid, t := range otm {
		data.Transactions = append(data.Transactions, t)
		if len(t.Subtransactions) == 0 {
			continue
		}
		// Make sure the subtransactions add up to the transaction total.
		amount := *t.Amount
		for _, s := range t.Subtransactions {
			amount -= *s.Amount
		}
		if amount != 0 {
			t.Subtransactions = append(t.Subtransactions, &models.SaveSubTransaction{
				Amount:    &amount,
				Memo:      orderURL + oid,
				PayeeName: payeeName,
			})
		}
		if len(t.Subtransactions) > 1 {
			continue
		}
		// Collapse single item orders.
		t.Memo = t.Subtransactions[0].Memo
		t.PayeeName = t.Subtransactions[0].PayeeName
		t.Subtransactions = nil
	}

	params := transactions.NewCreateTransactionParams().WithBudgetID(budgetID.String()).WithData(data)
	if _, err = client.Default.Transactions.CreateTransaction(params, authInfo); err != nil {
		log.Fatalf("CreateTransaction(): %v", err)
	}
}

// budgetAccount finds the named budget and account and returns the IDs.
func budgetAccount(budgetName, accountName string, authInfo runtime.ClientAuthInfoWriter) (*strfmt.UUID, *strfmt.UUID, error) {
	params := budgets.NewGetBudgetsParams().WithIncludeAccounts(ptrOf(true))
	budgets, err := client.Default.Budgets.GetBudgets(params, authInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("GetBudgets(): %w", err)
	}
	if budgets == nil || budgets.Payload == nil || budgets.Payload.Data == nil {
		return nil, nil, fmt.Errorf("GetBudgets(): %+v", budgets)
	}

	var bid, aid *strfmt.UUID
	for _, b := range budgets.Payload.Data.Budgets {
		if b == nil || b.ID == nil || b.Name == nil || !strings.EqualFold(*b.Name, budgetName) {
			continue
		}
		bid, budgetName = b.ID, *b.Name
		log.Printf("budget %q found with ID %s", budgetName, bid)

		for _, a := range b.Accounts {
			if a == nil || a.ID == nil || a.Name == nil || !strings.EqualFold(*a.Name, accountName) {
				continue
			}
			aid, accountName = a.ID, *a.Name
			log.Printf("account %q found with ID %s", budgetName, bid)
			break
		}

		break
	}
	if bid == nil {
		return nil, nil, fmt.Errorf("budget %q not found", budgetName)
	}
	if aid == nil {
		return nil, nil, fmt.Errorf("account %q not found in budget %q", accountName, budgetName)
	}

	return bid, aid, nil
}

// ptrOf returns a pointer to a value of any type.
func ptrOf[T any](v T) *T { return &v }

// orderTransactions parses an Amazon order CSV and returns a transaction for
// each order ID.
func orderTransactions(name string, accountID *strfmt.UUID) (map[string]*models.SaveTransaction, error) {
	rows, err := parseCSV(name, orderDate, orderID, shippingCharge, totalPromotions, totalCharged)
	if err != nil {
		return nil, fmt.Errorf("failed to parse orders CSV: %w", err)
	}

	otm := make(map[string]*models.SaveTransaction)
	for _, r := range rows {
		// Get the transaction amount.
		amount, err := parseMoney(r[totalCharged], true)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s %q: %w", totalCharged, r[totalCharged], err)
		}

		// Get the transaction date.
		date, err := parseDate(r[orderDate])
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s %q: %w", orderDate, r[orderDate], err)
		}

		// Add subtransactions for any shipping charges or promo discounts.
		var subs []*models.SaveSubTransaction
		for _, c := range []string{shippingCharge, totalPromotions} {
			if amount, err := parseMoney(r[c], c != totalPromotions); err != nil {
				return nil, fmt.Errorf("failed to parse %s %q: %w", c, r[c], err)
			} else if amount != 0 {
				subs = append(subs, &models.SaveSubTransaction{
					Amount:    &amount,
					Memo:      c,
					PayeeName: payeeName,
				})
			}
		}

		// Build the transaction.
		otm[r[orderID]] = &models.SaveTransaction{
			AccountID: accountID,
			Amount:    &amount,
			Date:      date,
			SaveTransactionWithOptionalFields: models.SaveTransactionWithOptionalFields{
				Cleared:         models.SaveTransactionWithOptionalFieldsClearedCleared,
				PayeeName:       payeeName,
				Subtransactions: subs,
			},
		}
	}

	return otm, nil
}

// itemTransactions parses an Amazon item CSV and returns a transaction for
// each order ID.
func itemTransactions(name string, accountID *strfmt.UUID) (map[string]*models.SaveTransaction, error) {
	rows, err := parseCSV(name, orderDate, orderID, title, seller, itemTotal)
	if err != nil {
		return nil, fmt.Errorf("failed to parse items CSV: %w", err)
	}

	itm := make(map[string]*models.SaveTransaction)
	for _, r := range rows {
		// Get the subtransaction amount.
		amount, err := parseMoney(r[itemTotal], true)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s %q: %w", itemTotal, r[itemTotal], err)
		}

		// Build the subtransaction.
		s := &models.SaveSubTransaction{
			Amount:    &amount,
			Memo:      truncate(r[title], 200),
			PayeeName: truncate(r[seller], 50),
		}

		// Add to an existing transaction.
		oid := r[orderID]
		if t, ok := itm[oid]; ok {
			*t.Amount += amount
			t.Subtransactions = append(t.Subtransactions, s)
			continue
		}

		// Get the transaction date.
		date, err := parseDate(r[orderDate])
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s %q: %w", orderDate, r[orderDate], err)
		}

		// Build a new transaction.
		itm[oid] = &models.SaveTransaction{
			AccountID: accountID,
			Amount:    ptrOf(amount), // Don't reuse the subtransaction pointer.
			Date:      date,
			SaveTransactionWithOptionalFields: models.SaveTransactionWithOptionalFields{
				Cleared:         models.SaveTransactionWithOptionalFieldsClearedCleared,
				PayeeName:       payeeName,
				Subtransactions: []*models.SaveSubTransaction{s},
			},
		}
	}

	return itm, nil
}

// parseCSV parses a CSV file and extracts named columns into a string map for
// each row.
func parseCSV(name string, cols ...string) (rows []map[string]string, err error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("os.Open(%q): %w", name, err)
	}
	defer func() {
		// Do we care about close errors?
		if ferr := f.Close(); ferr != nil && err == nil {
			err = fmt.Errorf("(os.File).Close(%q): %w", name, ferr)
		}
	}()
	reader := csv.NewReader(f)

	// Read the header row.
	row, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("(csv.Reader).Read(%q) header: %w", name, err)
	}

	// Build a column index map.
	sort.Strings(cols)
	colm := make(map[int]string)
	for i, c := range row {
		// Find each column name. Column names must match exactly.
		if j := sort.SearchStrings(cols, c); j < len(cols) && cols[j] == c {
			cols = append(cols[:j], cols[j+1:]...)
			colm[i] = c
		}
	}
	if len(cols) > 0 {
		return nil, fmt.Errorf("missing columns in %q: %v", name, cols)
	}

	// Build row maps of column names to values.
	for row, err = reader.Read(); err == nil; row, err = reader.Read() {
		rm := make(map[string]string)
		for i, c := range colm {
			if len(row) > i {
				rm[c] = row[i]
			}
		}
		rows = append(rows, rm)
	}
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("(csv.Reader).Read(%q) rows: %w", name, err)
	}

	return rows, nil
}

// parseDate returns the YNAB date representation of an Amazon CSV date string.
func parseDate(date string) (*strfmt.Date, error) {
	d, err := time.ParseInLocation(dateFormat, date, time.Local)
	if err != nil {
		return nil, err
	}
	return ptrOf(strfmt.Date(d)), nil
}

// parseMoney returns the YNAB int64 representation of an Amazon CSV currency
// string. E.g. "$12.34" becomes 12340.
func parseMoney(amount string, invert bool) (int64, error) {
	m := moneyRE.FindStringSubmatch(amount)
	if m == nil || len(m) != 4 {
		return 0, fmt.Errorf("failed to parse %q as money", amount)
	}
	// Make sure decimal has 3 digits.
	a, err := strconv.ParseInt(m[1]+m[2]+(m[3] + "000")[:3], 10, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %q as money: %w", amount, err)
	}
	if invert {
		return a * -1, nil
	}
	return a, nil
}

// Matches a currency string, e.g. "$12.34".
var moneyRE = regexp.MustCompile(`^([-])?[^\d]*(\d+)(?:[.](\d+))?$`)

// truncate as string to a maximum length.
func truncate(s string, l int) string {
	if l <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) < l {
		return s
	}
	return string([]rune(s)[:l])
}
