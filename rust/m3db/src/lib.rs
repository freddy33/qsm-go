//! PostgreSQL storage for the derived point pack.
//!
//! Table names, column lists, and expected row counts match the Go `m3db` /
//! `pointdb` definitions. Connections are opened with sqlx. Tests that talk to
//! a server run only when `QSM_DATABASE_URL` is set.

use m3point::{ConnectionDetails, GrowthContext, PointPackData, TrioDetails};
use m3util::{QsmEnvId, QsmError};
use sqlx::postgres::PgPoolOptions;
use sqlx::PgPool;
use std::time::Duration;

pub const CONNECTION_DETAILS_TABLE: &str = "connection_details";
pub const TRIO_DETAILS_TABLE: &str = "trio_details";
pub const GROWTH_CONTEXTS_TABLE: &str = "growth_contexts";
pub const TRIO_CUBES_TABLE: &str = "trio_cubes";
pub const PATH_BUILDERS_TABLE: &str = "path_builders";

pub const EXPECTED_CONNECTIONS: i32 = 50;
pub const EXPECTED_TRIOS: i32 = 200;
pub const EXPECTED_GROWTH_CONTEXTS: i32 = 52;
pub const EXPECTED_CUBES: i32 = 5192;

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct DbConnDetails {
    pub host: String,
    pub port: u16,
    pub user: String,
    pub password: String,
    pub db_name: String,
}

impl DbConnDetails {
    pub fn connection_string(&self) -> String {
        format!(
            "host={} port={} user={} password={} dbname={} sslmode=disable",
            self.host, self.port, self.user, self.password, self.db_name
        )
    }

    /// URI form accepted by sqlx.
    pub fn url(&self) -> String {
        format!(
            "postgres://{}:{}@{}:{}/{}?sslmode=disable",
            self.user, self.password, self.host, self.port, self.db_name
        )
    }
}

#[derive(Clone, Debug)]
pub struct DbEnvironment {
    pub id: QsmEnvId,
    pub details: DbConnDetails,
    pub schema_name: String,
}

impl DbEnvironment {
    pub fn new(id: QsmEnvId, details: DbConnDetails) -> Self {
        let schema_name = format!("qsm{}", id as i16);
        Self {
            id,
            details,
            schema_name,
        }
    }

    pub fn from_env(id: QsmEnvId) -> Result<Self, QsmError> {
        let host = std::env::var("QSM_DB_HOST").unwrap_or_else(|_| "localhost".to_string());
        let port = std::env::var("QSM_DB_PORT")
            .ok()
            .and_then(|p| p.parse().ok())
            .unwrap_or(5432u16);
        let user =
            std::env::var("QSM_DB_USER").map_err(|_| QsmError::new("missing QSM_DB_USER"))?;
        let password = std::env::var("QSM_DB_PASSWORD").unwrap_or_default();
        let db_name =
            std::env::var("QSM_DB_NAME").map_err(|_| QsmError::new("missing QSM_DB_NAME"))?;
        Ok(Self::new(
            id,
            DbConnDetails {
                host,
                port,
                user,
                password,
                db_name,
            },
        ))
    }
}

pub fn block_on<F: std::future::Future>(fut: F) -> F::Output {
    tokio::runtime::Builder::new_current_thread()
        .enable_all()
        .build()
        .expect("tokio runtime")
        .block_on(fut)
}

pub async fn connect(env: &DbEnvironment) -> Result<PgPool, QsmError> {
    PgPoolOptions::new()
        .max_connections(5)
        .acquire_timeout(Duration::from_secs(5))
        .connect(&env.details.url())
        .await
        .map_err(|err| {
            QsmError::wrap(
                err,
                format!("fail to open DB for environment {}", env.id as i16),
            )
        })
}

