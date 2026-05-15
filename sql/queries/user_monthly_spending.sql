-- name: AddMonthlySpending :exec
INSERT INTO user_monthly_spending (user_id, year_month, amount, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (user_id, year_month)
DO UPDATE SET 
    amount = user_monthly_spending.amount + EXCLUDED.amount,
    updated_at = NOW();

-- name: GetTotalSpendingFromSummary :one
SELECT COALESCE(SUM(amount), 0)::TEXT
FROM user_monthly_spending
WHERE user_id = $1 AND year_month >= $2;
