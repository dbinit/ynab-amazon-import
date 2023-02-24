#!/bin/bash
swagger generate client -f spec-v1-swagger.json --additional-initialism=YNAB --strict-responders --keep-spec-order \
	-O getBudgets \
	-M Account \
	-M AccountType \
	-M BudgetSummary \
	-M BudgetSummaryResponse \
	-M CurrencyFormat \
	-M DateFormat \
	-M ErrorDetail \
	-M ErrorResponse \
	-M LoanAccountPeriodicValue
go mod tidy
