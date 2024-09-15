package pgdb

import (
	"avito/internal/entity"
	"avito/internal/repo/repoerrs"
	"avito/pkg/postgres"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type EmployeeRepo struct {
	*postgres.Postgres
}

func NewEmployeeRepo(pg *postgres.Postgres) *EmployeeRepo {
	return &EmployeeRepo{pg}
}

func (r *EmployeeRepo) GetEmployeeIdByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.Pool.QueryRow(ctx, "SELECT id FROM employee WHERE username=$1", username).Scan(&id)
	if err != nil {
		log.Debugf("err: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return id, repoerrs.ErrNotFound
		}
		return id, fmt.Errorf("EmployeeRepo.GetEmployeeIdByUsername - r.Pool.QueryRow: %v", err)
	}
	return id, nil
}

func (r *EmployeeRepo) GetEmployeeById(ctx context.Context, id uuid.UUID) (*entity.Employee, error) {
	request := `SELECT * 
				FROM employee
				WHERE id = $1`
	rows, err := r.Pool.Query(ctx, request, id)
	defer rows.Close()
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, fmt.Errorf("EmployeeRepo.GetEmployeeById  - r.Pool.Query: %v", err)
	}
	e, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[entity.Employee])
	if err != nil {
		log.Debugf("err: %v", err)
		return nil, repoerrs.ErrNotFound
	}
	return &e, nil
}

func (r *EmployeeRepo) GetEmployeeOrgIdById(ctx context.Context, employeeId uuid.UUID) (uuid.UUID, error) {
	request := `SELECT organization_id
				FROM organization_responsible
				WHERE user_id = $1`
	var id uuid.UUID
	err := r.Pool.QueryRow(ctx, request, employeeId).Scan(&id)
	if err != nil {
		log.Debugf("err: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return id, repoerrs.ErrNotFound
		}
		return id, fmt.Errorf("EmployeeRepo.GetEmployeeOrgRespById - r.Pool.QueryRow: %v", err)
	}
	return id, nil
}
