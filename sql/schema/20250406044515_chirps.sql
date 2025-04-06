-- +goose Up
-- +goose StatementBegin
create table chirps
(
    id         uuid primary key default gen_random_uuid(),
    body       text      not null,
    user_id    uuid      not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    FOREIGN KEY(user_id)
REFERENCES users(id)
on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table chirps;
-- +goose StatementEnd