pub fn schema_ddl(schema: &str) -> String {
    format!(
        "CREATE SCHEMA IF NOT EXISTS {schema};\n\
         CREATE TABLE IF NOT EXISTS {schema}.{CONNECTION_DETAILS_TABLE} (\n\
             id smallint PRIMARY KEY,\n\
             x integer,\n\
             y integer,\n\
             z integer,\n\
             ds bigint\n\
         );\n\
         CREATE TABLE IF NOT EXISTS {schema}.{TRIO_DETAILS_TABLE} (\n\
             id smallint PRIMARY KEY,\n\
             conn1 smallint REFERENCES {schema}.{CONNECTION_DETAILS_TABLE} (id),\n\
             conn2 smallint REFERENCES {schema}.{CONNECTION_DETAILS_TABLE} (id),\n\
             conn3 smallint REFERENCES {schema}.{CONNECTION_DETAILS_TABLE} (id)\n\
         );\n\
         CREATE TABLE IF NOT EXISTS {schema}.{GROWTH_CONTEXTS_TABLE} (\n\
             id smallint PRIMARY KEY,\n\
             ctx_type smallint,\n\
             ctx_index smallint,\n\
             UNIQUE (ctx_type, ctx_index)\n\
         );\n"
    )
}

pub fn drop_schema_sql(schema: &str) -> String {
    format!("DROP SCHEMA IF EXISTS {schema} CASCADE")
}

pub async fn ensure_schema(pool: &PgPool, env: &DbEnvironment) -> Result<(), QsmError> {
    let sql = schema_ddl(&env.schema_name);
    for statement in sql.split(';').map(str::trim).filter(|s| !s.is_empty()) {
        sqlx::query(statement)
            .execute(pool)
            .await
            .map_err(|err| QsmError::wrap(err, format!("schema statement failed: {statement}")))?;
    }
    Ok(())
}

pub async fn save_connections(
    pool: &PgPool,
    schema: &str,
    rows: &[ConnectionDetails],
) -> Result<u64, QsmError> {
    let mut n = 0u64;
    for cd in rows {
        let sql = format!(
            "INSERT INTO {schema}.{CONNECTION_DETAILS_TABLE} (id, x, y, z, ds) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (id) DO NOTHING"
        );
        let res = sqlx::query(&sql)
            .bind(i16::from(cd.id.0))
            .bind(cd.vector.x())
            .bind(cd.vector.y())
            .bind(cd.vector.z())
            .bind(cd.conn_ds)
            .execute(pool)
            .await
            .map_err(|err| QsmError::wrap(err, "insert connection"))?;
        n += res.rows_affected();
    }
    Ok(n)
}

pub async fn save_trios(
    pool: &PgPool,
    schema: &str,
    rows: &[TrioDetails],
) -> Result<u64, QsmError> {
    let mut n = 0u64;
    for td in rows {
        let sql = format!(
            "INSERT INTO {schema}.{TRIO_DETAILS_TABLE} (id, conn1, conn2, conn3) VALUES ($1,$2,$3,$4) ON CONFLICT (id) DO NOTHING"
        );
        let res = sqlx::query(&sql)
            .bind(i16::from(td.id.0))
            .bind(i16::from(td.conns[0].id.0))
            .bind(i16::from(td.conns[1].id.0))
            .bind(i16::from(td.conns[2].id.0))
            .execute(pool)
            .await
            .map_err(|err| QsmError::wrap(err, "insert trio"))?;
        n += res.rows_affected();
    }
    Ok(n)
}

pub async fn save_growth_contexts(
    pool: &PgPool,
    schema: &str,
    rows: &[GrowthContext],
) -> Result<u64, QsmError> {
    let mut n = 0u64;
    for ctx in rows {
        let sql = format!(
            "INSERT INTO {schema}.{GROWTH_CONTEXTS_TABLE} (id, ctx_type, ctx_index) VALUES ($1,$2,$3) ON CONFLICT (id) DO NOTHING"
        );
        let res = sqlx::query(&sql)
            .bind(ctx.id as i16)
            .bind(i16::from(ctx.growth_type.0))
            .bind(ctx.growth_index as i16)
            .execute(pool)
            .await
            .map_err(|err| QsmError::wrap(err, "insert growth context"))?;
        n += res.rows_affected();
    }
    Ok(n)
}

