package generate_transaction_code

import (
	"github.com/google/uuid"
)

func GenerateUniqueID() string {
	id := uuid.New()
	return id.String()
}
