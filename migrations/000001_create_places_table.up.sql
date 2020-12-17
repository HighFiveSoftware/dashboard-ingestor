create table if not exists places
(
    id             INT  not null,
    iso_code       VARCHAR(2),
    fips_id        TEXT not null default '',
    admin          TEXT not null default '',
    province_state TEXT not null default '',
    country_region TEXT not null,
    coordinate point,
    combined_key TEXT,
    population bigint not null default 0,

    primary key (id)
);