/// Write connections, trios, and growth contexts. Cubes and path builders stay
/// derived in memory; their Go tables are documented in the README.
pub async fn save_point_pack(
    pool: &PgPool,
    env: &DbEnvironment,
    pack: &PointPackData,
) -> Result<(), QsmError> {
    ensure_schema(pool, env).await?;
    save_connections(pool, &env.schema_name, &pack.connections).await?;
    save_trios(pool, &env.schema_name, &pack.trios).await?;
    save_growth_contexts(pool, &env.schema_name, &pack.growth_contexts).await?;
    Ok(())
}

pub async fn count_rows(pool: &PgPool, schema: &str, table: &str) -> Result<i64, QsmError> {
    let sql = format!("SELECT COUNT(*) FROM {schema}.{table}");
    let n: (i64,) = sqlx::query_as(&sql)
        .fetch_one(pool)
        .await
        .map_err(|err| QsmError::wrap(err, format!("count {schema}.{table}")))?;
    Ok(n.0)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn db_conf_round_trip() {
        let details = DbConnDetails {
            host: "test-host".into(),
            port: 1234,
            user: "test-user".into(),
            password: "test-password".into(),
            db_name: "test-db".into(),
        };
        let env = DbEnvironment::new(QsmEnvId::PointTestEnv, details.clone());
        assert_eq!(env.details, details);
        assert_eq!(env.schema_name, "qsm5");
        let url = env.details.connection_string();
        assert!(url.contains("host=test-host"));
        assert!(url.contains("port=1234"));
        assert!(url.contains("user=test-user"));
        assert!(url.contains("password=test-password"));
        assert!(url.contains("dbname=test-db"));
        assert!(url.contains("sslmode=disable"));
    }

    #[test]
    fn schema_ddl_names_the_go_tables() {
        let sql = schema_ddl("qsm5");
        assert!(sql.contains("CREATE SCHEMA IF NOT EXISTS qsm5"));
        assert!(sql.contains("connection_details"));
        assert!(sql.contains("trio_details"));
        assert!(sql.contains("growth_contexts"));
        assert!(sql.contains("PRIMARY KEY"));
        assert!(sql.contains("UNIQUE (ctx_type, ctx_index)"));
        assert_eq!(EXPECTED_CONNECTIONS, 50);
        assert_eq!(EXPECTED_TRIOS, 200);
        assert_eq!(EXPECTED_GROWTH_CONTEXTS, 52);
        assert_eq!(EXPECTED_CUBES, 5192);
        assert!(drop_schema_sql("qsm5").contains("CASCADE"));
        assert_eq!(TRIO_CUBES_TABLE, "trio_cubes");
        assert_eq!(PATH_BUILDERS_TABLE, "path_builders");
    }

    #[test]
    fn save_point_pack_when_database_url_is_set() {
        let url = match std::env::var("QSM_DATABASE_URL") {
            Ok(url) if !url.is_empty() => url,
            _ => return,
        };
        block_on(async {
            let pool = PgPoolOptions::new()
                .connect(&url)
                .await
                .expect("connect QSM_DATABASE_URL");
            let env = DbEnvironment {
                id: QsmEnvId::PointTestEnv,
                details: DbConnDetails {
                    host: "from-url".into(),
                    port: 5432,
                    user: "from-url".into(),
                    password: String::new(),
                    db_name: "from-url".into(),
                },
                schema_name: "qsm_rust_test".into(),
            };
            let pack = m3point::PointPackData::shared();
            save_point_pack(&pool, &env, pack).await.expect("save");
            assert_eq!(
                count_rows(&pool, &env.schema_name, CONNECTION_DETAILS_TABLE)
                    .await
                    .unwrap(),
                EXPECTED_CONNECTIONS as i64
            );
            assert_eq!(
                count_rows(&pool, &env.schema_name, TRIO_DETAILS_TABLE)
                    .await
                    .unwrap(),
                EXPECTED_TRIOS as i64
            );
            assert_eq!(
                count_rows(&pool, &env.schema_name, GROWTH_CONTEXTS_TABLE)
                    .await
                    .unwrap(),
                EXPECTED_GROWTH_CONTEXTS as i64
            );
            sqlx::query(&drop_schema_sql(&env.schema_name))
                .execute(&pool)
                .await
                .unwrap();
        });
    }
}
