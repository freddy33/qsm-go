select * from path_contexts;

select count(*) from growth_contexts;

select count(*) from points;

select avg(x) from points;

select path_ctx_id, count(*) from path_nodes
group by path_ctx_id;

select d, count(*) from path_nodes
where path_ctx_id = 2
group by d
order by d;


select ct.d, count(ct.id), avg(ct.cart_d), min(ct.cart_d), max(ct.cart_d) from
(select pn.d d, pn.id id, sqrt(p.x*p.x+p.y*p.y+p.z*p.z) cart_d
from path_nodes as pn
    join points as p on pn.point_id = p.id
where pn.path_ctx_id=2) ct
group by d;






