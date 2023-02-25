// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// SaveTransactionsResponse save transactions response
//
// swagger:model SaveTransactionsResponse
type SaveTransactionsResponse struct {

	// data
	// Required: true
	Data *SaveTransactionsResponseData `json:"data"`
}

// Validate validates this save transactions response
func (m *SaveTransactionsResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateData(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SaveTransactionsResponse) validateData(formats strfmt.Registry) error {

	if err := validate.Required("data", "body", m.Data); err != nil {
		return err
	}

	if m.Data != nil {
		if err := m.Data.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("data")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("data")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this save transactions response based on the context it is used
func (m *SaveTransactionsResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateData(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SaveTransactionsResponse) contextValidateData(ctx context.Context, formats strfmt.Registry) error {

	if m.Data != nil {
		if err := m.Data.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("data")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("data")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *SaveTransactionsResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *SaveTransactionsResponse) UnmarshalBinary(b []byte) error {
	var res SaveTransactionsResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// SaveTransactionsResponseData save transactions response data
//
// swagger:model SaveTransactionsResponseData
type SaveTransactionsResponseData struct {

	// If multiple transactions were specified, a list of import_ids that were not created because of an existing `import_id` found on the same account
	DuplicateImportIds []string `json:"duplicate_import_ids"`

	// The knowledge of the server
	// Required: true
	ServerKnowledge *int64 `json:"server_knowledge"`

	// If a single transaction was specified, the transaction that was saved
	Transaction *TransactionDetail `json:"transaction,omitempty"`

	// The transaction ids that were saved
	// Required: true
	TransactionIds []string `json:"transaction_ids"`

	// If multiple transactions were specified, the transactions that were saved
	Transactions []*TransactionDetail `json:"transactions"`
}

// Validate validates this save transactions response data
func (m *SaveTransactionsResponseData) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateServerKnowledge(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTransaction(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTransactionIds(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTransactions(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SaveTransactionsResponseData) validateServerKnowledge(formats strfmt.Registry) error {

	if err := validate.Required("data"+"."+"server_knowledge", "body", m.ServerKnowledge); err != nil {
		return err
	}

	return nil
}

func (m *SaveTransactionsResponseData) validateTransaction(formats strfmt.Registry) error {
	if swag.IsZero(m.Transaction) { // not required
		return nil
	}

	if m.Transaction != nil {
		if err := m.Transaction.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("data" + "." + "transaction")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("data" + "." + "transaction")
			}
			return err
		}
	}

	return nil
}

func (m *SaveTransactionsResponseData) validateTransactionIds(formats strfmt.Registry) error {

	if err := validate.Required("data"+"."+"transaction_ids", "body", m.TransactionIds); err != nil {
		return err
	}

	return nil
}

func (m *SaveTransactionsResponseData) validateTransactions(formats strfmt.Registry) error {
	if swag.IsZero(m.Transactions) { // not required
		return nil
	}

	for i := 0; i < len(m.Transactions); i++ {
		if swag.IsZero(m.Transactions[i]) { // not required
			continue
		}

		if m.Transactions[i] != nil {
			if err := m.Transactions[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("data" + "." + "transactions" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("data" + "." + "transactions" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this save transactions response data based on the context it is used
func (m *SaveTransactionsResponseData) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateTransaction(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateTransactions(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SaveTransactionsResponseData) contextValidateTransaction(ctx context.Context, formats strfmt.Registry) error {

	if m.Transaction != nil {
		if err := m.Transaction.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("data" + "." + "transaction")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("data" + "." + "transaction")
			}
			return err
		}
	}

	return nil
}

func (m *SaveTransactionsResponseData) contextValidateTransactions(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Transactions); i++ {

		if m.Transactions[i] != nil {
			if err := m.Transactions[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("data" + "." + "transactions" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("data" + "." + "transactions" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *SaveTransactionsResponseData) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *SaveTransactionsResponseData) UnmarshalBinary(b []byte) error {
	var res SaveTransactionsResponseData
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}