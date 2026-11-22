package balance

import (
	"sort"
	"uuid"
)

type Transfer struct {
	From   uuid.UUID `json:"from_user_id"`
	To     uuid.UUID `json:"to_user_id"`
	Amount int64     `json:"amount"`
}

type account struct {
	ID     uuid.UUID
	Amount int64
}

func Minimize(balances map[uuid.UUID]int64) []Transfer {
	var creditors, debtors []account
	for id, bal := range balances {
		if bal > 0 {
			creditors = append(creditors, account{id, bal})
		} else if bal < 0 {
			debtors = append(debtors, account{id, -bal})
		}
	}

	sort.Slice(creditors, func(i, j int) bool {
		return creditors[i].Amount > creditors[j].Amount
	})

	sort.Slice(debtors, func(i, j int) bool {
		return debtors[i].Amount > debtors[j].Amount
	})

	var result []Transfer
	i, j := 0, 0
	for i < len(debtors) && j < len(creditors) {
		d, c := &debtors[i], &creditors[j]
		amount := min(d.Amount, c.Amount)

		result = append(result, Transfer{
			From:   d.ID,
			To:     c.ID,
			Amount: amount,
		})

		d.Amount -= amount
		c.Amount -= amount
		if d.Amount == 0 {
			i++
		}
		if c.Amount == 0 {
			j++
		}

	}
	return result
}
