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
	SELECT id, user_login, status, uploaded
	FROM public.orders
	WHERE user_login = $1
`
)
