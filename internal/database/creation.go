package database

const sqlCreateDatabase = `
	CREATE DATABASE "gophermart-loyalty"
	WITH
	OWNER = $1
	ENCODING = 'UTF8'
	CONNECTION LIMIT = -1
	IS_TEMPLATE = False;
	
	COMMENT ON DATABASE "gophermart-loyalty"
	IS 'Накопительная система лояльности "Гофермарт"'
`

const sqlCreateTableUsers = `
	CREATE TABLE IF NOT EXISTS public.users
	(
		login character varying COLLATE pg_catalog."default" NOT NULL,
		password character varying COLLATE pg_catalog."default" NOT NULL,
		CONSTRAINT users_pkey PRIMARY KEY (login)
	)
	
	TABLESPACE pg_default;	
`

const sqlCreateTableOrders = `
	CREATE TABLE IF NOT EXISTS public.orders
	(
		id character varying(20) COLLATE pg_catalog."default" NOT NULL,
		user_login character varying COLLATE pg_catalog."default" NOT NULL,
		status character varying(10) COLLATE pg_catalog."default" NOT NULL,
		uploaded timestamp with time zone NOT NULL,
		CONSTRAINT orders_pkey PRIMARY KEY (id)
	)
	
	TABLESPACE pg_default;
`

const sqlCreateTableAccounts = `
	CREATE TABLE IF NOT EXISTS public.accounts
	(
		user_login character varying COLLATE pg_catalog."default" NOT NULL,
		balance integer NOT NULL DEFAULT 0,
		withdrawn integer NOT NULL DEFAULT 0,
		CONSTRAINT accounts_pkey PRIMARY KEY (user_login)
	)
	
	TABLESPACE pg_default;
`

const sqlCreateTableTransactions = `
	CREATE TABLE IF NOT EXISTS public.transactions
	(
		order_number character varying COLLATE pg_catalog."default" NOT NULL,
		user_login character varying COLLATE pg_catalog."default" NOT NULL,
		type character varying(10) COLLATE pg_catalog."default" NOT NULL,
		amount integer NOT NULL DEFAULT 0,
		created_at timestamp with time zone NOT NULL,
		CONSTRAINT transactions_pkey PRIMARY KEY (order_number)
	)
	
	TABLESPACE pg_default;
`
