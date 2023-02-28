package main

//go:generate swagger generate client -f spec-v1-swagger.json --additional-initialism=OK --additional-initialism=YNAB -O createTransaction -O getBudgets -M Account -M AccountType -M BudgetSummary -M BudgetSummaryResponse -M CurrencyFormat -M DateFormat -M ErrorDetail -M ErrorResponse -M LoanAccountPeriodicValue -M PostTransactionsWrapper -M SaveSubTransaction -M SaveTransaction -M SaveTransactionsResponse -M SaveTransactionWithOptionalFields -M SubTransaction -M TransactionDetail -M TransactionSummary

import (
	"encoding/csv"
	"encoding/json"
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
	color   = flag.String("color", "", "Optional flag color for imported transactions")
	dryRun  = flag.Bool("dry_run", false, "Dry run.")
)

const (
	// Default payee names.
	defaultPayee = "Amazon"
	missingPayee = "Missing"

	// Amazon CSV date format.
	dateFormat = "01/02/06"

	// Amazon URL prefix for order details.
	orderURL = "https://amzn.com/order-details/?orderID="

	// CSV column names.
	orderDate       = "Order Date"
	orderID         = "Order ID"
	orderStatus     = "Order Status"
	shippingCharge  = "Shipping Charge"
	totalPromotions = "Total Promotions"
	taxCharged      = "Tax Charged"
	totalCharged    = "Total Charged"
	title           = "Title"
	seller          = "Seller"
	itemSubtotalTax = "Item Subtotal Tax"
	itemTotal       = "Item Total"

	// Shipped "Order Status" value.
	shipped = "Shipped"
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

	odm, err := parseOrders(*orders)
	if err != nil {
		log.Fatal(err)
	}

	idm, err := parseItems(*items)
	if err != nil {
		log.Fatal(err)
	}

	// Build the transactions.
	data := &models.PostTransactionsWrapper{Transactions: buildTransactions(accountID, odm, idm)}
	if len(data.Transactions) == 0 {
		log.Fatal("nothing to import")
	}

	if *dryRun {
		j, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			log.Fatalf("json.MarshalIndent(): %v", err)
		}
		log.Println(string(j))
		return
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

type orderDetail struct {
	date            *strfmt.Date
	shippingCharge  int64
	totalPromotions int64
	taxCharged      int64
	totalCharged    int64
	items           []*itemDetail
}

func (od *orderDetail) String() string {
	var items string
	for i, id := range od.items {
		items += fmt.Sprintf("\n\t%02d %s", i, id)
	}
	return fmt.Sprintf(
		"Date: %s, Ship: %d, Promo: %d, Tax: %d, Total: %d%s",
		od.date, od.shippingCharge, od.totalPromotions, od.taxCharged, od.totalCharged, items)
}

type itemDetail struct {
	title       string
	seller      string
	subTotalTax int64
	itemTotal   int64
}

func (id *itemDetail) String() string {
	return fmt.Sprintf(
		"Seller: %q, Tax: %d, Total: %d, Title: %q",
		id.seller, id.subTotalTax, id.itemTotal, id.title)
}

// parseOrders parses an Amazon order CSV and returns an orderDetail for each
// order ID.
func parseOrders(name string) (map[string]*orderDetail, error) {
	rows, err := parseCSV(name, orderStatus, orderID, orderDate, shippingCharge, totalPromotions, taxCharged, totalCharged)
	if err != nil {
		return nil, fmt.Errorf("failed to parse orders CSV: %w", err)
	}

	details := make(map[string]*orderDetail)
	for _, row := range rows {
		// Skip orders that haven't shipped.
		if row[orderStatus] != shipped {
			continue
		}

		// Get or add an order record.
		od, err := getOrAddOrder(details, row)
		if err != nil {
			return nil, err
		}

		// Parse the order amounts.
		amounts := []*int64{&od.shippingCharge, &od.totalPromotions, &od.taxCharged, &od.totalCharged}
		for i, col := range []string{shippingCharge, totalPromotions, taxCharged, totalCharged} {
			n, err := parseMoney(row[col], col != totalPromotions)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s %q: %w", col, row[col], err)
			}
			*amounts[i] += n
		}
	}

	return details, nil
}

// parseItems parses an Amazon item CSV and returns an orderDetail for each
// order ID.
func parseItems(name string) (map[string]*orderDetail, error) {
	rows, err := parseCSV(name, orderStatus, orderID, orderDate, title, seller, itemSubtotalTax, itemTotal)
	if err != nil {
		return nil, fmt.Errorf("failed to parse items CSV: %w", err)
	}

	details := make(map[string]*orderDetail)
	for _, row := range rows {
		// Skip items that haven't shipped.
		if row[orderStatus] != shipped {
			continue
		}

		// Get or add an order record.
		od, err := getOrAddOrder(details, row)
		if err != nil {
			return nil, err
		}

		// Create an item record.
		id := &itemDetail{title: row[title], seller: row[seller]}

		// Parse the item amounts.
		amounts := []*int64{&id.subTotalTax, &id.itemTotal}
		for i, col := range []string{itemSubtotalTax, itemTotal} {
			n, err := parseMoney(row[col], true)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s %q: %w", col, row[col], err)
			}
			*amounts[i] = n
		}

		// Add the item and amounts to the order.
		od.taxCharged += id.subTotalTax
		od.totalCharged += id.itemTotal
		od.items = append(od.items, id)
	}

	return details, nil
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

