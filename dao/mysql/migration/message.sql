create table message
(
    id        int auto_increment
        primary key,
    name      varchar(100)  not null,
    email     varchar(100)  not null,
    content   varchar(1024) null,
    create_at timestamp     null
);