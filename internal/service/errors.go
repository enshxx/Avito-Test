package service

import "fmt"

var (
	ErrCannotCreateTender              = fmt.Errorf("cannot create tender")
	ErrEmployeeDoesNotExist            = fmt.Errorf("employee doens`t exist")
	ErrCannotGetTender                 = fmt.Errorf("can not get tenders")
	ErrCannotGetStatus                 = fmt.Errorf("can not get status")
	ErrPermissionDenied                = fmt.Errorf("pemission denied")
	ErrTenderNotFound                  = fmt.Errorf("tender not found")
	ErrCannotPutStatus                 = fmt.Errorf("can not put status")
	ErrCannotEditTender                = fmt.Errorf("can not edit tender")
	ErrVersionNotFound                 = fmt.Errorf("version not found")
	ErrCannotCreateBid                 = fmt.Errorf("cannot create bid")
	ErrOrganisationResponsibleNotFound = fmt.Errorf("organisation responsible not found")
	ErrCannotGetBids                   = fmt.Errorf("can not get bid")
	ErrBidNotFound                     = fmt.Errorf("bid not found")
	ErrCannotEditBid                   = fmt.Errorf("can not edit bid")
)
