CREATE TABLE test
(
       id serial NOT NULL,
       code text NOT NULL,
       text text NOT NULL,
       is_test boolean NOT NULL,
       created_at timestamp with time zone NOT NULL,
       CONSTRAINT UQ_code UNIQUE (code),
       PRIMARY KEY (code)
);
