package expense

import (
	"fmt"
	"shared/apierror"
	"sort"
	"uuid"
)

func SplitEqual(total int64, userIDs []uuid.UUID) ([]Share, error) {
	if total <= 0 {
		return nil, apierror.NewBadRequestError(fmt.Errorf("total must be positive: %d", total))
	}

	n := int64(len(userIDs))
	if n == 0 {
		return nil, apierror.NewBadRequestError(fmt.Errorf("at least one participant required"))
	}

	sorted := make([]uuid.UUID, len(userIDs))
	copy(sorted, userIDs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].String() < sorted[j].String()
	})

	base := total / n
	resto := total - base*n

	shares := make([]Share, n)
	for i, id := range sorted {
		extra := int64(0)
		if int64(i) < resto {
			extra = 1
		}
		shares[i] = Share{UserID: id, Amount: base + extra}
	}
	return shares, nil
}
