# Decisions for new rust QSM

1. No more postgreSQL database
1. The static ref tables will be created and loaded in memory on startup
1. All the previous db code for points, path_nodes, events and nodes should be replaced with writting and loading the data files made of parquet files with iceberg index
1. We run CLI calculation/simulation run of QSM with always the name of the simulation and we do not use an API server anymore
1. The run with the name should point to a folder where all the parquet file and iceberg index will be created, or already exists
1. A simulation can restart from a previous run point using the files if it's the same name
1. No more okteto and all references should be removed
1. No more remote API execution, all visualization should be done providing the name and access to the folder with the iceberg and parquet files
1. Loading the different rust objects directly from the files as we always need to load big block or full table most of the time

# Next steps
Create a main entry point that will be executed as `qsm [simulation name]`
Every run should create the default `data/[simulation name]` folder for data files
Replace all db logic to load and save parquet and iceberg files
The crate `m3db` should be renamed `m3data` and contain all the code allowing the load and save of the struct into the data folder
