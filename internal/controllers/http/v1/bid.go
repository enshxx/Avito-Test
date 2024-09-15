package v1

import (
	errors2 "avito/internal/controllers/http/errors"
	"avito/internal/controllers/http/formating"
	tenders "avito/internal/controllers/http/parser"
	"avito/internal/controllers/validators"
	"avito/internal/service"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

type bidRoutes struct {
	bidService      service.Bid
	employeeService service.Employee
	tenderService   service.Tender
}

func newBidRoutes(g *echo.Group, bidService service.Bid, employeeService service.Employee, tenderService service.Tender) {
	r := &bidRoutes{
		bidService:      bidService,
		employeeService: employeeService,
		tenderService:   tenderService,
	}
	g.POST("/new", r.create)
	g.GET("/my", r.getMyBids)
	g.GET("/:bid_id/status", r.getStatus)
	g.PUT("/:bid_id/status", r.putStatus)
	g.POST("/:bid_id/edit", r.editBid)
	g.PUT("/:bid_id/rollback/:version", r.rollback)
	g.PUT("/:bid_id/submit_decision", r.submitDecision)
}

type CreateBidInput struct {
	Name        string    `json:"name" validate:"required,max=100"`
	Description string    `json:"description" validate:"required,max=500"`
	TenderId    uuid.UUID `json:"tenderId" validate:"required"`
	AuthorType  string    `json:"authorType" validate:"required,oneof=User Organization"`
	AuthorId    uuid.UUID `json:"authorId" validate:"required"`
}

func (r *bidRoutes) create(c echo.Context) error {
	var input CreateBidInput

	if err := c.Bind(&input); err != nil {
		var httpErr *echo.HTTPError
		if ok := errors.As(err, &httpErr); ok {
			return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
		}
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	if err := c.Validate(input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	_, err := r.employeeService.GetEmployeeById(c.Request().Context(), input.AuthorId)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}
	if input.AuthorType == "Organization" {
		_, err := r.employeeService.GetEmployeeOrgIdById(c.Request().Context(), input.AuthorId)
		if err != nil {
			return errors2.NewErrorResponse(c, http.StatusForbidden, err)
		}
	}
	bid, err := r.bidService.CreateBid(c.Request().Context(), service.BidCreateInput{
		Name:        input.Name,
		Description: input.Description,
		TenderId:    input.TenderId,
		AuthorType:  input.AuthorType,
		AuthorId:    input.AuthorId,
	})
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusInternalServerError, err)
	}

	type response struct {
		Id         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		Status     string    `json:"status"`
		AuthorType string    `json:"authorType"`
		AuthorId   uuid.UUID `json:"authorId"`
		Version    int       `json:"version"`
		CreatedAt  string    `json:"createdAt"`
	}

	return c.JSON(http.StatusOK, response{
		Id:         bid.Id,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorId:   bid.AuthorId,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt.Format(formating.TimeFormat),
	})
}

type GetMyBidsInput struct {
	Username string `query:"username" validate:"required"`
	Limit    int    `query:"limit"`
	Offset   int    `query:"offset"`
}

func (r *bidRoutes) getMyBids(c echo.Context) error {
	var input GetMyBidsInput
	if err := c.Bind(&input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := c.Validate(input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	rawQuery := c.Request().URL.RawQuery
	limit, offset, err := tenders.ParseLimitOffset(rawQuery)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	authorId, err := r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.Username)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}
	response, err := r.bidService.GetMyBids(c.Request().Context(), service.GetMyBidsInput{
		AuthorId: authorId,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, response)
}

type GetBidsInput struct {
	TenderId uuid.UUID `param:"tender_id" validate:"required"`
	Username string    `query:"username" validate:"required"`
	Limit    int       `query:"limit"`
	Offset   int       `query:"offset"`
}

type GetStatusInput struct {
	BidId    uuid.UUID `param:"bid_id" validate:"required"`
	Username string    `query:"username" validate:"required"`
}

func (r *bidRoutes) getStatus(c echo.Context) error {
	var input GetStatusInput
	if err := c.Bind(&input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := c.Validate(input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	bid, err := r.bidService.GetBidById(c.Request().Context(), input.BidId)

	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusNotFound, err)
	}
	employeeId, err := r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.Username)

	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}
	switch bid.AuthorType {
	case "User":
		if bid.AuthorId != employeeId {
			return errors2.NewErrorResponse(c, http.StatusForbidden, fmt.Errorf("permission denied"))
		}
	case "Organization":
		organizationEmployeeId, err := r.employeeService.GetEmployeeOrgIdById(c.Request().Context(), employeeId)
		if err != nil {
			return errors2.NewErrorResponse(c, http.StatusForbidden, fmt.Errorf("permission denied"))
		}
		organizationAuthorId, err := r.employeeService.GetEmployeeOrgIdById(c.Request().Context(), bid.AuthorId)
		if err != nil {
			return errors2.NewErrorResponse(c, http.StatusForbidden, fmt.Errorf("permission denied"))
		}
		if organizationAuthorId != organizationEmployeeId {
			fmt.Println(employeeId)
			return errors2.NewErrorResponse(c, http.StatusForbidden, fmt.Errorf("permission denied"))
		}
	}

	response, err := r.bidService.GetStatus(c.Request().Context(), service.GetBidStatusInput{
		BidId:    input.BidId,
		Username: input.Username,
	})
	if err != nil {
		if errors.Is(err, service.ErrBidNotFound) {
			return errors2.NewErrorResponse(c, http.StatusNotFound, err)
		}
	}
	return c.JSON(http.StatusOK, response)
}

