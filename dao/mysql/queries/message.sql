-- name: GetMessage :many
select *
from message
where id >0
order by create_at
desc
limit 5;

-- name: CreateMessage :exec
insert into message(name, email, content, create_at)
values (?,?,?,now());