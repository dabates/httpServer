-- +goose Up
-- +goose StatementBegin
create table refresh_tokens(
    token text primary key,
    user_id uuid not null ,
    expires_at timestamp default CURRENT_TIMESTAMP,
    revoked_at timestamp default null,
    created_at timestamp not null ,
    updated_at timestamp not null ,
    FOREIGN KEY(user_id)
        REFERENCES users(id)
        on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table refresh_tokens;
-- +goose StatementEnd
