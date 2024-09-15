package service

import (
	"avito/internal/controllers/http/formating"
	"avito/internal/entity"
	"avito/internal/repo"
	"avito/internal/repo/repoerrs"
	"context"
	"errors"
	"github.com/google/uuid"
)

type TenderService struct {
	tenderRepo repo.Tender
}

func NewTenderService(tenderRepo repo.Tender) *TenderService {
	return &TenderService{tenderRepo: tenderRepo}
}

func (s *TenderService) CreateTender(ctx context.Context, input TenderCreateInput) (*entity.Tender, error) {
	tender, err := s.tenderRepo.CreateTender(
		ctx,
		input.Name,
		input.Description,
		input.ServiceType,
		input.OrganizationId,
		input.CreatorUsername,
	)
	if err != nil {
		return nil, ErrCannotCreateTender
	}
	return tender, nil
}

func (s *TenderService) GetMyTenders(ctx context.Context, input GetMyTendersInput) ([]GetMyTendersOutput, error) {
	tenders, err := s.tenderRepo.GetMyTenders(
		ctx,
		input.Username,
		input.Limit,
		input.Offset,
	)
	if err != nil {
		return nil, ErrCannotGetTender
	}
	output := make([]GetMyTendersOutput, len(tenders))
	for i, tender := range tenders {
		output[i] = GetMyTendersOutput{
			Id:             tender.Id,
			Name:           tender.Name,
			Description:    tender.Description,
			Status:         tender.Status,
			ServiceType:    tender.Type,
			OrganizationId: tender.OrganizationId,
			Version:        tender.Version,
			CreatedAt:      tender.CreatedAt.Format(formating.TimeFormat),
		}
	}
	return output, nil
}

func (s *TenderService) GetTenders(ctx context.Context, input GetTendersInput) ([]GetMyTendersOutput, error) {
	tenders, err := s.tenderRepo.GetTenders(
		ctx,
		input.ServiceTypes,
		input.Limit,
		input.Offset,
	)
	if err != nil {
		return nil, ErrCannotGetTender
	}
	output := make([]GetMyTendersOutput, len(tenders))
	for i, tender := range tenders {
		output[i] = GetMyTendersOutput{
			Id:             tender.Id,
			Name:           tender.Name,
			Description:    tender.Description,
			Status:         tender.Status,
			ServiceType:    tender.Type,
			OrganizationId: tender.OrganizationId,
			Version:        tender.Version,
			CreatedAt:      tender.CreatedAt.Format(formating.TimeFormat),
		}
	}
	return output, nil
}

func (s *TenderService) GetStatus(ctx context.Context, input GetStatusInput) (string, error) {
	tender, err := s.tenderRepo.GetTenderById(ctx, input.TenderId)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return "", ErrTenderNotFound
		}
		return "", ErrCannotGetStatus
	}
	if tender.CreatorUsername != input.Username {
		return "", ErrPermissionDenied
	}
	return tender.Status, nil
}

func (s *TenderService) PutStatus(ctx context.Context, input PutStatusInput) (*PutStatusOutput, error) {
	tender, err := s.tenderRepo.GetTenderById(ctx, input.TenderId)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, ErrTenderNotFound
		}
		return nil, ErrCannotPutStatus
	}
	if tender.CreatorUsername != input.Username {
		return nil, ErrPermissionDenied
	}
	tender, err = s.tenderRepo.PutStatus(ctx, input.TenderId, input.Status)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, ErrTenderNotFound
		}
		return nil, ErrCannotPutStatus
	}

	return &PutStatusOutput{
		Id:          tender.Id,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		ServiceType: tender.Type,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt,
	}, nil
}

func (s *TenderService) EditTender(ctx context.Context, input EditTenderInput) (*EditTenderOutput, error) {
	tender, err := s.tenderRepo.EditTender(ctx, input.Id, input.Name, input.Description, input.ServiceType)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, ErrTenderNotFound
		}
		return nil, ErrCannotEditTender
	}
	return &EditTenderOutput{
		Id:          tender.Id,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		ServiceType: tender.Type,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt,
	}, nil
}

func (s *TenderService) RollbackVersion(ctx context.Context, input RollbackVersionInput) (*RollbackVersionOutput, error) {
	tender, err := s.tenderRepo.RollbackVersion(ctx, input.Id, input.Version)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, ErrTenderNotFound
		}
		if errors.Is(err, repoerrs.ErrVersionNotFound) {
			return nil, ErrVersionNotFound
		}
	}
	return &RollbackVersionOutput{
		Id:          tender.Id,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		ServiceType: tender.Type,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt,
	}, nil
}

func (s *TenderService) GetTenderById(ctx context.Context, id uuid.UUID) (*entity.Tender, error) {
	tender, err := s.tenderRepo.GetTenderById(ctx, id)
	if err != nil {
		return nil, ErrTenderNotFound
	}
	return tender, nil
}
