-- +goose Up
-- +goose StatementBegin
alter table users
    add column is_chirpy_red boolean default false not null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users
drop
column is_chripy_red;
-- +goose StatementEnd
