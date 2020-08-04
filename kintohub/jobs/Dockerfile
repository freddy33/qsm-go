FROM postgres:12.3

USER postgres

ENTRYPOINT ["psql", "-h", "db", "-Uqsm", "--list"]