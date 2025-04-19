package repository

const (
	// user
	QueryInsertUser = `INSERT INTO users (id, email, password, role)
                        VALUES ($1, $2, $3, $4)
                        ON CONFLICT (email) DO NOTHING`

	QueryUserByEmail = `SELECT id, password, role
                         FROM users
                         WHERE email = $1`

	// pvz
	QueryInsertPVZ = `INSERT INTO pvz (id, city, registration_date)
							VALUES ($1, $2, $3)
							RETURNING id, city, registration_date;`

	QuerySelectPVZByOpenReceptions = `WITH qualified_pvzs AS (
										SELECT DISTINCT pvz_id
										FROM receptions
										WHERE date_time <= $2
										AND (close_date_time >= $1 OR close_date_time IS NULL)
									)
									SELECT id, city, registration_date
									FROM pvz
									WHERE id IN (SELECT pvz_id FROM qualified_pvzs)
									ORDER BY registration_date DESC
									LIMIT $3 OFFSET $4;`

	QuerySelectAllPVZs = `SELECT id, city, registration_date FROM pvz`

	// recepiton
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
								SET 
									status = 'close',
									close_date_time = NOW()
								WHERE id IN (SELECT id FROM active)
								RETURNING id, date_time;`

	QueryGetReceptionsByPVZs = `SELECT id, pvz_id, date_time, status
								FROM receptions
								WHERE pvz_id = $1
								ORDER BY date_time DESC`

	// products
	QueryInsertProduct = `WITH active_reception AS (
								SELECT id FROM receptions 
								WHERE pvz_id = $1 AND status = 'in_progress'
								ORDER BY date_time DESC 
								LIMIT 1 
								FOR UPDATE
							)
							INSERT INTO products (id, reception_id, date_time, type)
							SELECT $2, id, $3, $4
							FROM active_reception
							RETURNING id;`

	QueryDeleteLastProduct = `WITH last AS (
								SELECT p.id
								FROM receptions r
								JOIN products p ON p.reception_id = r.id
								WHERE r.pvz_id = $1 AND r.status = 'in_progress'
								ORDER BY r.date_time DESC, p.date_time DESC
								LIMIT 1
								FOR UPDATE SKIP LOCKED
							)
							DELETE FROM products
							WHERE id = (SELECT id FROM last)
							RETURNING *;`

	QueryGetProductsByReceptions = `SELECT id, reception_id, date_time, type
							FROM products
							WHERE reception_id = ANY($1)
							ORDER BY date_time DESC`
)
