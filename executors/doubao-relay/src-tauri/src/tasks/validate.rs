//! 请求校验(F3.1 / F3.2 / F3.5 + `11-error-codes.md` `request_*`)。
//!
//! 只做**无状态**校验(字段齐全、枚举、范围、数量、callback_url);
//! "账号存在且已登录"与"reference_files ID 真实且未过期"需要运行时状态,
//! 由调度/HTTP 层在入队前另行检查(`account_not_found` / `upload_not_found`)。

use serde_json::json;
use url::{Host, Url};

use crate::error::{RelayError, RelayResult};
use crate::tasks::model::{GenerateRequest, TaskType};

/// 图像模型白名单(F3.1)。
pub const IMAGE_MODELS: &[&str] = &[
    "Seedream 5.0 Flash",
    "Seedream 5.0 Pro",
    "Seedream 4.5",
    "Seedream 4.0",
];

/// 视频模型白名单(F3.2)。
pub const VIDEO_MODELS: &[&str] = &[
    "Seedance 2.5",
    "Seedance 2.0",
    "Seedance 2.0 Fast",
    "Seedance 2.0 Mini",
];

/// 画幅比例白名单(F3.1 的统一列表 = 图像 ∪ 视频;spec 原列表,`request_invalid_enum.allowed` 按类型给)。
pub const RATIOS: &[&str] = &[
    "自动", "1:1", "16:9", "9:16", "4:3", "3:4", "3:2", "2:3", "21:9",
];

/// 图像比例白名单(真机 2.32.5 图像菜单:无 21:9,见 docs/analysis/10 Task 2.1)。
pub const IMAGE_RATIOS: &[&str] = &["自动", "1:1", "16:9", "9:16", "4:3", "3:4", "3:2", "2:3"];

/// 视频比例白名单(真机 2.32.5 视频菜单:无 3:2 / 2:3)。
pub const VIDEO_RATIOS: &[&str] = &["自动", "1:1", "16:9", "9:16", "4:3", "3:4", "21:9"];

/// 按任务类型给比例白名单。
pub fn ratios_for(task_type: TaskType) -> &'static [&'static str] {
    match task_type {
        TaskType::Image => IMAGE_RATIOS,
        TaskType::Video => VIDEO_RATIOS,
    }
}

/// prompt 最大长度(字符数,F3.1)。
pub const PROMPT_MAX_CHARS: usize = 10_000;

/// 参考文件数量上限(F3.1 / F3.2)。
pub const MAX_REFERENCES: usize = 10;

/// `Seedance 2.0 Fast` 的参考文件数量上限(F3.2)。
pub const MAX_REFERENCES_FAST: usize = 11;

/// 视频时长(秒)范围:`Seedance 2.5` 4–30,其余 4–15(F3.2)。
pub fn duration_range(model_name: &str) -> (u32, u32) {
    if model_name == "Seedance 2.5" {
        (4, 30)
    } else {
        (4, 15)
    }
}

/// 参考文件数量上限:`Seedance 2.0 Fast` 11,其余 10。
pub fn max_references(r: &GenerateRequest) -> usize {
    if r.task_type == TaskType::Video && r.model_name == "Seedance 2.0 Fast" {
        MAX_REFERENCES_FAST
    } else {
        MAX_REFERENCES
    }
}

