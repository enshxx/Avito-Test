package service

import (
	"avito/internal/entity"
	"avito/internal/repo"
	"context"
	"github.com/google/uuid"
)

type EmployeeService struct {
	employeeRepo repo.Employee
}

func NewEmployeeService(employeeRepo repo.Employee) *EmployeeService {
	return &EmployeeService{employeeRepo: employeeRepo}
}

func (s *EmployeeService) GetEmployeeIdByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	id, err := s.employeeRepo.GetEmployeeIdByUsername(ctx, username)
	if err != nil {
		return uuid.New(), ErrEmployeeDoesNotExist
	}
	return id, nil
}

func (s *EmployeeService) GetEmployeeById(ctx context.Context, id uuid.UUID) (*entity.Employee, error) {
	e, err := s.employeeRepo.GetEmployeeById(ctx, id)
	if err != nil {
		return nil, ErrEmployeeDoesNotExist
	}
	return e, nil
}

func (s *EmployeeService) GetEmployeeOrgIdById(ctx context.Context, employeeId uuid.UUID) (uuid.UUID, error) {
	id, err := s.employeeRepo.GetEmployeeOrgIdById(ctx, employeeId)
	if err != nil {
		return uuid.New(), ErrOrganisationResponsibleNotFound
	}
	return id, nil
}
