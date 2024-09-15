package service

import (
	"avito/internal/entity"
	"avito/internal/repo"
	"context"
	"github.com/google/uuid"
	"time"
)

type Services struct {
	Tender   Tender
	Employee Employee
	Bid      Bid
}

type ServicesDependencies struct {
	Repos *repo.Repositories
}

type TenderCreateInput struct {
	Name            string
	Description     string
	ServiceType     string
	OrganizationId  uuid.UUID
	CreatorUsername string
}

type GetMyTendersInput struct {
	Username string
	Limit    int
	Offset   int
}

type GetMyTendersOutput struct {
	Id             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	ServiceType    string    `json:"serviceType"`
	OrganizationId uuid.UUID `json:"organizationTd"`
	Version        int       `json:"version"`
	CreatedAt      string    `json:"createdAt"`
}

type GetTendersInput struct {
	ServiceTypes []string
	Limit        int
	Offset       int
}
type GetStatusInput struct {
	Username string
	TenderId uuid.UUID
}

type PutStatusInput struct {
	Username string
	TenderId uuid.UUID
	Status   string
}
type PutStatusOutput struct {
	Id          uuid.UUID
	Name        string
	Description string
	Status      string
	ServiceType string
	Version     int
	CreatedAt   time.Time
}

type EditTenderInput struct {
	Id          uuid.UUID
	Name        string
	Description string
	ServiceType string
}
type EditTenderOutput struct {
	Id          uuid.UUID
	Name        string
	Description string
	Status      string
	ServiceType string
	Version     int
	CreatedAt   time.Time
}
type RollbackVersionInput struct {
	Id      uuid.UUID
	Version int
}
type RollbackVersionOutput struct {
	Id          uuid.UUID
	Name        string
	Description string
	Status      string
	ServiceType string
	Version     int
	CreatedAt   time.Time
}
type BidCreateInput struct {
	Name        string
	Description string
	TenderId    uuid.UUID
	AuthorType  string
	AuthorId    uuid.UUID
}

type GetMyBidsOutput struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	AuthorType string    `json:"authorType"`
	AuthorId   uuid.UUID `json:"authorTd"`
	Version    int       `json:"version"`
	CreatedAt  string    `json:"createdAt"`
}

type GetMyBidsInput struct {
	AuthorId uuid.UUID
	Limit    int
	Offset   int
}

type GetBidStatusInput struct {
	BidId    uuid.UUID
	Username string
}

type PutBidStatusInput struct {
	BidId  uuid.UUID
	Status string
}

type PutBidStatusOutput struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	AuthorType string    `json:"authorType"`
	AuthorId   uuid.UUID `json:"authorTd"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"createdAt"`
}

type EditBidInput struct {
	Id          uuid.UUID
	Name        string
	Description string
}
type EditBidOutput struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	AuthorType string    `json:"authorType"`
	AuthorId   uuid.UUID `json:"authorTd"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"createdAt"`
}

type RollbackBidVersionOutput struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	AuthorType string    `json:"authorType"`
	AuthorId   uuid.UUID `json:"authorTd"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Tender interface {
	CreateTender(ctx context.Context, input TenderCreateInput) (*entity.Tender, error)
	GetMyTenders(ctx context.Context, input GetMyTendersInput) ([]GetMyTendersOutput, error)
	GetTenders(ctx context.Context, input GetTendersInput) ([]GetMyTendersOutput, error)
	GetStatus(ctx context.Context, input GetStatusInput) (string, error)
	PutStatus(ctx context.Context, input PutStatusInput) (*PutStatusOutput, error)
	EditTender(ctx context.Context, input EditTenderInput) (*EditTenderOutput, error)
	RollbackVersion(ctx context.Context, input RollbackVersionInput) (*RollbackVersionOutput, error)
	GetTenderById(ctx context.Context, id uuid.UUID) (*entity.Tender, error)
}
type Bid interface {
	CreateBid(ctx context.Context, input BidCreateInput) (*entity.Bid, error)
	GetMyBids(ctx context.Context, input GetMyBidsInput) ([]GetMyBidsOutput, error)
	GetStatus(ctx context.Context, input GetBidStatusInput) (string, error)
	GetBidById(ctx context.Context, id uuid.UUID) (*entity.Bid, error)
	PutStatus(ctx context.Context, input PutBidStatusInput) (*PutBidStatusOutput, error)
	EditBid(ctx context.Context, input EditBidInput) (*EditBidOutput, error)
	RollbackVersion(ctx context.Context, input RollbackVersionInput) (*RollbackBidVersionOutput, error)
	SubmitDecision(ctx context.Context, tenderId, bidId uuid.UUID, decision string) (*entity.Bid, error)
}

type Employee interface {
	GetEmployeeIdByUsername(ctx context.Context, username string) (uuid.UUID, error)
	GetEmployeeById(ctx context.Context, id uuid.UUID) (*entity.Employee, error)
	GetEmployeeOrgIdById(ctx context.Context, employeeId uuid.UUID) (uuid.UUID, error)
}

func NewServices(deps ServicesDependencies) *Services {
	return &Services{
		Tender:   NewTenderService(deps.Repos.Tender),
		Employee: NewEmployeeService(deps.Repos.Employee),
		Bid:      NewBidService(deps.Repos.Bid, deps.Repos.Tender),
	}
}
