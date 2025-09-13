package crypto

import "github.com/btouchard/ackify-ce/internal/domain/models"

// Internal test fixtures to avoid external dependencies

var (
	testUserAlice = &models.User{
		Sub:   "user-123-alice",
		Email: "alice@example.com",
		Name:  "Alice Smith",
	}

	testUserBob = &models.User{
		Sub:   "user-456-bob",
		Email: "bob@example.com",
		Name:  "Bob Johnson",
	}

	testUserCharlie = &models.User{
		Sub:   "user-789-charlie",
		Email: "charlie@example.com",
		Name:  "Charlie Brown",
	}
)
