package storage

import (
	"database/sql"
	"time"

	"forex-bot/internal/models"
)

func (ps *PostgresStorage) CreatePaymentOrder(o *models.PaymentOrder) error {
	_, err := ps.db.Exec(
		`INSERT INTO payment_orders (
			id, user_id, amount_usdt, plan, status, merchant_trade_no,
			binance_prepay_id, checkout_url, pay_method, tx_reference,
			expires_at, paid_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		o.ID, o.UserID, o.AmountUSDT, o.Plan, o.Status, o.MerchantTradeNo,
		o.BinancePrepayID, o.CheckoutURL, o.PayMethod, o.TxReference,
		o.ExpiresAt, o.PaidAt, o.CreatedAt, o.UpdatedAt,
	)
	return err
}

func (ps *PostgresStorage) GetPaymentOrder(id string) (*models.PaymentOrder, error) {
	return ps.scanPaymentOrder(ps.db.QueryRow(
		`SELECT id, user_id, amount_usdt, plan, status, merchant_trade_no,
			binance_prepay_id, checkout_url, pay_method, tx_reference,
			expires_at, paid_at, created_at, updated_at
		 FROM payment_orders WHERE id = $1`, id,
	))
}

func (ps *PostgresStorage) GetPaymentOrderByMerchantTradeNo(no string) (*models.PaymentOrder, error) {
	return ps.scanPaymentOrder(ps.db.QueryRow(
		`SELECT id, user_id, amount_usdt, plan, status, merchant_trade_no,
			binance_prepay_id, checkout_url, pay_method, tx_reference,
			expires_at, paid_at, created_at, updated_at
		 FROM payment_orders WHERE merchant_trade_no = $1`, no,
	))
}

func (ps *PostgresStorage) UpdatePaymentOrder(o *models.PaymentOrder) error {
	_, err := ps.db.Exec(
		`UPDATE payment_orders SET
			status = $1, binance_prepay_id = $2, checkout_url = $3,
			tx_reference = $4, paid_at = $5, updated_at = $6
		 WHERE id = $7`,
		o.Status, o.BinancePrepayID, o.CheckoutURL, o.TxReference, o.PaidAt, o.UpdatedAt, o.ID,
	)
	return err
}

func (ps *PostgresStorage) ListPaymentOrdersByUser(userID int64, limit int) ([]*models.PaymentOrder, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := ps.db.Query(
		`SELECT id, user_id, amount_usdt, plan, status, merchant_trade_no,
			binance_prepay_id, checkout_url, pay_method, tx_reference,
			expires_at, paid_at, created_at, updated_at
		 FROM payment_orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.PaymentOrder
	for rows.Next() {
		o, err := ps.scanPaymentOrderRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (ps *PostgresStorage) scanPaymentOrder(row *sql.Row) (*models.PaymentOrder, error) {
	return ps.scanPaymentOrderRow(row)
}

func (ps *PostgresStorage) scanPaymentOrderRow(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.PaymentOrder, error) {
	var o models.PaymentOrder
	var paid sql.NullTime
	err := scanner.Scan(
		&o.ID, &o.UserID, &o.AmountUSDT, &o.Plan, &o.Status, &o.MerchantTradeNo,
		&o.BinancePrepayID, &o.CheckoutURL, &o.PayMethod, &o.TxReference,
		&o.ExpiresAt, &paid, &o.CreatedAt, &o.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if paid.Valid {
		o.PaidAt = &paid.Time
	}
	return &o, nil
}

func (ps *PostgresStorage) ExpireStalePaymentOrders() error {
	_, err := ps.db.Exec(
		`UPDATE payment_orders SET status = 'expired', updated_at = $1
		 WHERE status = 'pending' AND expires_at < $1`,
		time.Now(),
	)
	return err
}
