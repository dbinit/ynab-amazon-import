// Code generated by go-swagger; DO NOT EDIT.

package budgets

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/dbinit/ynab-amazon-import/models"
)

// GetBudgetsReader is a Reader for the GetBudgets structure.
type GetBudgetsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetBudgetsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetBudgetsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewGetBudgetsNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewGetBudgetsDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetBudgetsOK creates a GetBudgetsOK with default headers values
func NewGetBudgetsOK() *GetBudgetsOK {
	return &GetBudgetsOK{}
}

/*
GetBudgetsOK describes a response with status code 200, with default header values.

The list of budgets
*/
type GetBudgetsOK struct {
	Payload *models.BudgetSummaryResponse
}

// IsSuccess returns true when this get budgets Ok response has a 2xx status code
func (o *GetBudgetsOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get budgets Ok response has a 3xx status code
func (o *GetBudgetsOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get budgets Ok response has a 4xx status code
func (o *GetBudgetsOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get budgets Ok response has a 5xx status code
func (o *GetBudgetsOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get budgets Ok response a status code equal to that given
func (o *GetBudgetsOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the get budgets Ok response
func (o *GetBudgetsOK) Code() int {
	return 200
}

func (o *GetBudgetsOK) Error() string {
	return fmt.Sprintf("[GET /budgets][%d] getBudgetsOk  %+v", 200, o.Payload)
}

func (o *GetBudgetsOK) String() string {
	return fmt.Sprintf("[GET /budgets][%d] getBudgetsOk  %+v", 200, o.Payload)
}

func (o *GetBudgetsOK) GetPayload() *models.BudgetSummaryResponse {
	return o.Payload
}

func (o *GetBudgetsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.BudgetSummaryResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetBudgetsNotFound creates a GetBudgetsNotFound with default headers values
func NewGetBudgetsNotFound() *GetBudgetsNotFound {
	return &GetBudgetsNotFound{}
}

/*
GetBudgetsNotFound describes a response with status code 404, with default header values.

No budgets were found
*/
type GetBudgetsNotFound struct {
	Payload *models.ErrorResponse
}

// IsSuccess returns true when this get budgets not found response has a 2xx status code
func (o *GetBudgetsNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get budgets not found response has a 3xx status code
func (o *GetBudgetsNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get budgets not found response has a 4xx status code
func (o *GetBudgetsNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get budgets not found response has a 5xx status code
func (o *GetBudgetsNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get budgets not found response a status code equal to that given
func (o *GetBudgetsNotFound) IsCode(code int) bool {
	return code == 404
}

// Code gets the status code for the get budgets not found response
func (o *GetBudgetsNotFound) Code() int {
	return 404
}

func (o *GetBudgetsNotFound) Error() string {
	return fmt.Sprintf("[GET /budgets][%d] getBudgetsNotFound  %+v", 404, o.Payload)
}

func (o *GetBudgetsNotFound) String() string {
	return fmt.Sprintf("[GET /budgets][%d] getBudgetsNotFound  %+v", 404, o.Payload)
}

func (o *GetBudgetsNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetBudgetsNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetBudgetsDefault creates a GetBudgetsDefault with default headers values
func NewGetBudgetsDefault(code int) *GetBudgetsDefault {
	return &GetBudgetsDefault{
		_statusCode: code,
	}
}

/*
GetBudgetsDefault describes a response with status code -1, with default header values.

An error occurred
*/
type GetBudgetsDefault struct {
	_statusCode int

	Payload *models.ErrorResponse
}

// IsSuccess returns true when this get budgets default response has a 2xx status code
func (o *GetBudgetsDefault) IsSuccess() bool {
	return o._statusCode/100 == 2
}

// IsRedirect returns true when this get budgets default response has a 3xx status code
func (o *GetBudgetsDefault) IsRedirect() bool {
	return o._statusCode/100 == 3
}

// IsClientError returns true when this get budgets default response has a 4xx status code
func (o *GetBudgetsDefault) IsClientError() bool {
	return o._statusCode/100 == 4
}

// IsServerError returns true when this get budgets default response has a 5xx status code
func (o *GetBudgetsDefault) IsServerError() bool {
	return o._statusCode/100 == 5
}

// IsCode returns true when this get budgets default response a status code equal to that given
func (o *GetBudgetsDefault) IsCode(code int) bool {
	return o._statusCode == code
}

// Code gets the status code for the get budgets default response
func (o *GetBudgetsDefault) Code() int {
	return o._statusCode
}

func (o *GetBudgetsDefault) Error() string {
	return fmt.Sprintf("[GET /budgets][%d] getBudgets default  %+v", o._statusCode, o.Payload)
}

func (o *GetBudgetsDefault) String() string {
	return fmt.Sprintf("[GET /budgets][%d] getBudgets default  %+v", o._statusCode, o.Payload)
}

func (o *GetBudgetsDefault) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetBudgetsDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
