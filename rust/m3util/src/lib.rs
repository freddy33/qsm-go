//! Utilities shared by the QSM crates.
//!
//! Integer helpers match the Go `m3util` package: positive moduli are bit masks,
//! and [`count_the_ones`] counts bits in a `u8`.

use std::fmt;
use std::sync::atomic::{AtomicU8, Ordering};

/// Environment slot used by the Go process-wide environment map.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
#[repr(i16)]
pub enum QsmEnvId {
    NoEnv = 0,
    MainEnv = 1,
    RunEnv = 2,
    PerfTestEnv = 3,
    ShellEnv = 4,
    PointTestEnv = 5,
    PathTestEnv = 6,
    SpaceTestEnv = 7,
    GlTestEnv = 8,
    TestServerEnv = 9,
    PointLoadTempEnv = 10,
    PointTempEnv = 11,
    PathTempEnv = 12,
    SpaceTempEnv = 13,
    PointLoadEnv = 14,
    TestClientEnv = 15,
    SpaceClientTempEnv = 16,
}

pub const MAX_NUMBER_OF_ENVIRONMENTS: i16 = 17;
pub const QSM_ENV_NUMBER_KEY: &str = "QSM_ENV_NUMBER";

pub const POINT_IDX: usize = 0;
pub const PATH_IDX: usize = 1;
pub const SPACE_IDX: usize = 2;
pub const GL_IDX: usize = 3;
pub const MAX_DATA_ENTRY: usize = 4;

impl QsmEnvId {
    pub fn from_i16(id: i16) -> Self {
        match id {
            0 => Self::NoEnv,
            1 => Self::MainEnv,
            2 => Self::RunEnv,
            3 => Self::PerfTestEnv,
            4 => Self::ShellEnv,
            5 => Self::PointTestEnv,
            6 => Self::PathTestEnv,
            7 => Self::SpaceTestEnv,
            8 => Self::GlTestEnv,
            9 => Self::TestServerEnv,
            10 => Self::PointLoadTempEnv,
            11 => Self::PointTempEnv,
            12 => Self::PathTempEnv,
            13 => Self::SpaceTempEnv,
            14 => Self::PointLoadEnv,
            15 => Self::TestClientEnv,
            16 => Self::SpaceClientTempEnv,
            _ => panic!("environment id {id} is outside 0..{MAX_NUMBER_OF_ENVIRONMENTS}"),
        }
    }
}

impl fmt::Display for QsmEnvId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", *self as i16)
    }
}

/// `i & 1`, the Go `PosMod2`.
#[inline]
pub fn pos_mod2(i: u64) -> u64 {
    i & 0x1
}

/// `i & 3`, the Go `PosMod4`.
#[inline]
pub fn pos_mod4(i: u64) -> u64 {
    i & 0x3
}

/// `i & 7`, the Go `PosMod8`.
#[inline]
pub fn pos_mod8(i: u64) -> u64 {
    i & 0x7
}

/// Population count of an 8-bit mask, matching `CountTheOnes`.
#[inline]
pub fn count_the_ones(m: u8) -> u8 {
    ((m >> 7) & 1)
        + ((m >> 6) & 1)
        + ((m >> 5) & 1)
        + ((m >> 4) & 1)
        + ((m >> 3) & 1)
        + ((m >> 2) & 1)
        + ((m >> 1) & 1)
        + (m & 1)
}

/// Error carrying a message and an optional cause, as `QsmError` does.
#[derive(Debug)]
pub struct QsmError {
    pub msg: String,
    pub cause: Option<Box<dyn std::error::Error + Send + Sync>>,
}

impl QsmError {
    pub fn new(msg: impl Into<String>) -> Self {
        Self {
            msg: msg.into(),
            cause: None,
        }
    }

    pub fn wrap(
        err: impl std::error::Error + Send + Sync + 'static,
        msg: impl Into<String>,
    ) -> Self {
        Self {
            msg: msg.into(),
            cause: Some(Box::new(err)),
        }
    }
}

impl fmt::Display for QsmError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.msg)
    }
}

impl std::error::Error for QsmError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        self.cause
            .as_ref()
            .map(|c| c.as_ref() as &(dyn std::error::Error + 'static))
    }
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord)]
#[repr(u8)]
pub enum LogLevel {
    Trace = 0,
    Debug = 1,
    Info = 2,
    Warn = 3,
    Error = 4,
    Fatal = 5,
}

/// Small logger. Levels are atomic so tests can quiet a shared logger.
#[derive(Debug)]
pub struct Logger {
    pub prefix: &'static str,
    level: AtomicU8,
}

