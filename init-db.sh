#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE ecommerce_users;
    CREATE DATABASE ecommerce_products;
    CREATE DATABASE ecommerce_orders;
    CREATE DATABASE ecommerce_payments;
EOSQL
