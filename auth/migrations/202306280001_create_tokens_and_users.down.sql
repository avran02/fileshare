-- +migrate Down
-- SQL in section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS tokens;

DROP TABLE IF EXISTS users;
