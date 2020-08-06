FROM postgres:12.3

USER postgres

ENTRYPOINT ["psql", "--list"]