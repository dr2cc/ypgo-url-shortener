create table if not exists urls(
    created_by varchar(36) not null,
    original_url varchar unique not null,
    id varchar(12) unique not null,
    correlation_id varchar
)