impl Logger {
    pub const fn new(prefix: &'static str, level: LogLevel) -> Self {
        Self {
            prefix,
            level: AtomicU8::new(level as u8),
        }
    }

    pub fn level(&self) -> LogLevel {
        match self.level.load(Ordering::Relaxed) {
            0 => LogLevel::Trace,
            1 => LogLevel::Debug,
            2 => LogLevel::Info,
            3 => LogLevel::Warn,
            4 => LogLevel::Error,
            _ => LogLevel::Fatal,
        }
    }

    pub fn set_level(&self, level: LogLevel) {
        self.level.store(level as u8, Ordering::Relaxed);
    }

    pub fn enabled(&self, level: LogLevel) -> bool {
        self.level() <= level
    }

    pub fn log(&self, level: LogLevel, args: fmt::Arguments<'_>) {
        if self.enabled(level) {
            eprintln!("{} {:?}", self.prefix, args);
        }
    }
}

#[macro_export]
macro_rules! qsm_info {
    ($log:expr, $($arg:tt)*) => {
        $log.log($crate::LogLevel::Info, format_args!($($arg)*))
    };
}

static TEST_MODE: std::sync::atomic::AtomicBool = std::sync::atomic::AtomicBool::new(false);

pub fn set_to_test_mode() {
    TEST_MODE.store(true, Ordering::Relaxed);
}

pub fn test_mode() -> bool {
    TEST_MODE.load(Ordering::Relaxed)
}

/// Parse an environment id the way `ReadEnvId` does (base 10, 16-bit).
pub fn read_env_id(source_info: &str, env_id_str: &str) -> QsmEnvId {
    match env_id_str.parse::<i16>() {
        Ok(id) => QsmEnvId::from_i16(id),
        Err(_) => panic!("The variable from {source_info} is not a DB number but {env_id_str}"),
    }
}

pub fn get_default_env_id() -> QsmEnvId {
    match std::env::var(QSM_ENV_NUMBER_KEY) {
        Ok(value) if !value.is_empty() => {
            read_env_id(&format!("env var \"{QSM_ENV_NUMBER_KEY}\""), &value)
        }
        _ => QsmEnvId::MainEnv,
    }
}

pub fn get_compulsory_env(key: &str) -> Result<String, QsmError> {
    match std::env::var(key) {
        Ok(value) if !value.is_empty() => Ok(value),
        _ => Err(QsmError::new(format!("missing {key}"))),
    }
}

pub fn get_compulsory_env_int(key: &str) -> Result<i32, QsmError> {
    let value = get_compulsory_env(key)?;
    value
        .parse::<i32>()
        .map_err(|_| QsmError::new(format!("error parsing {key} to int")))
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::error::Error;

    #[test]
    fn pos_mod_masks_match_go() {
        assert_eq!(pos_mod2(5), 1);
        assert_eq!(pos_mod2(4), 0);
        assert_eq!(pos_mod2(3), 1);
        assert_eq!(pos_mod2(2), 0);
        assert_eq!(pos_mod2(1), 1);
        assert_eq!(pos_mod2(0), 0);

        assert_eq!(pos_mod4(5), 1);
        assert_eq!(pos_mod4(4), 0);
        assert_eq!(pos_mod4(3), 3);
        assert_eq!(pos_mod4(2), 2);
        assert_eq!(pos_mod4(1), 1);
        assert_eq!(pos_mod4(0), 0);

        assert_eq!(pos_mod8(9), 1);
        assert_eq!(pos_mod8(8), 0);
        assert_eq!(pos_mod8(7), 7);
        assert_eq!(pos_mod8(6), 6);
        assert_eq!(pos_mod8(5), 5);
        assert_eq!(pos_mod8(4), 4);
        assert_eq!(pos_mod8(3), 3);
        assert_eq!(pos_mod8(2), 2);
        assert_eq!(pos_mod8(1), 1);
        assert_eq!(pos_mod8(0), 0);
    }

    #[test]
    fn count_ones_matches_popcount() {
        assert_eq!(count_the_ones(0), 0);
        assert_eq!(count_the_ones(0b1010_0101), 4);
        assert_eq!(count_the_ones(0xff), 8);
    }

    #[test]
    fn env_id_round_trip() {
        assert_eq!(QsmEnvId::PointTestEnv.to_string(), "5");
        assert_eq!(read_env_id("test", "14"), QsmEnvId::PointLoadEnv);
    }

    #[test]
    fn qsm_error_displays_message() {
        let err = QsmError::new("missing table");
        assert_eq!(err.to_string(), "missing table");
        let wrapped = QsmError::wrap(err, "while loading");
        assert_eq!(wrapped.to_string(), "while loading");
        assert!(wrapped.source().is_some());
    }
}
