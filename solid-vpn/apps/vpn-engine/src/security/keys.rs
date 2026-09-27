pub fn is_valid_public_key(key: &str) -> bool {
    key.len() == 44 && key.chars().all(|c| c.is_ascii_alphanumeric() || c == '+' || c == '/' || c == '=')
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn valid_key_format_accepted() {
        let key = "mNb8O2FNkBkJo5tXhA3UGN4sbeE6DKBP3gRKe3DXnWk=";
        assert!(is_valid_public_key(key));
    }

    #[test]
    fn short_key_rejected() {
        assert!(!is_valid_public_key("tooshort"));
    }

    #[test]
    fn key_with_invalid_chars_rejected() {
        let bad = "mNb8O2FNkBkJo5tXhA3UGN4sbeE6DKBP3gRKe3DX!!!=";
        assert!(!is_valid_public_key(bad));
    }
}
