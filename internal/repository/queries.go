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

	QueryInsertReception = `WITH locked AS (
								SELECT id
								FROM pvz
								WHERE id = $1
								FOR UPDATE
							)
							INSERT INTO receptions (id, pvz_id, date_time, status)
							SELECT $2, locked.id, $3, 'in_progress'
							FROM locked
							RETURNING id`

	QueryCloseActiveReception = `WITH active AS (
									SELECT id 
									FROM receptions 
									WHERE pvz_id = $1 AND status = 'in_progress'
									ORDER BY date_time DESC
									LIMIT 1
									FOR UPDATE
									)
								UPDATE receptions
								SET status = 'close'
								WHERE id IN (SELECT id FROM active)
								RETURNING id;`
)
