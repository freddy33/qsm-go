use std::fmt;

/// Failure while opening a simulation folder or reading and writing its tables.
#[derive(Debug)]
pub struct Error {
    message: String,
}

impl Error {
    pub fn new(message: impl Into<String>) -> Self {
        Self {
            message: message.into(),
        }
    }
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(&self.message)
    }
}

impl std::error::Error for Error {}

impl From<iceberg::Error> for Error {
    fn from(err: iceberg::Error) -> Self {
        Self::new(err.to_string())
    }
}

impl From<arrow::error::ArrowError> for Error {
    fn from(err: arrow::error::ArrowError) -> Self {
        Self::new(err.to_string())
    }
}

impl From<std::io::Error> for Error {
    fn from(err: std::io::Error) -> Self {
        Self::new(err.to_string())
    }
}

impl From<serde_json::Error> for Error {
    fn from(err: serde_json::Error) -> Self {
        Self::new(err.to_string())
    }
}

impl From<m3util::QsmError> for Error {
    fn from(err: m3util::QsmError) -> Self {
        Self::new(err.to_string())
    }
}

pub type Result<T> = std::result::Result<T, Error>;