/// 校验请求;失败返回 `request_*` 错误(HTTP 400),调用方不得入队。
///
/// 检查顺序:account → prompt(缺/过长)→ model_name → ratio → duration(视频)
/// → reference_files 数量 → callback_url。
pub fn validate_request(r: &GenerateRequest) -> RelayResult<()> {
    if r.account.trim().is_empty() {
        return Err(missing_field("account"));
    }
    if r.prompt.trim().is_empty() {
        return Err(missing_field("prompt"));
    }
    let prompt_len = r.prompt.chars().count();
    if prompt_len > PROMPT_MAX_CHARS {
        return Err(prompt_too_long(prompt_len));
    }

    if r.model_name.trim().is_empty() {
        return Err(missing_field("model_name"));
    }
    let allowed_models = match r.task_type {
        TaskType::Image => IMAGE_MODELS,
        TaskType::Video => VIDEO_MODELS,
    };
    if !allowed_models.contains(&r.model_name.as_str()) {
        return Err(invalid_enum("model_name", &r.model_name, allowed_models));
    }

    if r.ratio.trim().is_empty() {
        return Err(missing_field("ratio"));
    }
    let allowed_ratios = ratios_for(r.task_type);
    if !allowed_ratios.contains(&r.ratio.as_str()) {
        return Err(invalid_enum("ratio", &r.ratio, allowed_ratios));
    }

    if r.task_type == TaskType::Video {
        let duration = r.duration.ok_or_else(|| missing_field("duration"))?;
        let (min, max) = duration_range(&r.model_name);
        if duration < min || duration > max {
            return Err(invalid_range("duration", duration, min, max));
        }
    }

    for (i, id) in r.reference_files.iter().enumerate() {
        if id.trim().is_empty() {
            return Err(missing_field(&format!("reference_files[{i}]")));
        }
    }
    let max = max_references(r);
    if r.reference_files.len() > max {
        return Err(too_many_references(r.reference_files.len(), max));
    }

    if let Some(url) = &r.callback_url {
        if !is_valid_callback_url(url) {
            return Err(invalid_callback_url(url));
        }
    }
    Ok(())
}

/// callback_url 规则(12-security "URL 校验"):`http(s)://`、有 host、
/// 且 host 不是 loopback / unspecified(`localhost`、`*.localhost`、`127/8`、`::1`、`0.0.0.0`、`::`)。
pub fn is_valid_callback_url(url: &str) -> bool {
    let Ok(parsed) = Url::parse(url) else {
        return false;
    };
    if !matches!(parsed.scheme(), "http" | "https") {
        return false;
    }
    match parsed.host() {
        Some(Host::Domain(d)) => {
            let d = d.to_ascii_lowercase();
            let d = d.trim_end_matches('.');
            !(d.is_empty() || d == "localhost" || d.ends_with(".localhost"))
        }
        Some(Host::Ipv4(ip)) => !(ip.is_loopback() || ip.is_unspecified()),
        Some(Host::Ipv6(ip)) => !(ip.is_loopback() || ip.is_unspecified()),
        None => false,
    }
}

// ---- request_* 错误构造(11-error-codes 请求类) ----

/// `request_missing_field` `{ field }`。
pub fn missing_field(field: &str) -> RelayError {
    RelayError::new(
        "request_missing_field",
        format!("缺少必填字段 {field},请补全后重试"),
    )
    .with_details(json!({ "field": field }))
}

/// `request_invalid_enum` `{ field, value, allowed }`。
pub fn invalid_enum(field: &str, value: &str, allowed: &[&str]) -> RelayError {
    RelayError::new(
        "request_invalid_enum",
        format!("字段 {field} 的值 {value} 不在允许范围内,请从允许的取值中选择"),
    )
    .with_details(json!({ "field": field, "value": value, "allowed": allowed }))
}

/// `request_invalid_range` `{ field, value, min, max }`。
pub fn invalid_range(field: &str, value: u32, min: u32, max: u32) -> RelayError {
    RelayError::new(
        "request_invalid_range",
        format!("字段 {field} 的值 {value} 超出范围,请使用 {min} 到 {max} 之间的值"),
    )
    .with_details(json!({ "field": field, "value": value, "min": min, "max": max }))
}

/// `request_prompt_too_long` `{ length }`。
pub fn prompt_too_long(length: usize) -> RelayError {
    RelayError::new(
        "request_prompt_too_long",
        format!("提示词长度 {length} 超过 {PROMPT_MAX_CHARS} 字上限,请缩短后重试"),
    )
    .with_details(json!({ "length": length }))
}

