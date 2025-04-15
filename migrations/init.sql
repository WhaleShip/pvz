CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('employee', 'moderator'))
);

CREATE TABLE pvz (
    id UUID PRIMARY KEY,
    city VARCHAR(255) NOT NULL,
    registration_date TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE receptions (
    id UUID PRIMARY KEY,
    pvz_id UUID NOT NULL,
    date_time TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL CHECK (status IN ('in_progress', 'close')),
    CONSTRAINT fk_receptions_pvz
        FOREIGN KEY (pvz_id)
            REFERENCES pvz(id)
            ON DELETE CASCADE
);

CREATE TABLE products (
    id UUID PRIMARY KEY,
    reception_id UUID NOT NULL,
    date_time TIMESTAMP NOT NULL DEFAULT NOW(),
    type VARCHAR(50) NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
    CONSTRAINT fk_products_reception
        FOREIGN KEY (reception_id)
            REFERENCES receptions(id)
            ON DELETE CASCADE
);
