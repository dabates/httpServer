-- +goose Up
-- +goose StatementBegin
alter table users
add column hashed_password text not null default 'unset';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users
drop column hashed_password;
-- +goose StatementEnd
