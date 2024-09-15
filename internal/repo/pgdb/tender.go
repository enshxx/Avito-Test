package pgdb

import (
	"avito/internal/entity"
	"avito/internal/repo/repoerrs"
	"avito/pkg/postgres"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type TenderRepo struct {
	*postgres.Postgres
}

func NewTenderRepo(pg *postgres.Postgres) *TenderRepo {
	return &TenderRepo{pg}
}

func (r *TenderRepo) CreateTender(ctx context.Context, name, description, serviceType string, organisationId uuid.UUID, creatorUsername string) (*entity.Tender, error) {
	rows, err := r.Pool.Query(ctx, "INSERT INTO tender (name, description, type, organization_id, creator_username) VALUES ($1, $2, $3, $4, $5) RETURNING *", name, description, serviceType, organisationId, creatorUsername)
	defer rows.Close()
	t, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Tender])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("TenderRepo.CreateTender - r.Pool.Query: %v", err)
	}
	return &t, nil
}

func (r *TenderRepo) GetMyTenders(ctx context.Context, username string, limit, offset int) ([]entity.Tender, error) {
	request := `SELECT *
				FROM tender
				WHERE creator_username=$1 AND version = (SELECT MAX(version)
                	FROM tender AS t
                	WHERE t.id = tender.id)
				ORDER BY name
				LIMIT $2
				OFFSET $3`
	rows, err := r.Pool.Query(ctx, request, username, limit, offset)
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("TenderRepo.GetMyTenders - r.Pool.Query: %v", err)
	}
	defer rows.Close()
	tenders := make([]entity.Tender, 0)
	for rows.Next() {
		var tender entity.Tender
		err = rows.Scan(&tender.Id, &tender.Name, &tender.Description, &tender.Type, &tender.Status, &tender.OrganizationId, &tender.Version, &tender.CreatorUsername, &tender.CreatedAt, &tender.UpdatedAt)
		if err != nil {
			log.Debugf("err: %v", err)
			return nil, fmt.Errorf("TenderRepo.GetMyTenders - rows.Scan: %v", err)
		}
		tenders = append(tenders, tender)
	}
	return tenders, nil
}

func (r *TenderRepo) GetTenders(ctx context.Context, serviceTypes []string, limit, offset int) ([]entity.Tender, error) {
	request := `SELECT *
				FROM tender
				WHERE type = ANY($1) AND status='Published'
				AND version = (SELECT MAX(version)
                	FROM tender AS t
                	WHERE t.id = tender.id)
				ORDER BY name
				LIMIT $2
				OFFSET $3;`

	rows, err := r.Pool.Query(ctx, request, pq.Array(serviceTypes), limit, offset)
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("TenderRepo.GetTenders - r.Pool.Query: %v", err)
	}
	defer rows.Close()
	tenders := make([]entity.Tender, 0)
	for rows.Next() {
		var tender entity.Tender
		err = rows.Scan(&tender.Id, &tender.Name, &tender.Description, &tender.Type, &tender.Status, &tender.OrganizationId, &tender.Version, &tender.CreatorUsername, &tender.CreatedAt, &tender.UpdatedAt)
		if err != nil {
			log.Debugf("err: %v", err)
			return nil, fmt.Errorf("TenderRepo.GetTenders - rows.Scan: %v", err)
		}
		tenders = append(tenders, tender)
	}
	return tenders, nil
}

func (r *TenderRepo) GetTenderById(ctx context.Context, tenderId uuid.UUID) (*entity.Tender, error) {
	request := `SELECT * 
				FROM tender 
				WHERE id = $1 AND version = (SELECT MAX(version)
                	FROM tender AS t
                	WHERE t.id = tender.id)`

	rows, err := r.Pool.Query(ctx, request, tenderId)
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("TenderRepo.GetStatus - r.Pool.Query: %v", err)
	}
	defer rows.Close()
	t, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Tender])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, repoerrs.ErrNotFound
	}
	return &t, nil
}

func (r *TenderRepo) PutStatus(ctx context.Context, tenderId uuid.UUID, status string) (*entity.Tender, error) {
	request := `UPDATE tender 
				SET status=$1 
				WHERE id=$2 AND version = (SELECT MAX(version)
                	FROM tender AS t
                	WHERE t.id = tender.id)
                RETURNING *`
	rows, err := r.Pool.Query(ctx, request, status, tenderId)
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("TenderRepo.PutStatus - r.Pool.Query: %v", err)
	}
	defer rows.Close()
	t, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Tender])

	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("TenderRepo.PutTenderById - r.Pool.Query: %v", err)
	}
	return &t, nil
}

func (r *TenderRepo) EditTender(ctx context.Context, tenderId uuid.UUID, name, description, serviceType string) (*entity.Tender, error) {
	prevVReq := `SELECT *
				 FROM tender
			     WHERE id=$1 AND version = (SELECT MAX(version)
				 FROM tender AS t
				 WHERE t.id = tender.id)
				 `
	rows, err := r.Pool.Query(ctx, prevVReq, tenderId)
	defer rows.Close()
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("TenderRepo.EditTender - GetPrevVersion - r.Pool.Query: %v", err)
	}
	t, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Tender])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, repoerrs.ErrNotFound
	}
	request := `INSERT INTO tender (id, name, description, type, organization_id, creator_username, status, version)
				VALUES 
				    ($1, $2, $3, $4, $5, $6, $7, (
				    	(SELECT COALESCE(MAX(version), 0) + 1
				    	 FROM tender AS t
				     	WHERE t.id = $1)))
				RETURNING *`
	if name == "" {
		name = t.Name
	}
	if description == "" {
		description = t.Description
	}
	if serviceType == "" {
		serviceType = t.Type
	}
	result, err := r.Pool.Query(ctx, request, t.Id, name, description, serviceType, t.OrganizationId, t.CreatorUsername, t.Status)
	defer result.Close()
	t, err = pgx.CollectOneRow(result, pgx.RowToStructByName[entity.Tender])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("TenderRepo.EditTender - r.Pool.Query: %v", err)
	}
	return &t, nil
}

func (r *TenderRepo) RollbackVersion(ctx context.Context, tenderId uuid.UUID, version int) (*entity.Tender, error) {
	prevVReq := `SELECT *
				 FROM tender
			     WHERE id=$1 AND version = $2
				 `
	rows, err := r.Pool.Query(ctx, prevVReq, tenderId, version)
	defer rows.Close()
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("TenderRepo.EditTender - GetPrevVersion - r.Pool.Query: %v", err)
	}
	t, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Tender])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, repoerrs.ErrNotFound
	}
	lastVReq := `SELECT MAX(version)
				FROM tender
				WHERE id=$1`
	var lastV int
	if err := r.Pool.QueryRow(ctx, lastVReq, tenderId).Scan(&lastV); err != nil {
		log.Debugf("err: %v", err)
		return nil, repoerrs.ErrVersionNotFound
	}
	request := `INSERT INTO tender (id, name, description, type, organization_id, creator_username, status, version)
				VALUES 
				    ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING *`

	result, err := r.Pool.Query(ctx, request, t.Id, t.Name, t.Description, t.Type, t.OrganizationId, t.CreatorUsername, t.Status, lastV+1)
	defer result.Close()
	t, err = pgx.CollectOneRow(result, pgx.RowToStructByName[entity.Tender])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("TenderRepo.RollbackVersion - r.Pool.Query: %v", err)
	}
	return &t, nil
}
