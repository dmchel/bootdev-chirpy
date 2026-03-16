-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD hashed_password TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN hashed_password;
-- +goose StatementEnd