/// `request_too_many_references` `{ count, max }`。
pub fn too_many_references(count: usize, max: usize) -> RelayError {
    RelayError::new(
        "request_too_many_references",
        format!("参考文件数量 {count} 超过上限 {max},请减少参考文件后重试"),
    )
    .with_details(json!({ "count": count, "max": max }))
}

/// `request_invalid_callback_url` `{ url }`。
pub fn invalid_callback_url(url: &str) -> RelayError {
    RelayError::new(
        "request_invalid_callback_url",
        "callback_url 不合法,请使用非本机回环地址的 http 或 https 地址",
    )
    .with_details(json!({ "url": url }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::tasks::model::ConfirmationMode;

    fn image() -> GenerateRequest {
        GenerateRequest {
            account: "张三".into(),
            task_type: TaskType::Image,
            model_name: "Seedream 5.0 Pro".into(),
            prompt: "一只蓝色小鸟".into(),
            ratio: "16:9".into(),
            duration: None,
            reference_files: vec![],
            confirmation_mode: ConfirmationMode::Auto,
            auto_use_membership: true,
            callback_url: None,
        }
    }

    fn video(model: &str, duration: Option<u32>) -> GenerateRequest {
        GenerateRequest {
            task_type: TaskType::Video,
            model_name: model.into(),
            duration,
            ..image()
        }
    }

    fn refs(n: usize) -> Vec<String> {
        (0..n).map(|i| format!("01HQZ9UPLOAD{i:014}")).collect()
    }

    fn err(r: &GenerateRequest) -> RelayError {
        validate_request(r).expect_err("expected validation error")
    }

    #[test]
    fn valid_image_and_video_pass() {
        validate_request(&image()).unwrap();
        for m in IMAGE_MODELS {
            validate_request(&GenerateRequest {
                model_name: (*m).into(),
                ..image()
            })
            .unwrap();
        }
        for ratio in IMAGE_RATIOS {
            validate_request(&GenerateRequest {
                ratio: (*ratio).into(),
                ..image()
            })
            .unwrap();
        }
        for ratio in VIDEO_RATIOS {
            validate_request(&GenerateRequest {
                ratio: (*ratio).into(),
                ..video("Seedance 2.5", Some(10))
            })
            .unwrap();
        }
        for m in VIDEO_MODELS {
            validate_request(&video(m, Some(4))).unwrap();
            validate_request(&video(m, Some(15))).unwrap();
        }
    }

    #[test]
    fn missing_account_and_prompt() {
        let e = err(&GenerateRequest {
            account: "".into(),
            ..image()
        });
        assert_eq!(e.code, "request_missing_field");
        assert_eq!(e.http_status(), 400);
        assert_eq!(e.details.unwrap()["field"], "account");

        let e = err(&GenerateRequest {
            prompt: "".into(),
            ..image()
        });
        assert_eq!(e.code, "request_missing_field");
        assert_eq!(e.details.unwrap()["field"], "prompt");

        let e = err(&GenerateRequest {
            prompt: "   \n".into(),
            ..image()
        });
        assert_eq!(e.code, "request_missing_field");
        assert_eq!(e.details.unwrap()["field"], "prompt");
    }

    #[test]
    fn prompt_too_long_counts_chars_not_bytes() {
        validate_request(&GenerateRequest {
            prompt: "鸟".repeat(10_000),
            ..image()
        })
        .unwrap();
        let e = err(&GenerateRequest {
            prompt: "鸟".repeat(10_001),
            ..image()
        });
        assert_eq!(e.code, "request_prompt_too_long");
        assert_eq!(e.details.unwrap()["length"], 10_001);
    }

    #[test]
    fn model_not_in_whitelist() {
        let e = err(&GenerateRequest {
            model_name: "Seedream 3.0".into(),
            ..image()
        });
        assert_eq!(e.code, "request_invalid_enum");
        let d = e.details.unwrap();
        assert_eq!(d["field"], "model_name");
        assert_eq!(d["value"], "Seedream 3.0");
        assert_eq!(d["allowed"], json!(IMAGE_MODELS));

        // 视频模型用于图像任务也算非法,allowed 按类型给。
        let e = err(&GenerateRequest {
            model_name: "Seedance 2.5".into(),
            ..image()
        });
        assert_eq!(e.code, "request_invalid_enum");
        assert_eq!(e.details.unwrap()["allowed"], json!(IMAGE_MODELS));

        let e = err(&video("Seedream 5.0 Pro", Some(5)));
        assert_eq!(e.code, "request_invalid_enum");
        let d = e.details.unwrap();
        assert_eq!(d["field"], "model_name");
        assert_eq!(d["allowed"], json!(VIDEO_MODELS));

        let e = err(&GenerateRequest {
            model_name: "".into(),
            ..image()
        });
        assert_eq!(e.code, "request_missing_field");
        assert_eq!(e.details.unwrap()["field"], "model_name");
    }

    #[test]
    fn ratio_not_in_whitelist() {
        let e = err(&GenerateRequest {
            ratio: "5:4".into(),
            ..image()
        });
        assert_eq!(e.code, "request_invalid_enum");
        let d = e.details.unwrap();
        assert_eq!(d["field"], "ratio");
        assert_eq!(d["value"], "5:4");
        assert_eq!(d["allowed"], json!(IMAGE_RATIOS));

        let e = err(&GenerateRequest {
            ratio: "".into(),
            ..image()
        });
        assert_eq!(e.code, "request_missing_field");
        assert_eq!(e.details.unwrap()["field"], "ratio");
    }

    /// 真机(Task 2.1):图像菜单没有 21:9,视频菜单没有 3:2 / 2:3,统一列表会在 prepare 阶段才失败。
    #[test]
    fn ratio_whitelist_is_split_by_task_type() {
        let e = err(&GenerateRequest {
            ratio: "21:9".into(),
            ..image()
        });
        assert_eq!(e.code, "request_invalid_enum");
        assert_eq!(e.details.unwrap()["allowed"], json!(IMAGE_RATIOS));
        for bad in ["3:2", "2:3"] {
            let e = err(&GenerateRequest {
                ratio: bad.into(),
                ..video("Seedance 2.5", Some(10))
            });
            assert_eq!(e.code, "request_invalid_enum", "{bad}");
            assert_eq!(e.details.unwrap()["allowed"], json!(VIDEO_RATIOS));
        }
        validate_request(&GenerateRequest {
            ratio: "21:9".into(),
            ..video("Seedance 2.5", Some(10))
        })
        .unwrap();
        // 统一列表仍是两者之并(spec 01 F3.1 原列表)
        for r in IMAGE_RATIOS.iter().chain(VIDEO_RATIOS) {
            assert!(RATIOS.contains(r), "{r}");
        }
        assert_eq!(ratios_for(TaskType::Image), IMAGE_RATIOS);
        assert_eq!(ratios_for(TaskType::Video), VIDEO_RATIOS);
    }

    #[test]
    fn video_requires_duration() {
        let e = err(&video("Seedance 2.0", None));
        assert_eq!(e.code, "request_missing_field");
        assert_eq!(e.details.unwrap()["field"], "duration");
        // 图像任务带 duration 不报错(忽略)。
        validate_request(&GenerateRequest {
            duration: Some(99),
            ..image()
        })
        .unwrap();
    }

    #[test]
    fn duration_ranges_per_model() {
        validate_request(&video("Seedance 2.5", Some(30))).unwrap();
        let e = err(&video("Seedance 2.5", Some(31)));
        assert_eq!(e.code, "request_invalid_range");
        let d = e.details.unwrap();
        assert_eq!(d["field"], "duration");
        assert_eq!(d["value"], 31);
        assert_eq!(d["min"], 4);
        assert_eq!(d["max"], 30);

        let e = err(&video("Seedance 2.0", Some(16)));
        assert_eq!(e.code, "request_invalid_range");
        let d = e.details.unwrap();
        assert_eq!(d["value"], 16);
        assert_eq!(d["min"], 4);
        assert_eq!(d["max"], 15);

        for m in ["Seedance 2.0 Fast", "Seedance 2.0 Mini"] {
            assert_eq!(err(&video(m, Some(16))).code, "request_invalid_range");
        }
        let e = err(&video("Seedance 2.5", Some(3)));
        assert_eq!(e.code, "request_invalid_range");
        assert_eq!(e.details.unwrap()["min"], 4);
        let e = err(&video("Seedance 2.0", Some(0)));
        assert_eq!(e.code, "request_invalid_range");
    }

    #[test]
    fn reference_file_limits() {
        validate_request(&GenerateRequest {
            reference_files: refs(10),
            ..image()
        })
        .unwrap();
        let e = err(&GenerateRequest {
            reference_files: refs(11),
            ..image()
        });
        assert_eq!(e.code, "request_too_many_references");
        let d = e.details.unwrap();
        assert_eq!(d["count"], 11);
        assert_eq!(d["max"], 10);

        // 普通视频 ≤10,Fast ≤11。
        let e = err(&GenerateRequest {
            reference_files: refs(11),
            ..video("Seedance 2.0", Some(5))
        });
        assert_eq!(e.code, "request_too_many_references");
        assert_eq!(e.details.unwrap()["max"], 10);
        validate_request(&GenerateRequest {
            reference_files: refs(11),
            ..video("Seedance 2.0 Fast", Some(5))
        })
        .unwrap();
        let e = err(&GenerateRequest {
            reference_files: refs(12),
            ..video("Seedance 2.0 Fast", Some(5))
        });
        assert_eq!(e.code, "request_too_many_references");
        let d = e.details.unwrap();
        assert_eq!(d["count"], 12);
        assert_eq!(d["max"], 11);

        // 空字符串 ID 算缺字段。
        let e = err(&GenerateRequest {
            reference_files: vec!["".into()],
            ..image()
        });
        assert_eq!(e.code, "request_missing_field");
        assert_eq!(e.details.unwrap()["field"], "reference_files[0]");
    }

    #[test]
    fn callback_url_rules() {
        for ok in [
            "https://my-service/doubao-webhook",
            "http://192.168.1.10:8080/hook",
            "https://example.com/x?y=1",
            "http://10.0.0.5/x",
        ] {
            validate_request(&GenerateRequest {
                callback_url: Some(ok.into()),
                ..image()
            })
            .unwrap_or_else(|e| panic!("{ok} should pass: {e:?}"));
        }
        for bad in [
            "http://127.0.0.1/x",
            "http://127.1.2.3:9000/x",
            "https://localhost/x",
            "http://LOCALHOST:8787/api",
            "http://foo.localhost/x",
            "http://[::1]:8787/x",
            "http://0.0.0.0/x",
            "http://[::]/x",
            "ftp://example.com/x",
            "example.com/x",
            "not a url",
            "",
            "file:///etc/passwd",
            "https://",
        ] {
            let e = err(&GenerateRequest {
                callback_url: Some(bad.into()),
                ..image()
            });
            assert_eq!(e.code, "request_invalid_callback_url", "{bad}");
            assert_eq!(e.http_status(), 400);
            assert_eq!(e.details.unwrap()["url"], bad);
        }
    }

    #[test]
    fn error_messages_have_no_trailing_punctuation() {
        let cases = [
            GenerateRequest {
                account: "".into(),
                ..image()
            },
            GenerateRequest {
                prompt: "鸟".repeat(10_001),
                ..image()
            },
            GenerateRequest {
                model_name: "x".into(),
                ..image()
            },
            video("Seedance 2.0", Some(99)),
            GenerateRequest {
                reference_files: refs(11),
                ..image()
            },
            GenerateRequest {
                callback_url: Some("ftp://x".into()),
                ..image()
            },
        ];
        for c in cases {
            let e = err(&c);
            let last = e.message.chars().last().unwrap();
            assert!(
                !matches!(last, '。' | '.' | '!' | '\u{FF01}' | ',' | '\u{FF0C}'),
                "{}",
                e.message
            );
        }
    }
}
