#!/bin/sh
set -e

cat <<EOF > /etc/pgbouncer/userlist.txt
"${POSTGRES_USER}" "${POSTGRES_PASSWORD}"
EOF

cat <<EOF > /etc/pgbouncer/pgbouncer.ini
[databases]
${POSTGRES_DB} = host=${POSTGRES_HOST} port=${POSTGRES_PORT} dbname=${POSTGRES_DB} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD}

[pgbouncer]
listen_addr = *
listen_port = ${PGBOUNCER_PORT}
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction
EOF

echo "Configuration file /etc/pgbouncer/pgbouncer.ini created:"
cat /etc/pgbouncer/pgbouncer.ini

exec pgbouncer /etc/pgbouncer/pgbouncer.ini