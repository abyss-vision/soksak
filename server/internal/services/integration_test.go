package services_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	"abyss-view/internal/db"
	"abyss-view/internal/services"
	"abyss-view/internal/testutil"
)

// setupDB starts a test DB and runs all migrations.
func setupDB(t *testing.T) *sqlx.DB {
	t.Helper()
	database := testutil.SetupTestDB(t)
	ctx := context.Background()
	if err := db.RunMigrations(ctx, database.DB); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	return database
}

func TestCompanyService_CreateListGet(t *testing.T) {
	database := setupDB(t)
	ctx := context.Background()

	svc := services.NewCompanyService(database)

	// List — should be empty initially.
	companies, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(companies) != 0 {
		t.Errorf("initial List: expected 0, got %d", len(companies))
	}

	// Create a company.
	created, err := svc.Create(ctx, services.CreateCompanyInput{
		Name:        "Test Corp",
		IssuePrefix: "TC",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.UUID == "" {
		t.Error("created company UUID is empty")
	}
	if created.Name != "Test Corp" {
		t.Errorf("Name = %q, want %q", created.Name, "Test Corp")
	}

	// List — should have one company now.
	companies, err = svc.List(ctx)
	if err != nil {
		t.Fatalf("List after create: %v", err)
	}
	if len(companies) != 1 {
		t.Errorf("List after create: expected 1, got %d", len(companies))
	}

	// Get by UUID.
	got, err := svc.Get(ctx, created.UUID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Test Corp" {
		t.Errorf("Get name = %q, want %q", got.Name, "Test Corp")
	}
}

func TestCompanyService_Update(t *testing.T) {
	database := setupDB(t)
	ctx := context.Background()
	svc := services.NewCompanyService(database)

	created, err := svc.Create(ctx, services.CreateCompanyInput{
		Name:        "Original Corp",
		IssuePrefix: "OC",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	newName := "Updated Corp"
	updated, err := svc.Update(ctx, created.UUID, services.UpdateCompanyInput{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("updated name = %q, want %q", updated.Name, newName)
	}
}

func TestCompanyService_Delete(t *testing.T) {
	database := setupDB(t)
	ctx := context.Background()
	svc := services.NewCompanyService(database)

	created, err := svc.Create(ctx, services.CreateCompanyInput{
		Name:        "To Delete Corp",
		IssuePrefix: "TD",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(ctx, created.UUID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// After delete, Get should fail.
	_, err = svc.Get(ctx, created.UUID)
	if err == nil {
		t.Fatal("Get after Delete: expected error, got nil")
	}
}

func TestCompanyService_Get_NotFound(t *testing.T) {
	database := setupDB(t)
	ctx := context.Background()
	svc := services.NewCompanyService(database)

	_, err := svc.Get(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("Get nonexistent: expected error, got nil")
	}
}
