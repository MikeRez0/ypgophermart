package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/MikeRez0/ypgophermart/internal/adapter/storage"
	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/govalues/decimal"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	db *storage.DB
}

func NewRepository(db *storage.DB) (*Repository, error) {
	return &Repository{db: db}, nil
}

func (or *Repository) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	statement := or.db.QueryBuilder.Insert("orders").
		Columns("user_id", "number", "accrual", "status", "uploaded_at").
		Values(order.UserID, order.Number, order.Accrual, order.Status, order.UploadedAt)

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = or.db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, domain.ErrConflictingData
		}
		return nil, err
	}
	return order, nil
}

func (or *Repository) ReadOrder(ctx context.Context, orderID uint64) (*domain.Order, error) {
	statement := or.db.QueryBuilder.
		Select("user_id", "number", "accrual", "status", "uploaded_at").
		From("orders").
		Where(sq.Eq{"number": orderID})

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	order := domain.Order{}

	err = or.db.QueryRow(ctx, sql, args...).Scan(
		&order.UserID,
		&order.Number,
		&order.Accrual,
		&order.Status,
		&order.UploadedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &order, nil

}
func (or *Repository) UpdateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	return nil, nil
}

func (or *Repository) ListOrdersByUser(ctx context.Context, userID uint64) ([]*domain.Order, error) {
	statement := or.db.QueryBuilder.
		Select("user_id", "number", "accrual", "status", "uploaded_at").
		From("orders").
		Where(sq.Eq{"user_id": userID})

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := or.db.Query(ctx, sql, args...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	list := make([]*domain.Order, 0)
	for rows.Next() {
		order := domain.Order{}
		err := rows.Scan(
			&order.UserID,
			&order.Number,
			&order.Accrual,
			&order.Status,
			&order.UploadedAt,
		)
		list = append(list, &order)
		if err != nil {
			return nil, err
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return list, nil
}
func (or *Repository) ListOrdersByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	return nil, nil
}

func (or *Repository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	err := pgx.BeginFunc(ctx, or.db, func(tx pgx.Tx) error {
		userSt := or.db.QueryBuilder.
			Insert("users").
			Columns("login", "password").
			Values(user.Login, user.Password).
			Suffix("returning id")

		sql, args, err := userSt.ToSql()
		if err != nil {
			return err
		}

		err = tx.QueryRow(ctx, sql, args...).Scan(&(user.ID))
		if err != nil {
			return err
		}

		balanceSt := or.db.QueryBuilder.
			Insert("balance").
			Columns("user_id", "current", "withdrawn").
			Values(user.ID, decimal.Zero, decimal.Zero)

		sql, args, err = balanceSt.ToSql()
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, domain.ErrConflictingData
		}
		return nil, err
	}

	return user, nil
}
func (or *Repository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	statement := or.db.QueryBuilder.
		Select("id", "login", "password").
		From("users").
		Where(sq.Eq{"login": login})

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	user := domain.User{}

	err = or.db.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (or *Repository) ReadBalanceByUserID(ctx context.Context, userID uint64) (*domain.Balance, error) {
	return nil, nil
}
func (or *Repository) UpdateBalance(ctx context.Context, balance *domain.Balance) (*domain.Balance, error) {
	return nil, nil
}
