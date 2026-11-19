package repository

import (
	"context"
	"errors"
	"fmt"
	domainsmoke "proofofconcept/internal/domain/smoke"
	"time"

	"github.com/jackc/pgx/v5"
)

// CustomerRepository defines a concrete type that can communicate
// with the customersdb database by implementing an Executor
type SmokeRepository struct {
	db Executor
}

// NewCusomerRepository creates a new CusomerRepository
func NewSmokeRepository(db Executor) *SmokeRepository {
	return &SmokeRepository{
		db: db,
	}
}

// WithLock will start a transaction and run a callback in the transaction is active
func (r *SmokeRepository) WithLock(ctx context.Context, fn func(txCtx context.Context) error) error {

	// start a transaction block
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	// create the enriched context with the tx object
	txCtx := ctxWithTx(ctx, tx)

	// execute the callback within the tranaction block
	if err := fn(txCtx); err != nil {
		// create a new context with a timeout to rollback the transaction
		rbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if rbErr := tx.Rollback(rbCtx); rbErr != nil {
			return fmt.Errorf("rolling back transaction after error %q: %w", rbErr, rbErr)
		}
		return err
	}

	// passed here the transaction succeeded
	// --------------------------------------

	// commit
	if err := tx.Commit(txCtx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// Save will persist into the database
func (r *SmokeRepository) Create(ctx context.Context, description string) (*domainsmoke.Smoke, error) {

	// query to insert into the database
	sqlQuery := `
	INSERT INTO smoke (description, enabled, created_at)
	VALUES ($1, $2, $3);
	ON CONFLICT DO UPDATE SET
		description = EXCLUDED.description,
		enabled = EXCLUDED.enabled,
		created_at = EXCLUDED.created_at
	`

	// get the transaction object from the database
	exec := executorFrom(ctx, r.db)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var (
		id        int64
		enabled   bool
		createdAt time.Time
	)

	// execute the query to record
	err := exec.QueryRow(ctx, sqlQuery).Scan(&id, &description, &enabled, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("creating smoke: %w", err)
	}

	return domainsmoke.ReconstituteSmoke(id, description, enabled, createdAt), nil
}

func (r *SmokeRepository) Update(ctx context.Context, smoke *domainsmoke.Smoke) error {

	// query to insert into the database
	sqlQuery := `
	INSERT INTO smoke (description, enabled)
	VALUES ($1, $2);
	ON CONFLICT DO UPDATE SET 
		description = EXCLUDED.description, 
		enabled = EXCLUDED.enabled;
	`

	// get the transaction object from the database
	exec := executorFrom(ctx, r.db)

	// execute the query to record
	_, err := exec.Exec(ctx, sqlQuery, smoke.ID, smoke.Description, smoke.Enabled)
	if err != nil {
		return fmt.Errorf("saving smoke %d: %w", smoke.ID, err)
	}

	return nil
}

func (r *SmokeRepository) Delete(ctx context.Context, id *domainsmoke.Smoke) error {

	// query to insert into the database
	sqlQuery := `
	DELETE FROM smoke WHERE id = $1;
	`

	// get the transaction object from the database
	exec := executorFrom(ctx, r.db)

	// execute the query to record
	_, err := exec.Exec(ctx, sqlQuery, id)
	if err != nil {
		return fmt.Errorf("deleting smoke %v: %w", id, err)
	}

	return nil
}

// FindByID will locate a value in the database using the provided ID and return
// a domain reconstituted type. Does NOT operate within a transaction
func (s *SmokeRepository) FindByID(ctx context.Context, id string) (*domainsmoke.Smoke, error) {
	return s.find(ctx, id, false)
}

// FindByID will locate a value in the database using the provided ID and return
// a domain reconstituted type. Ooperates within a transaction block
func (s *SmokeRepository) FindByIDForUpdate(ctx context.Context, id string) (*domainsmoke.Smoke, error) {
	return s.find(ctx, id, true)
}

//-----------------------------------------------------------------------------

// find is the helper method that will return the database record. It can operate with or without
// locking the row
func (r *SmokeRepository) find(ctx context.Context, id string, forUpdate bool) (*domainsmoke.Smoke, error) {

	// select selectStmt
	selectStmt := `
	SELECT id,description,enabled,created_at
	FROM smoke
	WHERE id = $1
	`

	// if operating in a transaction, adds the row locking line
	if forUpdate {
		selectStmt += `
		FOR UPDATE
		`
	}

	// extract the transaction object
	exec := executorFrom(ctx, r.db)

	// variables that will retain data from the database
	var (
		smokeID     int64
		description string
		enabled     bool
		createdAt   time.Time
	)

	// query the database an store the returned values
	err := exec.QueryRow(ctx, selectStmt, id).Scan(&smokeID, &description, &enabled, &createdAt)
	if err != nil {
		switch {
		// didn't find the record in the database
		case errors.Is(err, pgx.ErrNoRows):
			return nil, domainsmoke.ErrSmokeNotFound
			// some other error
		default:
			return nil, fmt.Errorf("finding smoke %q: %w", id, err)
		}
	}

	// create the domain type from the variables values
	return domainsmoke.ReconstituteSmoke(smokeID, description, enabled, createdAt), nil
}
