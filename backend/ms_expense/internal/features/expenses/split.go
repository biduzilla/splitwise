package expenses

import (
	"errors"
	"fmt"
	"sort"
	"uuid"
)

var ErrInvalidExactSplit = errors.New("sum of shares must equal total amount")

func CalculateEqualShares(total int64, userIDs []uuid.UUID) []Share {
	n := int64(len(userIDs))
	if n == 0 {
		return nil
	}

	base := total / n
	residue := total - base*n

	sorted := make([]uuid.UUID, len(userIDs))
	copy(sorted, userIDs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].String() < sorted[j].String()
	})

	shares := make([]Share, n)
	for i, id := range sorted {
		amt := base
		if int64(i) < residue {
			amt++
		}
		shares[i] = Share{UserID: id, Amount: amt}
	}
	return shares
}

func ValidateExactShares(amount int64, shares []ShareDTO) error {
	if len(shares) == 0 {
		return errors.New("at least one share is required")
	}

	seen := make(map[uuid.UUID]struct{}, len(shares))
	var sum int64
	for _, s := range shares {
		if s.UserID == uuid.Nil() {
			return errors.New("share has invalid user_id")
		}
		if s.Amount < 0 {
			return errors.New("share amount must be >= 0")
		}
		if _, dup := seen[s.UserID]; dup {
			return errors.New("duplicate user_id in shares")
		}
		seen[s.UserID] = struct{}{}
		sum += s.Amount
	}

	if sum != amount {
		return fmt.Errorf("%w (got %d, want %d)", ErrInvalidExactSplit, sum, amount)
	}
	return nil
}
