package database

const (
	queryInsertUser = `
	INSERT INTO public.users
	    (
			login, password
		)
	VALUES ($1, $2);
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
)
