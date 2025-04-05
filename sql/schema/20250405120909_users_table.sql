-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id         uuid primary key default gen_random_uuid(),
    email      text      not null unique,
    created_at timestamp not null,
    updated_at timestamp not null
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