type PutBidStatusInput struct {
	BidId    uuid.UUID `param:"bid_id" validate:"required"`
	Username string    `query:"username" validate:"required"`
	Status   string    `query:"status" validate:"required,oneof=Created Published Canceled"`
}

func (r *bidRoutes) putStatus(c echo.Context) error {
	var input PutBidStatusInput
	if err := c.Bind(&input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	b := echo.DefaultBinder{}
	if err := b.BindQueryParams(c, &input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	if err := c.Validate(input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	bid, err := r.bidService.GetBidById(c.Request().Context(), input.BidId)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusNotFound, err)
	}

	employeeId, err := r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.Username)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}

	switch bid.AuthorType {
	case "User":
		if bid.AuthorId != employeeId {
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
	case "Organization":
		organizationEmployeeId, err := r.employeeService.GetEmployeeOrgIdById(c.Request().Context(), employeeId)
		if err != nil {
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
		organizationAuthorId, err := r.employeeService.GetEmployeeOrgIdById(c.Request().Context(), bid.AuthorId)
		if err != nil {
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
		if organizationAuthorId != organizationEmployeeId {
			fmt.Println(employeeId)
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
	}

	output, err := r.bidService.PutStatus(c.Request().Context(), service.PutBidStatusInput{
		BidId:  input.BidId,
		Status: input.Status,
	})

	if err != nil {
		if errors.Is(err, service.ErrTenderNotFound) {
			return errors2.NewErrorResponse(c, http.StatusNotFound, err)
		}

		if errors.Is(err, service.ErrPermissionDenied) {
			return errors2.NewErrorResponse(c, http.StatusForbidden, err)
		}
		return errors2.NewErrorResponse(c, http.StatusInternalServerError, err)
	}
	type response struct {
		Id         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		Status     string    `json:"status"`
		AuthorType string    `json:"authorType"`
		AuthorId   uuid.UUID `json:"authorId"`
		Version    int       `json:"version"`
		CreatedAt  string    `json:"createdAt"`
	}

	return c.JSON(http.StatusOK, response{
		Id:         output.Id,
		Name:       output.Name,
		Status:     output.Status,
		AuthorType: output.AuthorType,
		AuthorId:   output.AuthorId,
		Version:    output.Version,
		CreatedAt:  output.CreatedAt.Format(formating.TimeFormat),
	})
}

type EditBidInput struct {
	BidId       uuid.UUID `param:"bid_id"`
	Username    string    `query:"username" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty"`
	Description *string   `json:"description" validate:"omitempty"`
}

func (r *bidRoutes) editBid(c echo.Context) error {
	var input EditBidInput
	if err := c.Bind(&input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	b := echo.DefaultBinder{}
	if err := b.BindQueryParams(c, &input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	if err := c.Validate(input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	if err := validators.EditTenderValidate(input.Name, input.Description); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	bid, err := r.bidService.GetBidById(c.Request().Context(), input.BidId)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusNotFound, err)
	}

	employeeId, err := r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.Username)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}

	switch bid.AuthorType {
	case "User":
		if bid.AuthorId != employeeId {
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
	case "Organization":
		organizationEmployeeId, err := r.employeeService.GetEmployeeOrgIdById(c.Request().Context(), employeeId)
		if err != nil {
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
		organizationAuthorId, err := r.employeeService.GetEmployeeOrgIdById(c.Request().Context(), bid.AuthorId)
		if err != nil {
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
		if organizationAuthorId != organizationEmployeeId {
			fmt.Println(employeeId)
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
	}
	var inputName, inputDescription string
	if input.Name != nil {
		inputName = *input.Name
	}
	if input.Description != nil {
		inputDescription = *input.Description
	}
	output, err := r.bidService.EditBid(c.Request().Context(), service.EditBidInput{
		Id:          input.BidId,
		Name:        inputName,
		Description: inputDescription,
	})
	if err != nil {
		if errors.Is(err, service.ErrTenderNotFound) {
			return errors2.NewErrorResponse(c, http.StatusNotFound, err)
		}

		if errors.Is(err, service.ErrPermissionDenied) {
			return errors2.NewErrorResponse(c, http.StatusForbidden, err)
		}
		return errors2.NewErrorResponse(c, http.StatusInternalServerError, err)
	}
	type response struct {
		Id         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		Status     string    `json:"status"`
		AuthorType string    `json:"authorType"`
		AuthorId   uuid.UUID `json:"authorId"`
		Version    int       `json:"version"`
		CreatedAt  string    `json:"createdAt"`
	}

	return c.JSON(http.StatusOK, response{
		Id:         output.Id,
		Name:       output.Name,
		Status:     output.Status,
		AuthorType: output.AuthorType,
		AuthorId:   output.AuthorId,
		Version:    output.Version,
		CreatedAt:  output.CreatedAt.Format(formating.TimeFormat),
	})
}

type RollbackBidInput struct {
	BidId    uuid.UUID `param:"bid_id" validate:"required"`
	Version  int       `param:"version" validate:"required,gte=1"`
	Username string    `query:"username" validate:"required"`
}

func (r *bidRoutes) rollback(c echo.Context) error {
	var input RollbackBidInput
	if err := c.Bind(&input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	b := echo.DefaultBinder{}
	if err := b.BindQueryParams(c, &input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	if err := c.Validate(input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	bid, err := r.bidService.GetBidById(c.Request().Context(), input.BidId)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusNotFound, err)
	}

	employeeId, err := r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.Username)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}

	switch bid.AuthorType {
	case "User":
		if bid.AuthorId != employeeId {
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
	case "Organization":
		organizationEmployeeId, err := r.employeeService.GetEmployeeOrgIdById(c.Request().Context(), employeeId)
		if err != nil {
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
		organizationAuthorId, err := r.employeeService.GetEmployeeOrgIdById(c.Request().Context(), bid.AuthorId)
		if err != nil {
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
		if organizationAuthorId != organizationEmployeeId {
			fmt.Println(employeeId)
			return errors2.NewErrorResponse(c, http.StatusForbidden, service.ErrPermissionDenied)
		}
	}
	output, err := r.bidService.RollbackVersion(c.Request().Context(), service.RollbackVersionInput{
		Id:      input.BidId,
		Version: input.Version,
	})
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return errors2.NewErrorResponse(c, http.StatusForbidden, err)
		}
		if errors.Is(err, service.ErrVersionNotFound) {
			return errors2.NewErrorResponse(c, http.StatusNotFound, err)
		}
		if errors.Is(err, service.ErrBidNotFound) {
			return errors2.NewErrorResponse(c, http.StatusNotFound, err)
		}
		return errors2.NewErrorResponse(c, http.StatusInternalServerError, err)
	}
	type response struct {
		Id         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		Status     string    `json:"status"`
		AuthorType string    `json:"authorType"`
		AuthorId   uuid.UUID `json:"authorId"`
		Version    int       `json:"version"`
		CreatedAt  string    `json:"createdAt"`
	}

	return c.JSON(http.StatusOK, response{
		Id:         output.Id,
		Name:       output.Name,
		Status:     output.Status,
		AuthorType: output.AuthorType,
		AuthorId:   output.AuthorId,
		Version:    output.Version,
		CreatedAt:  output.CreatedAt.Format(formating.TimeFormat),
	})
}

type SubmitDecisionInput struct {
	BidId    uuid.UUID `param:"bid_id" validate:"required"`
	Decision string    `query:"decision" validate:"required,oneof=Approved Rejected"`
	Username string    `query:"username" validate:"required"`
}

func (r *bidRoutes) submitDecision(c echo.Context) error {
	var input SubmitDecisionInput
	if err := c.Bind(&input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	b := echo.DefaultBinder{}
	if err := b.BindQueryParams(c, &input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	if err := c.Validate(input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	bid, err := r.bidService.GetBidById(c.Request().Context(), input.BidId)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusNotFound, err)
	}

	employeeId, err := r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.Username)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}
	tender, err := r.tenderService.GetTenderById(c.Request().Context(), bid.TenderId)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusNotFound, err)
	}

	employeeOrg, err := r.employeeService.GetEmployeeOrgIdById(c.Request().Context(), employeeId)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusForbidden, err)
	}
	if employeeOrg != tender.OrganizationId {
		return errors2.NewErrorResponse(c, http.StatusForbidden, err)
	}
	output, err := r.bidService.SubmitDecision(c.Request().Context(), bid.TenderId, bid.Id, input.Decision)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusInternalServerError, err)
	}

	type response struct {
		Id         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		Status     string    `json:"status"`
		AuthorType string    `json:"authorType"`
		AuthorId   uuid.UUID `json:"authorId"`
		Version    int       `json:"version"`
		CreatedAt  string    `json:"createdAt"`
	}

	return c.JSON(http.StatusOK, response{
		Id:         output.Id,
		Name:       output.Name,
		Status:     output.Status,
		AuthorType: output.AuthorType,
		AuthorId:   output.AuthorId,
		Version:    output.Version,
		CreatedAt:  output.CreatedAt.Format(formating.TimeFormat),
	})
}
