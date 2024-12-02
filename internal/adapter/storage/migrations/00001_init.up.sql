BEGIN TRANSACTION;

CREATE TYPE OrderStatus as enum ('NEW', 'PROCESSING', 'PROCESSED', 'INVALID');

CREATE TABLE
	users (
		id bigserial PRIMARY KEY,
		login varchar NOT NULL,
		"password" varchar NOT NULL,
		CONSTRAINT users_unique UNIQUE (login)
	);

CREATE TABLE
	orders (
		user_id int8 NOT NULL,
		"number" varchar NOT NULL,
		accrual numeric(15, 2) NOT NULL,
		withdrawal numeric(15, 2) NOT NULL,
		status public."orderstatus" NOT NULL,
		uploaded_at timestamp NOT NULL,
		CONSTRAINT orders_pk PRIMARY KEY (number),
		CONSTRAINT orders_users_fk FOREIGN KEY (user_id) REFERENCES users (id)
	);

CREATE TABLE
	public.balance (
		user_id int8 NOT NULL,
		"current" numeric(15, 2) NOT NULL,
		withdrawn numeric(15, 2) not NULL,
		CONSTRAINT balance_pk PRIMARY KEY (user_id)
	);

COMMIT TRANSACTION;