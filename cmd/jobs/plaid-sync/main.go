package main

import (
	"context"
	"log"
	"os"

	"github.com/mauro-afa91/spendsense/internal/db"
	plaidclient "github.com/mauro-afa91/spendsense/internal/plaid"
	"github.com/mauro-afa91/spendsense/internal/repository"
	"github.com/mauro-afa91/spendsense/internal/service"
	sqlcdb "github.com/mauro-afa91/spendsense/internal/sqlc"
)

// plaid-sync fetches incremental transaction changes from Plaid for every active
// plaid_item, then writes new/updated/removed Variable transactions into the
// matching budget period with Plaid-resolved categories.
func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	clientID := os.Getenv("PLAID_CLIENT_ID")
	secret := os.Getenv("PLAID_SECRET")
	plaidEnv := os.Getenv("PLAID_ENV")
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if clientID == "" || secret == "" {
		log.Fatal("PLAID_CLIENT_ID and PLAID_SECRET are required")
	}
	if encryptionKey == "" {
		log.Fatal("ENCRYPTION_KEY is required")
	}
	if plaidEnv == "" {
		plaidEnv = "sandbox"
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	queries := sqlcdb.New(pool)
	plaidRepo := repository.NewPlaidRepository(queries)
	budgetRepo := repository.NewBudgetProfileRepository(queries)
	txRepo := repository.NewTransactionRepository(queries)
	feRepo := repository.NewFixedExpenseRepository(queries)
	reviewRepo := repository.NewTransactionReviewRepository(queries)

	pc, err := plaidclient.New(clientID, secret, plaidEnv)
	if err != nil {
		log.Fatalf("plaid: init client: %v", err)
	}

	svc := service.NewPlaidService(pc, plaidRepo, budgetRepo, nil, txRepo, feRepo, reviewRepo, encryptionKey)

	items, err := plaidRepo.ListActiveForSync(ctx)
	if err != nil {
		log.Fatalf("list active items: %v", err)
	}

	log.Printf("syncing %d plaid items", len(items))

	for _, item := range items {
		if err := svc.SyncItem(ctx, item); err != nil {
			log.Printf("item %s: sync failed: %v", item.ID, err)
		}
	}
}
