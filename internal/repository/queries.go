package repository

const (
	QueryInsertUser = `INSERT INTO users (id, email, password, role)
                        VALUES ($1, $2, $3, $4)
                        ON CONFLICT (email) DO NOTHING`

	QueryUserByEmail = `SELECT id, password, role
                         FROM users
                         WHERE email = $1`

	QueryInsertPVZ = `INSERT INTO pvz (id, city, registration_date)
							VALUES ($1, $2, $3)
							RETURNING id, city, registration_date;`
)
