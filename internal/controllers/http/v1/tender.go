package v1

import (
	errors2 "avito/internal/controllers/http/errors"
	"avito/internal/controllers/http/formating"
	tenders "avito/internal/controllers/http/parser"
	"avito/internal/controllers/validators"
	"avito/internal/service"
	"errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

type tenderRoutes struct {
	tenderService   service.Tender
	employeeService service.Employee
}

func newTenderRoutes(g *echo.Group, tenderService service.Tender, employeeService service.Employee) {
	r := &tenderRoutes{
		tenderService:   tenderService,
		employeeService: employeeService,
	}
	g.POST("/new", r.create)
	g.GET("/my", r.getMyTenders)
	g.GET("", r.getTenders)
	g.GET("/:tender_id/status", r.getStatus)
	g.PUT("/:tender_id/status", r.putStatus)
	g.PATCH("/:tender_id/edit", r.editTender)
	g.PUT("/:tender_id/rollback/:version", r.rollback)
}

type TenderCreationInput struct {
	Name            string    `json:"name" validate:"required,max=100"`
	Description     string    `json:"description" validate:"required,max=500"`
	ServiceType     string    `json:"serviceType" validate:"required,oneof=Construction Delivery Manufacture"`
	OrganizationId  uuid.UUID `json:"organizationId" validate:"required"`
	CreatorUsername string    `json:"creatorUsername" validate:"required,max=50"`
}

func (r *tenderRoutes) create(c echo.Context) error {
	input := TenderCreationInput{}

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
	employeeId, err := r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.CreatorUsername)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}
	_, err = r.employeeService.GetEmployeeOrgIdById(c.Request().Context(), employeeId)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusForbidden, err)
	}
	tender, err := r.tenderService.CreateTender(c.Request().Context(), service.TenderCreateInput{
		Name:            input.Name,
		Description:     input.Description,
		ServiceType:     input.ServiceType,
		OrganizationId:  input.OrganizationId,
		CreatorUsername: input.CreatorUsername,
	})
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusInternalServerError, err)
	}

	type response struct {
		Id             uuid.UUID `json:"id"`
		Name           string    `json:"name"`
		Description    string    `json:"description"`
		Status         string    `json:"status"`
		ServiceType    string    `json:"serviceType"`
		Version        int       `json:"version"`
		OrganizationId uuid.UUID `json:"organization_id"`
		CreatedAt      string    `json:"createdAt"`
	}

	return c.JSON(http.StatusOK, response{
		Id:             tender.Id,
		Name:           tender.Name,
		Description:    tender.Description,
		Status:         tender.Status,
		ServiceType:    tender.Type,
		OrganizationId: tender.OrganizationId,
		Version:        tender.Version,
		CreatedAt:      tender.CreatedAt.Format(formating.TimeFormat),
	})
}

type GetMyTendersInput struct {
	Username string `query:"username" validate:"required"`
	Limit    int    `query:"limit"`
	Offset   int    `query:"offset"`
}

func (r *tenderRoutes) getMyTenders(c echo.Context) error {
	var input GetMyTendersInput
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
	_, err = r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.Username)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}
	response, err := r.tenderService.GetMyTenders(c.Request().Context(), service.GetMyTendersInput{
		Username: input.Username,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, response)
}

type GetTendersInput struct {
	ServiceTypes []string `query:"service_type"`
	Limit        int      `query:"limit"`
	Offset       int      `query:"offset"`
}

