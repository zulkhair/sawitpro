/**
  This is the SQL script that will be used to initialize the database schema.
  We will evaluate you based on how well you design your database.
  1. How you design the tables.
  2. How you choose the data types and keys.
  3. How you name the fields.
  In this assignment we will use PostgreSQL as the database.
  */

/** This is test table. Remove this table and replace with your own tables. */
CREATE TABLE IF NOT EXISTS test (
	id serial PRIMARY KEY,
	name VARCHAR ( 50 ) UNIQUE NOT NULL
);

INSERT INTO test (name) VALUES ('test1');
INSERT INTO test (name) VALUES ('test2');

CREATE TABLE IF NOT EXISTS public.user (
    id UUID PRIMARY KEY,
    phone VARCHAR ( 13 ) UNIQUE NOT NULL,
    name VARCHAR ( 60 ) NOT NULL,
    password VARCHAR ( 64 ) NOT NULL,
    salt VARCHAR ( 64 ) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);