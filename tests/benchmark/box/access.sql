
CREATE SCHEMA access;

CREATE TABLE access.auth_apikey (
    api_key text NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE access.auth_resources (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    action text NOT NULL,
    path text NOT NULL,
    method text NOT NULL,
    permitted_roles text[] DEFAULT '{}'::text[] NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE access.auth_users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    username text NOT NULL,
    password_hash text NOT NULL,
    groups text[] DEFAULT '{}'::text[] NOT NULL,
    roles text[] DEFAULT '{}'::text[] NOT NULL,
    active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY access.auth_apikey
    ADD CONSTRAINT auth_apikey_pkey PRIMARY KEY (api_key);

ALTER TABLE ONLY access.auth_apikey
    ADD CONSTRAINT auth_apikey_user_id_key UNIQUE (user_id);

ALTER TABLE ONLY access.auth_resources
    ADD CONSTRAINT auth_resources_pkey PRIMARY KEY (id);

ALTER TABLE ONLY access.auth_users
    ADD CONSTRAINT auth_users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY access.auth_users
    ADD CONSTRAINT auth_users_username_key UNIQUE (username);

CREATE INDEX idx_auth_resources_method ON access.auth_resources USING btree (method);

ALTER TABLE ONLY access.auth_apikey
    ADD CONSTRAINT auth_apikey_user_id_fkey FOREIGN KEY (user_id) REFERENCES access.auth_users(id) ON DELETE CASCADE;


INSERT INTO access.auth_resources (id, action, path, method, permitted_roles, created_at, updated_at) VALUES ('f5e3870b-1669-4276-b417-18a6a98c17df', 'kunjungan.create', '/api/kunjungan', 'POST', '{write}', '2026-08-27 08:07:23.194934+00', '2026-08-27 08:07:23.194934+00');
