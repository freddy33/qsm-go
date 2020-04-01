select *
from path_contexts;

select count(*)
from growth_contexts;

select count(*)
from points;

select avg(x)
from points;

select path_ctx_id, count(*)
from path_nodes
group by path_ctx_id;

select d, count(*)
from path_nodes
where path_ctx_id = 1
group by d
order by d;


select pn.d, p.x, p.y, p.z, sqrt(p.x*p.x+p.y*p.y+p.z*p.z) cart_d
from path_nodes as pn
    join points as p on pn.point_id = p.id
where pn.path_ctx_id=35 and pn.d<=6
order by cart_d;


select ct.d as D,
       count(ct.id) as NB,
       avg(ct.cart_d) as avg_cart_d,
       avg(ct.cart_d)/ct.d as "ratio(avg/d)",
       min(ct.cart_d) as min_cart_d,
       max(ct.cart_d) as max_cart_d
from (select pn.d d, pn.id id, sqrt(p.x * p.x + p.y * p.y + p.z * p.z) cart_d
      from path_nodes as pn
               join points as p on pn.point_id = p.id
      where pn.path_ctx_id = 2) ct
group by d;

select d
        , count(id) FILTER (WHERE connection_mask = 273) AS NB_3F
        , count(id) FILTER (WHERE connection_mask in (274, 289, 529)) AS NB_2F
        , count(id) FILTER (WHERE connection_mask in (290, 530, 545)) AS NB_1F
from path_nodes pn
where pn.path_ctx_id = 1
group by d;

select d, id, path_node1, path_node2, path_node3,
       (case
            when connection_mask = 546 then 'center'
            when connection_mask = 273 then '3F'
            when connection_mask in (274, 289, 529) then '2F'
            when connection_mask in (290, 530, 545) then '1F'
            else 'open'
           end) mask_display
from path_nodes pn
where pn.path_ctx_id = 1 and pn.d <= 5;

select ct.d, count(ct.id), min(ct.cart_d), max(ct.cart_d),
       avg(ct.cart_d),avg(ct.cart_d)/ct.d "ratio(avg/d)"
from (select pn.d d, pn.id id, sqrt(p.x * p.x + p.y * p.y + p.z * p.z) cart_d
      from path_nodes as pn
               join points as p on pn.point_id = p.id
      where pn.path_ctx_id = 35 and pn.d>=1 and pn.d<=53) ct
group by d
order by d;



-- All three from: 273
-- 2 froms: 274 289 529
-- 1 from: 290 530 545
select trio_id, (case
            when connection_mask = 546 then 'center'
            when connection_mask = 273 then '3F'
            when connection_mask in (274, 289, 529) then '2F'
            when connection_mask in (290, 530, 545) then '1F'
            else 'open'
           end) mask_display,
       count(*)
from path_nodes
where path_ctx_id = 1 and trio_id <= 8
group by trio_id, mask_display;

select pn.trio_id, count(*)
from path_nodes pn
where pn.path_ctx_id = 1 and connection_mask = 273 and pn.trio_id < 8
group by pn.trio_id;

select pn.trio_id, p.x, p.y, p.z, pn.path_node1, pn.path_node2, pn.path_node3
from path_nodes pn
         join points as p on pn.point_id = p.id
where path_ctx_id = 1 and d <= 12 and connection_mask = 273 and pn.trio_id < 8;

select d, cast(connection_mask::int4 as bit(12)) conn_mask, count(*)
from path_nodes
where path_ctx_id = 1
  and d = 0
group by d, conn_mask;

select d, connection_mask conn_mask, count(*)
from path_nodes
where path_ctx_id = 1
  and d = 0
group by d, conn_mask;




