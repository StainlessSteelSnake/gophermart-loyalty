package database

const (
	queryInsertUser = `
	INSERT INTO public.users
	    (
			login, password
		)
	VALUES ($1, $2)
`
	querySelectPassword = `
	SELECT password 
	FROM public.users
	WHERE login = $1
`

	queryInsertOrder = `
	INSERT INTO public.orders
		(
 			id, user_login, status, uploaded
		)
	VALUES ($1, $2, 'NEW', $3)
`
	queryGetOrderUserByID = `
	SELECT user_login
	FROM public.orders
	WHERE id = $1
`
	queryGetOrdersByUser = `
	SELECT o.id, o.status, COALESCE(t.amount, 0), o.uploaded
	FROM public.orders AS o 
	LEFT JOIN public.transactions AS t 
	ON t.order_number = o.id AND t.type = 'ACCRUAL'
	WHERE o.user_login = $1
	ORDER BY o.uploaded ASC
`
	queryGetOrdersToProcess = `
	SELECT id, user_login, status, uploaded
	FROM public.orders
	WHERE status IN ('NEW', 'PROCESSING')
	ORDER BY uploaded ASC
`
	queryUpdateOrder = `
	UPDATE public.orders
	SET status = $2
	WHERE id = $1
`

	queryInsertUserAccount = `
	INSERT INTO public.accounts
	    (
			user_login, balance, withdrawn
		)
	VALUES ($1, $2, $3)
`
	queryGetUserAccount = `
	SELECT user_login, balance, withdrawn
	FROM public.accounts
	WHERE user_login = $1
`
	queryUpdateUserAccount = `
	UPDATE public.accounts
	SET balance = $2, withdrawn = $3
	WHERE user_login = $1
`

	queryInsertTransaction = `
	INSERT INTO public.transactions
	( order_number, user_login, type, amount, created_at)
	VALUES ($1, $2, $3, $4, $5)
`
	queryGetTransaction = `
	SELECT order_number, user_login, type, amount, created_at
	FROM  public.transactions
	WHERE order_number = $1
`
	queryGetTransactions = `
	SELECT order_number, user_login, type, amount, created_at
	FROM  public.transactions
	WHERE user_login = $1 AND type = $2
	ORDER BY created_at ASC
`
)
