package pgdb

import (
	"avito/internal/entity"
	"avito/internal/repo/repoerrs"
	"avito/pkg/postgres"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
)

type BidRepo struct {
	*postgres.Postgres
}

func NewBidRepo(pg *postgres.Postgres) *BidRepo {
	return &BidRepo{pg}
}

func (r *BidRepo) CreateBid(ctx context.Context, name, description string, tenderId uuid.UUID, authorType string, authorId uuid.UUID) (*entity.Bid, error) {
	request := `INSERT INTO bid (name, description, tender_id, author_type, author_id)
				VALUES 
				    ($1, $2, $3, $4, $5)
				RETURNING *`
	rows, err := r.Pool.Query(ctx, request, name, description, tenderId, authorType, authorId)
	defer rows.Close()
	b, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Bid])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("BidRepo.CreateBid - r.Pool.Query: %v", err)
	}
	return &b, nil
}

func (r *BidRepo) GetMyBids(ctx context.Context, authorId uuid.UUID, limit, offset int) ([]entity.Bid, error) {
	request := `SELECT *
				FROM bid
				WHERE author_id=$1 AND version = (SELECT MAX(version)
                	FROM bid AS b
                	WHERE b.id = bid.id)
				ORDER BY name
				LIMIT $2
				OFFSET $3`
	rows, err := r.Pool.Query(ctx, request, authorId, limit, offset)
	defer rows.Close()
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("BidRepo.GetMyBids - r.Pool.Query: %v", err)
	}
	bids, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Bid])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("BidRepo.GetMyBids - pgx.CollectRows: %v", err)
	}
	return bids, nil
}

func (r *BidRepo) GetBidById(ctx context.Context, bidId uuid.UUID) (*entity.Bid, error) {
	request := `SELECT *
				FROM bid
				WHERE id=$1 AND version = (SELECT MAX(version)
                	FROM bid AS b
                	WHERE b.id = bid.id)`
	rows, err := r.Pool.Query(ctx, request, bidId)
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("BidRepo.GetBidById - r.Pool.Query: %v", err)
	}
	b, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Bid])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("BidRepo.GetBidById - r.Pool.Query: %v", err)
	}
	return &b, nil
}

func (r *BidRepo) PutStatus(ctx context.Context, BidId uuid.UUID, status string) (*entity.Bid, error) {
	request := `UPDATE bid 
				SET status=$1 
				WHERE id=$2 AND version = (SELECT MAX(version)
                	FROM bid AS b
                	WHERE b.id = bid.id)
                RETURNING *`
	rows, err := r.Pool.Query(ctx, request, status, BidId)
	defer rows.Close()
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("BidRepo.PutStatus - r.Pool.Query: %v", err)
	}

	b, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Bid])

	if err != nil {
		log.Debugf("err: %v", err)
		return nil, repoerrs.ErrNotFound
	}
	return &b, nil
}

func (r *BidRepo) EditBid(ctx context.Context, bidId uuid.UUID, name, description string) (*entity.Bid, error) {
	prevVReq := `SELECT *
				 FROM bid
			     WHERE id=$1 AND version = (SELECT MAX(version)
				 FROM bid AS b
				 WHERE b.id = bid.id)
				 `
	rows, err := r.Pool.Query(ctx, prevVReq, bidId)

	defer rows.Close()
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("BidRepo.EditBid - GetPrevVersion - r.Pool.Query: %v", err)
	}
	b, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Bid])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, repoerrs.ErrNotFound
	}

	request := `INSERT INTO bid (id, name, description, tender_id, status, author_type, author_id, version)
				VALUES 
				    ($1, $2, $3, $4, $5, $6, $7, (
				    	(SELECT COALESCE(MAX(version), 0) + 1
				    	 FROM bid AS b
				     	 WHERE b.id = $1)))
				RETURNING *`
	if name == "" {
		name = b.Name
	}
	if description == "" {
		description = b.Description
	}

	result, err := r.Pool.Query(ctx, request, b.Id, name, description, b.TenderId, b.Status, b.AuthorType, b.AuthorId)
	defer result.Close()
	b, err = pgx.CollectOneRow(result, pgx.RowToStructByName[entity.Bid])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("BidRepo.EditBid - r.Pool.Query: %v", err)
	}
	return &b, nil
}

func (r *BidRepo) RollbackVersion(ctx context.Context, bidId uuid.UUID, version int) (*entity.Bid, error) {
	prevVReq := `SELECT *
				 FROM bid
			     WHERE id=$1 AND version = $2
				 `
	rows, err := r.Pool.Query(ctx, prevVReq, bidId, version)
	defer rows.Close()
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("BidRepo.EditTender - GetPrevVersion - r.Pool.Query: %v", err)
	}
	b, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Bid])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, repoerrs.ErrVersionNotFound
	}
	lastVReq := `SELECT MAX(version)
				FROM bid
				WHERE id=$1`
	var lastV int
	if err := r.Pool.QueryRow(ctx, lastVReq, bidId).Scan(&lastV); err != nil {
		log.Debugf("err: %v", err)
		return nil, repoerrs.ErrVersionNotFound
	}
	request := `INSERT INTO bid (id, name, description, tender_id, status, author_type, author_id, version)
				VALUES 
				    ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING *`

	result, err := r.Pool.Query(ctx, request, b.Id, b.Name, b.Description, b.TenderId, b.Status, b.AuthorType, b.AuthorId, lastV+1)
	defer result.Close()
	b, err = pgx.CollectOneRow(result, pgx.RowToStructByName[entity.Bid])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("BidRepo.RollbackVersion - r.Pool.Query: %v", err)
	}
	return &b, nil
}
