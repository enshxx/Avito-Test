package repo

import (
	"avito/internal/entity"
	"avito/internal/repo/pgdb"
	"avito/pkg/postgres"
	"context"
	"github.com/google/uuid"
)

type Tender interface {
	CreateTender(ctx context.Context, name, description, serviceType string, organisationId uuid.UUID, creatorUsername string) (*entity.Tender, error)
	GetMyTenders(ctx context.Context, username string, limit, offset int) ([]entity.Tender, error)
	GetTenders(ctx context.Context, serviceTypes []string, limit, offset int) ([]entity.Tender, error)
	GetTenderById(ctx context.Context, tenderId uuid.UUID) (*entity.Tender, error)
	PutStatus(ctx context.Context, tenderId uuid.UUID, status string) (*entity.Tender, error)
	EditTender(ctx context.Context, tenderId uuid.UUID, name, description, serviceType string) (*entity.Tender, error)
	RollbackVersion(ctx context.Context, tenderId uuid.UUID, version int) (*entity.Tender, error)
}

type Employee interface {
	GetEmployeeIdByUsername(ctx context.Context, username string) (uuid.UUID, error)
	GetEmployeeById(ctx context.Context, id uuid.UUID) (*entity.Employee, error)
	GetEmployeeOrgIdById(ctx context.Context, employeeId uuid.UUID) (uuid.UUID, error)
}
type Bid interface {
	CreateBid(ctx context.Context, name, description string, tenderId uuid.UUID, authorType string, authorId uuid.UUID) (*entity.Bid, error)
	GetMyBids(ctx context.Context, authorId uuid.UUID, limit, offset int) ([]entity.Bid, error)
	GetBidById(ctx context.Context, bidId uuid.UUID) (*entity.Bid, error)
	PutStatus(ctx context.Context, BidId uuid.UUID, status string) (*entity.Bid, error)
	EditBid(ctx context.Context, bidId uuid.UUID, name, description string) (*entity.Bid, error)
	RollbackVersion(ctx context.Context, bidId uuid.UUID, version int) (*entity.Bid, error)
}

type Repositories struct {
	Tender
	Employee
	Bid
}

func NewRepositories(pg *postgres.Postgres) *Repositories {
	return &Repositories{
		Tender:   pgdb.NewTenderRepo(pg),
		Employee: pgdb.NewEmployeeRepo(pg),
		Bid:      pgdb.NewBidRepo(pg),
	}
}
