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

type BidService struct {
	bidRepo    repo.Bid
	tenderRepo repo.Tender
}

func NewBidService(bidRepo repo.Bid, tenderRepo repo.Tender) *BidService {
	return &BidService{
		bidRepo:    bidRepo,
		tenderRepo: tenderRepo,
	}
}

func (s *BidService) CreateBid(ctx context.Context, input BidCreateInput) (*entity.Bid, error) {
	bid, err := s.bidRepo.CreateBid(
		ctx,
		input.Name,
		input.Description,
		input.TenderId,
		input.AuthorType,
		input.AuthorId,
	)
	if err != nil {
		return nil, ErrCannotCreateBid
	}
	return bid, nil
}

func (s *BidService) GetMyBids(ctx context.Context, input GetMyBidsInput) ([]GetMyBidsOutput, error) {
	bids, err := s.bidRepo.GetMyBids(
		ctx,
		input.AuthorId,
		input.Limit,
		input.Offset,
	)
	if err != nil {
		return nil, ErrCannotGetBids
	}
	output := make([]GetMyBidsOutput, len(bids))
	for i, bid := range bids {
		output[i] = GetMyBidsOutput{
			Id:         bid.Id,
			Name:       bid.Name,
			Status:     bid.Status,
			AuthorType: bid.AuthorType,
			AuthorId:   bid.AuthorId,
			Version:    bid.Version,
			CreatedAt:  bid.CreatedAt.Format(formating.TimeFormat),
		}
	}
	return output, nil
}

func (s *BidService) GetBidById(ctx context.Context, id uuid.UUID) (*entity.Bid, error) {
	bid, err := s.bidRepo.GetBidById(ctx, id)
	if err != nil {
		return nil, ErrBidNotFound
	}
	return bid, nil
}

func (s *BidService) GetStatus(ctx context.Context, input GetBidStatusInput) (string, error) {
	bid, err := s.bidRepo.GetBidById(ctx, input.BidId)
	if err != nil {
		return "", ErrBidNotFound
	}
	return bid.Status, nil
}

func (s *BidService) PutStatus(ctx context.Context, input PutBidStatusInput) (*PutBidStatusOutput, error) {
	bid, err := s.bidRepo.PutStatus(ctx, input.BidId, input.Status)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, ErrBidNotFound
		}
		return nil, ErrCannotPutStatus
	}
	return &PutBidStatusOutput{
		Id:         bid.Id,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorId:   bid.AuthorId,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt,
	}, nil
}

func (s *BidService) EditBid(ctx context.Context, input EditBidInput) (*EditBidOutput, error) {
	bid, err := s.bidRepo.EditBid(ctx, input.Id, input.Name, input.Description)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, ErrBidNotFound
		}
		return nil, ErrCannotEditBid
	}
	return &EditBidOutput{
		Id:         bid.Id,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorId:   bid.AuthorId,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt,
	}, nil
}

func (s *BidService) RollbackVersion(ctx context.Context, input RollbackVersionInput) (*RollbackBidVersionOutput, error) {
	bid, err := s.bidRepo.RollbackVersion(ctx, input.Id, input.Version)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, ErrBidNotFound
		}
		if errors.Is(err, repoerrs.ErrVersionNotFound) {
			return nil, ErrVersionNotFound
		}
	}
	return &RollbackBidVersionOutput{
		Id:         bid.Id,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorId:   bid.AuthorId,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt,
	}, nil
}

func (s *BidService) SubmitDecision(ctx context.Context, tenderId, bidId uuid.UUID, decision string) (*entity.Bid, error) {
	if decision == "Approved" {
		bid, err := s.bidRepo.GetBidById(ctx, bidId)
		if err != nil {
			return nil, ErrBidNotFound
		}
		_, err = s.tenderRepo.PutStatus(ctx, tenderId, "Closed")
		if err != nil {
			return nil, ErrTenderNotFound
		}
		return bid, nil
	} else {
		bid, err := s.bidRepo.PutStatus(ctx, bidId, "Canceled")
		if err != nil {
			return nil, ErrBidNotFound
		}
		return bid, nil
	}
}