func (r *tenderRoutes) getTenders(c echo.Context) error {
	var input GetTendersInput
	if err := c.Bind(&input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := c.Validate(input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	rawQuery := c.Request().URL.RawQuery
	limit, offset, serviceTypes, err := tenders.ParseLimitOffsetService(rawQuery)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}
	response, err := r.tenderService.GetTenders(c.Request().Context(), service.GetTendersInput{
		ServiceTypes: serviceTypes,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, response)
}

type getStatusInput struct {
	TenderId uuid.UUID `param:"tender_id"`
	Username string    `query:"username" validate:"required"`
}

func (r *tenderRoutes) getStatus(c echo.Context) error {
	var input getStatusInput
	if err := c.Bind(&input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := c.Validate(input); err != nil {
		return errors2.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	_, err := r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.Username)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}

	response, err := r.tenderService.GetStatus(c.Request().Context(), service.GetStatusInput{
		TenderId: input.TenderId,
		Username: input.Username,
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
	return c.JSON(http.StatusOK, response)
}

type PutStatusInput struct {
	TenderId uuid.UUID `param:"tender_id" validate:"required"`
	Username string    `query:"username" validate:"required"`
	Status   string    `query:"status" validate:"required,oneof=Created Published Closed"`
}

func (r *tenderRoutes) putStatus(c echo.Context) error {
	var input PutStatusInput
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
	_, err := r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.Username)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}
	tender, err := r.tenderService.PutStatus(c.Request().Context(), service.PutStatusInput{
		TenderId: input.TenderId,
		Username: input.Username,
		Status:   input.Status,
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
		Id          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Status      string    `json:"status"`
		ServiceType string    `json:"serviceType"`
		Version     int       `json:"version"`
		CreatedAt   string    `json:"createdAt"`
	}
	return c.JSON(http.StatusOK, response{
		Id:          tender.Id,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		ServiceType: tender.ServiceType,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt.Format(formating.TimeFormat),
	})
}

type EditTenderInput struct {
	TenderId    uuid.UUID `param:"tender_id"`
	Username    string    `query:"username" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty"`
	Description *string   `json:"description" validate:"omitempty"`
	ServiceType *string   `json:"service_type" validate:"omitempty,oneof=Construction Delivery Manufacture"`
}

func (r *tenderRoutes) editTender(c echo.Context) error {
	var input EditTenderInput
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
	_, err := r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.Username)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}
	var inputName, inputDescription, inputServiceType string
	if input.Name != nil {
		inputName = *input.Name
	}
	if input.Description != nil {
		inputDescription = *input.Description
	}
	if input.ServiceType != nil {
		inputServiceType = *input.ServiceType
	}
	tender, err := r.tenderService.EditTender(c.Request().Context(), service.EditTenderInput{
		Id:          input.TenderId,
		Name:        inputName,
		Description: inputDescription,
		ServiceType: inputServiceType,
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
		Id          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Status      string    `json:"status"`
		ServiceType string    `json:"serviceType"`
		Version     int       `json:"version"`
		CreatedAt   string    `json:"createdAt"`
	}
	return c.JSON(http.StatusOK, response{
		Id:          tender.Id,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		ServiceType: tender.ServiceType,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt.Format(formating.TimeFormat),
	})
}

type RollbackInput struct {
	TenderId uuid.UUID `param:"tender_id" validate:"required"`
	Version  int       `param:"version" validate:"required,gte=1"`
	Username string    `query:"username" validate:"required"`
}

func (r *tenderRoutes) rollback(c echo.Context) error {
	var input RollbackInput
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
	_, err := r.employeeService.GetEmployeeIdByUsername(c.Request().Context(), input.Username)
	if err != nil {
		return errors2.NewErrorResponse(c, http.StatusUnauthorized, err)
	}

	tender, err := r.tenderService.RollbackVersion(c.Request().Context(), service.RollbackVersionInput{
		Id:      input.TenderId,
		Version: input.Version,
	})
	if err != nil {
		if errors.Is(err, service.ErrTenderNotFound) {
			return errors2.NewErrorResponse(c, http.StatusNotFound, err)
		}
		if errors.Is(err, service.ErrPermissionDenied) {
			return errors2.NewErrorResponse(c, http.StatusForbidden, err)
		}
		if errors.Is(err, service.ErrVersionNotFound) {
			return errors2.NewErrorResponse(c, http.StatusNotFound, err)
		}
		return errors2.NewErrorResponse(c, http.StatusInternalServerError, err)
	}
	type response struct {
		Id          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Status      string    `json:"status"`
		ServiceType string    `json:"serviceType"`
		Version     int       `json:"version"`
		CreatedAt   string    `json:"createdAt"`
	}
	return c.JSON(http.StatusOK, response{
		Id:          tender.Id,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		ServiceType: tender.ServiceType,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt.Format(formating.TimeFormat),
	})
}
