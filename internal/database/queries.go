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
)
