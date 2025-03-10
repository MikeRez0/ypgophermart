package repository

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/MikeRez0/ypgophermart/internal/adapter/storage"
	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
	"github.com/govalues/decimal"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	db *storage.DB
}

type queryAble interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func NewRepository(db *storage.DB) (*Repository, error) {
	return &Repository{db: db}, nil
}

func (r *Repository) insertOrder(ctx context.Context, tx queryAble, order *domain.Order) (*domain.Order, error) {
	statement := r.db.QueryBuilder.Insert("orders").
		Columns("user_id", "number", "accrual", "withdrawal", "status", "uploaded_at").
		Values(order.UserID, order.Number, order.Accrual, order.Withdrawal, order.Status, order.UploadedAt)

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, domain.ErrConflictingData
		}
		return nil, err
	}
	return order, nil
}

func (r *Repository) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	return r.insertOrder(ctx, r.db.Pool, order)
}

func (r *Repository) selectOrder(ctx context.Context, tx queryAble, orderID domain.OrderNumber, forUpdate bool) (*domain.Order, error) {
	statement := r.db.QueryBuilder.
		Select("user_id", "number", "accrual", "withdrawal", "status", "uploaded_at").
		From("orders").
		Where(sq.Eq{"number": orderID})
	if forUpdate {
		statement = statement.Suffix("FOR UPDATE")
	}

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	order := domain.Order{}

	err = tx.QueryRow(ctx, sql, args...).Scan(
		&order.UserID,
		&order.Number,
		&order.Accrual,
		&order.Withdrawal,
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

func (r *Repository) ReadOrder(ctx context.Context, orderID domain.OrderNumber) (*domain.Order, error) {
	return r.selectOrder(ctx, r.db.Pool, orderID, false)
}

func (r *Repository) updateOrder(ctx context.Context, tx queryAble, order *domain.Order) (*domain.Order, error) {
	orderSt := r.db.QueryBuilder.
		Update("orders").
		Set("accrual", order.Accrual).
		Set("withdrawal", order.Withdrawal).
		Set("status", order.Status).
		Where(sq.Eq{"number": order.Number})

	sql, args, err := orderSt.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return order, nil
}
func (r *Repository) UpdateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	return r.updateOrder(ctx, r.db.Pool, order)
}

func (r *Repository) listOrders(ctx context.Context, tx queryAble,
	userID uint64, statusList []domain.OrderStatus) ([]*domain.Order, error) {
	statement := r.db.QueryBuilder.
		Select("user_id", "number", "accrual", "withdrawal", "status", "uploaded_at").
		From("orders")

	if userID != 0 {
		statement = statement.Where(sq.Eq{"user_id": userID})
	}

	if len(statusList) > 0 {
		statement = statement.Where(sq.Eq{"status": statusList})
	}

	statement = statement.OrderBy("uploaded_at desc")

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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
			&order.Withdrawal,
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

func (r *Repository) ListOrdersByUser(ctx context.Context, userID uint64) ([]*domain.Order, error) {
	return r.listOrders(ctx, r.db.Pool, userID, nil)
}
func (r *Repository) ListOrdersByStatus(ctx context.Context,
	statusList []domain.OrderStatus) ([]*domain.Order, error) {
	return r.listOrders(ctx, r.db.Pool, 0, statusList)
}

func (r *Repository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		userSt := r.db.QueryBuilder.
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

		balanceSt := r.db.QueryBuilder.
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

func (r *Repository) selectUser(ctx context.Context, tx queryAble, login string, forUpdate bool) (*domain.User, error) {
	statement := r.db.QueryBuilder.
		Select("id", "login", "password").
		From("users").
		Where(sq.Eq{"login": login})

	if forUpdate {
		statement = statement.Suffix("FOR UPDATE")
	}

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	user := domain.User{}

	err = tx.QueryRow(ctx, sql, args...).Scan(
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

func (r *Repository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	return r.selectUser(ctx, r.db.Pool, login, false)
}

func (r *Repository) selectBalanceByUserID(ctx context.Context, tx queryAble, userID uint64, forUpdate bool) (*domain.Balance, error) {
	statement := r.db.QueryBuilder.
		Select("user_id", "current", "withdrawn").
		From("balance").
		Where(sq.Eq{"user_id": userID})
	if forUpdate {
		statement = statement.Suffix("FOR UPDATE")
	}

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	balance := domain.Balance{}

	err = tx.QueryRow(ctx, sql, args...).Scan(
		&balance.UserID,
		&balance.Current,
		&balance.Withdrawn,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &balance, nil
}

func (r *Repository) ReadBalanceByUserID(ctx context.Context, userID uint64) (*domain.Balance, error) {
	return r.selectBalanceByUserID(ctx, r.db.Pool, userID, false)
}
func (r *Repository) UpdateUserBalanceByOrder(ctx context.Context,
	order *domain.Order, isNewOrder bool, updateFn port.UpdateBalanceFn) (*domain.Balance, error) {

	if order == nil {
		return nil, domain.ErrBadRequest

	}

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		balance, err := r.selectBalanceByUserID(ctx, tx, order.UserID, true)
		if err != nil {
			return err
		}

		if isNewOrder {
			order, err = r.insertOrder(ctx, tx, order)
			if err != nil {
				return err
			}
		} else {
			order, err = r.selectOrder(ctx, tx, order.Number, true)
			if err != nil {
				return err
			}
		}

		err = updateFn(balance, order)
		if err != nil {
			return err
		}

		_, err = r.updateOrder(ctx, tx, order)
		if err != nil {
			return err
		}

		balanceSt := r.db.QueryBuilder.
			Update("balance").
			Set("current", balance.Current).
			Set("withdrawn", balance.Withdrawn).
			Where(sq.Eq{"user_id": balance.UserID})

		sql, args, err := balanceSt.ToSql()
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
		return nil, err
	}

	return r.selectBalanceByUserID(ctx, r.db.Pool, order.UserID, false)
}
