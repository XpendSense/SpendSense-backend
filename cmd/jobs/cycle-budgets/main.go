package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mauro-afa91/spendsense/internal/db"
	"github.com/mauro-afa91/spendsense/internal/repository"
	"github.com/mauro-afa91/spendsense/internal/service"
	sqlcdb "github.com/mauro-afa91/spendsense/internal/sqlc"
)

// cycle-budgets is a daily job that finds every BudgetProfile whose latest period
// has ended (end_date < today) and creates the next period, pre-filling recurring
// income entries and carrying forward fixed+recurring transactions.
func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	queries := sqlcdb.New(pool)
	profileRepo      := repository.NewBudgetProfileRepository(queries)
	txRepo           := repository.NewTransactionRepository(queries)
	fixedExpenseRepo := repository.NewFixedExpenseRepository(queries)
	userRepo         := repository.NewUserRepository(queries)

	svc := service.NewBudgetProfileService(profileRepo, txRepo, fixedExpenseRepo, userRepo)

	today := time.Now().UTC()
	yesterday := today.AddDate(0, 0, -1)
	cutoff := pgtype.Date{Time: yesterday, Valid: true}

	profileIDs, err := profileRepo.ListProfileIDsWithExpiredPeriod(ctx, cutoff)
	if err != nil {
		log.Fatalf("list expired: %v", err)
	}

	log.Printf("cycling %d profiles", len(profileIDs))
	for _, id := range profileIDs {
		// CreateBudgetPeriod needs an owner check — use the profile directly.
		profile, err := profileRepo.GetByID(ctx, id)
		if err != nil {
			log.Printf("skip %s: get profile: %v", id, err)
			continue
		}
		_, err = svc.CreateBudgetPeriod(ctx, id, profile.UserID)
		if err != nil {
			log.Printf("skip %s: create period: %v", id, err)
			continue
		}
		log.Printf("cycled %s", id)
	}
}
