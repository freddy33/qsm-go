select * from path_contexts;

select count(*) from growth_contexts;

select count(*) from points;

select avg(x) from points;

select path_ctx_id, count(*) from path_nodes
group by path_ctx_id;

select d, count(*) from path_nodes
where path_ctx_id = 35
group by d
order by d;

select pn.d, p.x, p.y, p.z, sqrt(p.x*p.x+p.y*p.y+p.z*p.z) cart_d
from path_nodes as pn
    join points as p on pn.point_id = p.id
where pn.path_ctx_id=35 and pn.d=53
order by cart_d;






