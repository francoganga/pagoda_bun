SET statement_timeout = 0;

--bun:split


CREATE TABLE users (
    id serial primary key,
    name varchar not null,
    email varchar unique not null,
    password varchar not null,
    verified boolean,
    created_at timestamp not null default now()
);

--bun:split

CREATE TABLE password_tokens (
    id serial primary key,
    hash varchar not null,
    created_at timestamp not null default now(),
    user_id int references users(id)
);