// getOrAddOrder checks for an existing orderDetail and returns it or adds a new
// one and returns it.
func getOrAddOrder(details map[string]*orderDetail, row map[string]string) (*orderDetail, error) {
	// Check if there is already a record for the order ID.
	oid := row[orderID]
	if d, ok := details[oid]; ok {
		return d, nil
	}

	// Get the transaction date.
	date, err := parseDate(row[orderDate])
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s %q: %w", orderDate, row[orderDate], err)
	}

	d := &orderDetail{date: date}
	details[oid] = d
	return d, nil
}

// parseDate returns the YNAB date representation of an Amazon CSV date string.
func parseDate(date string) (*strfmt.Date, error) {
	d, err := time.ParseInLocation(dateFormat, date, time.Local)
	if err != nil {
		return nil, err
	}
	return ptrOf(strfmt.Date(d)), nil
}

// ptrOf returns a pointer to a value of any type.
func ptrOf[T any](v T) *T { return &v }

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

// buildTransactions builds new transactions from order details.
func buildTransactions(accountID *strfmt.UUID, odm, idm map[string]*orderDetail) []*models.SaveTransaction {
	var transactions []*models.SaveTransaction
	for oid, od := range mergeOrders(odm, idm) {
		t := &models.SaveTransaction{
			AccountID: accountID,
			Amount:    &od.totalCharged,
			Date:      od.date,
			SaveTransactionWithOptionalFields: models.SaveTransactionWithOptionalFields{
				Cleared:   models.SaveTransactionWithOptionalFieldsClearedCleared,
				FlagColor: color,
			},
		}
		transactions = append(transactions, t)
		if len(od.items) == 0 {
			// Missing items.
			t.Memo = truncate(orderURL+oid, 200)
			t.PayeeName = truncate(defaultPayee, 50)
			continue
		}
		if len(od.items) == 1 {
			// Single item.
			t.Memo = truncate(od.items[0].title, 200)
			t.PayeeName = truncate(od.items[0].seller, 50)
			continue
		}
		// Apply any promotional amounts to shipping charges.
		if n := od.shippingCharge + od.totalPromotions; n < 0 {
			// Create a subtransaction for the remaining shipping charge.
			t.Subtransactions = append(t.Subtransactions, &models.SaveSubTransaction{
				Amount:    &n,
				Memo:      truncate(shippingCharge, 200),
				PayeeName: truncate(defaultPayee, 50),
			})
		} else if n > 0 {
			// Create a subtransaction for the remaining promo total.
			t.Subtransactions = append(t.Subtransactions, &models.SaveSubTransaction{
				Amount:    &n,
				Memo:      truncate(totalPromotions, 200),
				PayeeName: truncate(defaultPayee, 50),
			})
		}
		// Create subtransactions for each of the order items.
		var multiPayee bool
		for _, id := range od.items {
			payeeName := truncate(id.seller, 50)
			t.Subtransactions = append(t.Subtransactions, &models.SaveSubTransaction{
				Amount:    &id.itemTotal,
				Memo:      truncate(id.title, 200),
				PayeeName: payeeName,
			})
			if multiPayee || payeeName == t.PayeeName {
				continue
			}
			// If all items have the same payee, propagate it to the transaction.
			if multiPayee = t.PayeeName != ""; multiPayee {
				t.PayeeName = ""
			} else {
				t.PayeeName = payeeName
			}
		}
	}
	return transactions
}

// mergeItems merges parsed orders and parsed items.
func mergeOrders(odm, idm map[string]*orderDetail) map[string]*orderDetail {
	for oid, id := range idm {
		od, ok := odm[oid]
		if !ok {
			// No matching order, so just copy the item pseudo-order.
			log.Printf("Missing order: %s, %s\n\n", oid, id)
			odm[oid] = id
			continue
		}

		// Copy over items.
		od.items = id.items

		// If there is more total tax than item tax, assume it is for shipping.
		if st := od.taxCharged - id.taxCharged; st < 0 && od.shippingCharge < 0 {
			od.shippingCharge += st
		}

		// Add an item for any remaining balance.
		if r := od.totalCharged - od.shippingCharge - od.totalPromotions - id.totalCharged; r != 0 {
			od.items = append(od.items, &itemDetail{
				title:     orderURL + oid,
				seller:    missingPayee,
				itemTotal: r,
			})
		}
	}
	return odm
}
