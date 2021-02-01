create table if not exists cases
(
    id generated always as identity,
    country id int not null,
    entry_date timestamp not null,
    confirmed int default 0,
    deaths int default 0,
    recovered int default 0,
    confirmed_today int default 0,
    deaths_today int default 0,
    recovered_today int default 0,
)