//! 大整数 JSON 修复(`docs/spec/09-cdp-contract.md` "大整数 JSON 处理")。
//!
//! 豆包的 ID 字段(`conversation_id` / `local_message_id` / `message_id` / `bot_reply_message_id`)
//! 是 >16 位整数,超过 JS Number 与 `serde_json::Number::as_i64` 的安全范围。
//! Rust 侧从 CDP 返回的文本里读 JSON 前,先把这些字段的裸数字包成字符串;
//! 这些 ID 在 Rust 结构里始终是 `String`。

use std::sync::LazyLock;

use regex::Regex;

/// 09-cdp-contract 的正则用了 lookahead `(?=\s*[,}])`,Rust `regex` 不支持;
/// 这里把后缀分隔符作为第 3 组捕获并原样写回,匹配语义相同(另外多接受 `]` 作为分隔符)。
static BIG_INT_FIELDS: LazyLock<Regex> = LazyLock::new(|| {
    Regex::new(
        r#""(conversation_id|local_message_id|message_id|bot_reply_message_id)"\s*:\s*(-?\d{16,})(\s*[,}\]])"#,
    )
    .expect("BIG_INT_FIELDS 正则非法")
});

/// 把指定 ID 字段的 ≥16 位整数值(含负数)包成 JSON 字符串,其它内容原样返回。
pub fn fix_big_integers(json_text: &str) -> String {
    BIG_INT_FIELDS
        .replace_all(json_text, r#""${1}":"${2}"${3}"#)
        .into_owned()
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::{json, Value};

    #[test]
    fn wraps_17_digit_message_id() {
        let fixed = fix_big_integers(r#"{"message_id":12345678901234567,"x":1}"#);
        assert_eq!(fixed, r#"{"message_id":"12345678901234567","x":1}"#);
        let v: Value = serde_json::from_str(&fixed).unwrap();
        assert_eq!(v["message_id"], "12345678901234567");
    }

    #[test]
    fn wraps_negative_and_all_four_fields() {
        let text = r#"{"conversation_id": -9876543210987654321, "local_message_id" :1234567890123456 ,"message_id":12345678901234567890,"bot_reply_message_id":98765432109876543}"#;
        let v: Value = serde_json::from_str(&fix_big_integers(text)).unwrap();
        assert_eq!(v["conversation_id"], "-9876543210987654321");
        assert_eq!(v["local_message_id"], "1234567890123456");
        assert_eq!(v["message_id"], "12345678901234567890");
        assert_eq!(v["bot_reply_message_id"], "98765432109876543");
    }

    #[test]
    fn wraps_nested_and_array_items() {
        let text = r#"{"data":{"messages":[{"message_id":11111111111111111,"body":"a"},{"message_id":22222222222222222}],"conversation_id":33333333333333333}}"#;
        let v: Value = serde_json::from_str(&fix_big_integers(text)).unwrap();
        assert_eq!(v["data"]["messages"][0]["message_id"], "11111111111111111");
        assert_eq!(v["data"]["messages"][1]["message_id"], "22222222222222222");
        assert_eq!(v["data"]["conversation_id"], "33333333333333333");
    }

    #[test]
    fn leaves_already_string_values_alone() {
        let text = r#"{"message_id":"12345678901234567","conversation_id":"-1"}"#;
        assert_eq!(fix_big_integers(text), text);
    }

    #[test]
    fn leaves_small_numbers_and_other_fields_alone() {
        // 15 位在 2^53 内,不动;非白名单字段不动;浮点不动。
        let text = r#"{"message_id":123456789012345,"other_id":12345678901234567,"score":1234567890123456.5,"message_id_x":12345678901234567}"#;
        assert_eq!(fix_big_integers(text), text);
        let v: Value = serde_json::from_str(&fix_big_integers(text)).unwrap();
        assert_eq!(v["message_id"], json!(123456789012345_i64));
        assert_eq!(v["other_id"], json!(12345678901234567_i64));
    }

    #[test]
    fn handles_whitespace_and_last_key_before_brace() {
        let text = "{\n  \"message_id\" : 12345678901234567 \n}";
        let v: Value = serde_json::from_str(&fix_big_integers(text)).unwrap();
        assert_eq!(v["message_id"], "12345678901234567");
    }

    #[test]
    fn does_not_touch_digits_inside_string_values() {
        // 值是字符串、且字符串里恰好含 `"message_id":1234...` 这种假阳性文本时,
        // 正则不会越过引号匹配(字符串内部的引号被转义)。
        let text = r#"{"text":"see \"message_id\":12345678901234567 inside"}"#;
        assert_eq!(fix_big_integers(text), text);
    }

    #[test]
    fn empty_and_non_json_input_pass_through() {
        assert_eq!(fix_big_integers(""), "");
        assert_eq!(fix_big_integers("not json"), "not json");
    }
}
