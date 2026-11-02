//! 各端点(`docs/spec/04-http-api.md`)。每个子模块提供 `router()`,由 `http::router` 组装;
//! `health` 不鉴权,其余都在 `auth_middleware` 之后。

pub mod accounts;
pub mod doubao;
pub mod generate;
pub mod health;
pub mod openai;
pub mod tasks;
pub mod uploads;
